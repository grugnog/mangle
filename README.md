# mangle
--
    import "github.com/grugnog/mangle"

Sanitization / data masking library for Go (golang).


### Purpose

This library provides functionality to sanitize text and HTML data. This can be
integrated into tools that export or manipulate databases/files so that
confidential data is not exposed to staging, development or testing systems.


Getting started

Install mangle:

    go get github.com/grugnog/mangle

Try the command line tool:

    go get github.com/grugnog/mangle/manglefile

    # Generate a list of words using your system dictionary and download "War and Peace" for testing.
    aspell -d en dump master | aspell -l en expand > corpus.txt
    wget http://www.gutenberg.org/cache/epub/2600/pg2600.txt

    # Run the sample text through manglefile via stdin and stdout.
    cat pg2600.txt | manglefile -corpus=corpus.txt -secret=replace-with-a-secure-passphrase | less

For basic usage of mangle as a Go library for strings, see the Example below.
For usage using io.Reader/io.Writer for text and html, see the manglefile source
code.


### Description

Secure - every non-punctuation, non-tag word is replaced, with no reuse of
source words.

Fast - around 350,000 words/sec ("War and Peace" mangled in < 2s).

Deterministic - words are selected from a corpus deterministically - subsequent
runs will result in the same word replacements as long as the user secret is not
changed. This allows automated tests to run against the sanitized data, as well
as for efficient rsync's of mangled versions of large slowly changing data sets.

Accurate - maintains a high level of resemblance to the source text, to allow
for realistic testing of search indexes, page layout etc. Replacement words are
all natural words from a user defined corpus. Source text word length,
punctuation, title case and all caps, HTML tags and attributes are maintained.

## Usage

#### func  BuildCorpus

```go
func BuildCorpus(scanner *bufio.Scanner) ([255][]string, error)
```
ReadCorpus is a helper function that reads a bufio.Scanner of words and returns
an array of word lengths, each containing an array of words of that length.

#### func  ReadCorpus

```go
func ReadCorpus(filepath string) ([255][]string, error)
```
ReadCorpus is a helper function that opens and reads a corpus file of words and
returns an array of word lengths, each containing an array of words of that
length.

#### type Mangle

```go
type Mangle struct {
	// Corpus of words to use as replacements. An array of word lengths, each
	// containing an array of words of that length.
	Corpus [255][]string
	// A sufficiently long secret, used as a salt so rainbow tables cannot be
	// used to reverse the hashes.
	Secret string
}
```

Configures a Mangle prior to use.

#### func (Mangle) MangleHTML

```go
func (m Mangle) MangleHTML(r io.Reader, w io.Writer) error
```
MangleHTML operates on HTML using an io interface, preserving all HTML tags
(including tag attributes), but mangling all content around and between tags.

#### func (Mangle) MangleIO

```go
func (m Mangle) MangleIO(r io.Reader, w io.Writer) error
```
MangleIO operates on an io interface, parsing as plain text, and is preferable
for long strings.

#### func (Mangle) MangleString

```go
func (m Mangle) MangleString(s string) string
```
MangleString operates on strings, and is preferable if you have many short
strings to operate on.
