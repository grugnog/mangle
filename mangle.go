/*
Sanitization / data masking library for Go (golang).


Purpose

This library provides functionality to sanitize text and HTML data. This can be integrated into tools that export or manipulate databases/files so that confidential data is not exposed to staging, development or testing systems.

Getting started

Install mangle:

		go get github.com/grugnog/mangle

Try the command line tool:

		# Generate a list of words using your system dictionary and download "War and Peace" for testing.
		aspell -d en dump master | aspell -l en expand > corpus.txt
		wget http://www.gutenberg.org/cache/epub/2600/pg2600.txt

		# Run the sample text through manglefile via stdin and stdout.
		cat pg2600.txt | manglefile -corpus=corpus.txt -secret=replace-with-a-secure-passphrase | less

For basic usage of mangle as a Go library for strings, see the Example below.  For usage using io.Reader/io.Writer for text and html, see the manglefile source code.

Description

Secure - every non-punctuation, non-tag word is replaced, with no reuse of source words.

Fast - around 350,000 words/sec ("War and Peace" mangled in < 2s).

Deterministic - words are selected from a corpus deterministically - subsequent runs will result in the same word replacements as long as the user secret is not changed. This allows automated tests to run against the sanitized data, as well as for efficient rsync's of mangled versions of large slowly changing data sets.

Accurate - maintains a high level of resemblance to the source text, to allow for realistic testing of search indexes, page layout etc. Replacement words are all natural words from a user defined corpus. Source text word length, punctuation, title case and all caps, HTML tags and attributes are maintained.
*/
package mangle

import (
	"bufio"
	//	"bytes"
	"code.google.com/p/go.net/html"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/jimsmart/bufrr"
	"hash/crc32"
	"io"
	"os"
	"strings"
	"unicode"
)

// Configures a Mangle prior to use.
type Mangle struct {
	// Corpus of words to use as replacements. An array of word lengths, each
	// containing an array of words of that length.
	Corpus [255][]string
	// A sufficiently long secret, used as a salt so rainbow tables cannot be
	// used to reverse the hashes.
	Secret string
}

// ReadCorpus is a helper function that opens and reads a corpus file of words
// and returns an array of word lengths, each containing an array of words of
// that length.
func ReadCorpus(filepath string) ([255][]string, error) {
	var corpus [255][]string
	file, err := os.Open(filepath)
	if err != nil {
		return corpus, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	corpus, err = BuildCorpus(scanner)
	return corpus, err
}

// ReadCorpus is a helper function that reads a bufio.Scanner of words and
// returns an array of word lengths, each containing an array of words of that
// length.
func BuildCorpus(scanner *bufio.Scanner) ([255][]string, error) {
	var corpus [255][]string
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		wordlen := len(scanner.Text())
		corpus[wordlen] = append(corpus[wordlen], scanner.Text())
	}
	if len(corpus[1]) == 0 {
		return corpus, errors.New("Corpus must contain at least one single character word.")
	}
	return corpus, scanner.Err()
}

// MangleString operates on strings, and is preferable if you have many short
// strings to operate on.
func (m Mangle) MangleString(s string) string {
	var output string
	var word []rune
	runes := []rune(s)
	strlen := len(runes)
	for i := 0; i < strlen; i++ {
		rune := runes[i]
		if unicode.IsLetter(rune) || unicode.IsNumber(rune) {
			// In word.
			word = append(word, rune)
		} else {
			// Inter-word.
			if len(word) > 0 {
				// Process previous word.
				output += m.mangleWord(word)
				// Reset word.
				word = word[0:0]
			}
			output += string(rune)
		}
	}
	// Process last word.
	output += m.mangleWord(word)
	return output
}

// MangleHTML operates on HTML using an io interface, preserving all HTML tags
// (including tag attributes), but mangling all content around and between
// tags.
func (m Mangle) MangleHTML(r io.Reader, w io.Writer) error {
	err := m.mangleHTMLParser(r, w)
	if err.Error() == "EOF" {
		err = nil
	}
	return err
}

// MangleIO operates on an io interface, parsing as plain text, and is preferable for long strings.
func (m Mangle) MangleIO(r io.Reader, w io.Writer) error {
	var word []rune
	var rune rune
	var err error
	bufr := bufrr.NewReader(r)
	for {
		rune, _, err = bufr.ReadRune()
		if err != nil {
			return err
		}
		if rune == bufrr.EOF {
			// Process last word.
			fmt.Fprint(w, m.mangleWord(word))
			return nil
		}
		if unicode.IsLetter(rune) || unicode.IsNumber(rune) {
			// In word.
			word = append(word, rune)
		} else {
			// Inter-word.
			if len(word) > 0 {
				// Process previous word.
				fmt.Fprint(w, m.mangleWord(word))
				// Reset word.
				word = word[0:0]
			}
			fmt.Fprint(w, string(rune))
		}
	}
}

// Operates the HTML tokenizer, skipping tags but mangling content.
func (m Mangle) mangleHTMLParser(r io.Reader, w io.Writer) error {
	z := html.NewTokenizer(r)
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return z.Err()
		case html.TextToken:
			token := string(z.Text())
			fmt.Fprint(w, m.MangleString(token))
		default:
			fmt.Fprint(w, string(z.Raw()))
		}
	}
}

// Performs the core mangling function on a word. The approach is to hash the
// word and the secret salt, then map the hash value into the available corpus
// words of the appropriate length (or the longest available length, padding
// with whitespace as needed). The resulting word is then adjusted to
// match the original capitalization.
func (m Mangle) mangleWord(word []rune) string {
	var crc float64
	var pos uint32
	var replacement_runes []rune
	const MaxUint32 = 1<<32 - 1
	replacement := ""

	pad := 0
	word_len := len(word)
	if word_len > 0 {
		// SHA256 the string, together with the secret.
		hash := sha256.New()
		hash.Write([]byte(string(word)))
		hash.Write([]byte(m.Secret))
		// Use crc32 to map the hash to a conveniently sized number.
		crc = float64(crc32.ChecksumIEEE([]byte(hash.Sum(nil))))
		// If we can't find a sufficiently long string, look for a shorter one.
		for len(m.Corpus[word_len]) == 0 && word_len != 0 {
			word_len--
			pad++
		}
		if word_len != 0 {
			// Map the CRC value onto the available corpus words.
			pos = uint32((crc / MaxUint32) * float64(len(m.Corpus[word_len])))
			// Select the word from the corpus and pad it if it was shorter than the original.
			replacement = m.Corpus[word_len][pos] + strings.Repeat(" ", pad)
			// Capitalize as per the original string.
			if word_len > 0 && unicode.IsUpper(word[0]) {
				if word_len > 1 && unicode.IsUpper(word[1]) {
					replacement = strings.ToUpper(replacement)
				} else {
					replacement_runes = []rune(replacement)
					replacement = strings.ToUpper(string(replacement_runes[0])) + string(replacement_runes[1:])
				}
			}
		}
	}
	return replacement
}
