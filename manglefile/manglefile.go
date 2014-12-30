/*
Simple command line text sanitization / data masking tool.

Usage

To print command help run:
		manglefile -h

Note that the -secret flag and a corpus containing a list of replacement words are required.
*/
package main

import (
	"flag"
	"fmt"
	"github.com/grugnog/mangle"
	"io"
	"log"
	"os"
	"runtime/pprof"
)

var corpus = flag.String("corpus", "corpus.txt", "File containing corpus of words to use as replacements.")
var secret = flag.String("secret", "", "Required. A secret, used as a salt - must be at least 16 characters.")
var filetype = flag.String("type", "", "The file type: \"text\" (default) or \"html\".")
var profile = flag.Bool("profile", false, "If set, performance profiling data will be stored in this file.")

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Simple command line text sanitization / data masking tool.")
		fmt.Fprintln(os.Stderr, "Accepts input on stdin and output on stdout.")
		fmt.Fprintln(os.Stderr, "Example: echo \"Hello world!\" | manglefile -corpus=corpus.txt -secret=replace-with-a-secure-passphrase\n")
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	// Check secret salt.
	if *secret == "" {
		flag.Usage()
		return
	}
	if len(*secret) < 16 {
		log.Fatalf("The secret must be at least 16 characters long.")
	}

	// Enable profiling if requested.
	if *profile == true {
		profile, err := os.Create("profile")
		if err != nil {
			log.Fatalf("Unable to open profile file.", err)
		}
		pprof.StartCPUProfile(profile)
	}

	// Read corpus.
	corpus, err := mangle.ReadCorpus(*corpus)
	if err != nil {
		log.Fatalf("Corpus read error: %s", err)
	}

	// Open stdin and stdout and mangle.
	w := io.Writer(os.Stdout)
	r := io.Reader(os.Stdin)
	mangler := mangle.Mangle{corpus, *secret}
	if *filetype == "html" {
		err = mangler.MangleHTML(r, w)
	} else {
		err = mangler.MangleIO(r, w)
	}
	if err != nil {
		log.Fatalf("IO error: %s", err.Error())
	}

	// Complete profiling.
	if *profile == true {
		pprof.StopCPUProfile()
	}
}
