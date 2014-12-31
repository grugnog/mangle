package mangle_test

import (
	"fmt"
	"github.com/grugnog/mangle"
	"log"
)

// Reads a corpus from a file, initializes a Mangle, runs a string through it
// and prints the output.
func Example() {
	corpus, err := mangle.ReadCorpus("corpus.txt")
	if err != nil {
		log.Fatalf("Corpus read error: %s", err)
	}
	mangler := mangle.Mangle{corpus, "replace-with-a-secure-passphrase"}
	out := mangler.MangleString("Hello world!")
	fmt.Println(out)
}
