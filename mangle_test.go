package mangle_test

import (
	"bufio"
	"bytes"
	"github.com/grugnog/mangle"
	"strings"
	"testing"
)

var corpus = initTestCorpus()
var salta = "123456789012345678901234567890"
var saltb = "123456789012345678901234567890-"

func initTestCorpus() [255][]string {
	words := "a b c dd ee ff ggg hhh iii jjjj kkkk llll mmmmm nnnnn ooooo pppppp qqqqqq rrrrrr sssssss ttttttt uuuuuuu"
	reader := strings.NewReader(words)
	scanner := bufio.NewScanner(reader)
	corpus, _ := mangle.BuildCorpus(scanner)
	return corpus
}

var tests = []struct {
	in  string
	out string
}{
	// Long words, short words, capitalization, punctuation etc.
	{"A", "B"},
	{"Jumping", "Ttttttt"},
	{"The quick brown fox\njumps over the lazy dog", "Iii ooooo ooooo iii\nooooo jjjj iii llll hhh"},
	{"So funny, ROFL.", "Dd ooooo, KKKK."},
	{"+1 (123) 456-7890", "+c (iii) iii-jjjj"},
	{"Antidisestablishmentarianism", "Sssssss                     "},
	// Unicode tests.
	{"是一个专为语文教学而设计的电脑软件。在当今这个电脑时代，＂电脑辅助教学＂", "uuuuuuu          。ttttttt  ，＂rrrrrr＂"},
	{"よばれる　－　呼ばれる りゅうは、ごく　－　理由は、", "kkkk　－　llll jjjj、dd　－　ggg、"},
	{"авианосцы 'Уинсон' и 'Мидуэй' приблизились", "sssssss   'Pppppp' b 'Qqqqqq' uuuuuuu     "},
	{"صدر الدين عيني", "hhh ooooo jjjj"},
}

// Tests string based interface.
func TestStrings(t *testing.T) {
	mangler := mangle.Mangle{Corpus: corpus, Secret: salta}
	for _, tt := range tests {
		out := mangler.MangleString(tt.in)
		if out != tt.out {
			t.Errorf("MangleString(%s) => %q, want %q", tt.in, out, tt.out)
		}
	}
}

// Tests IO based interface.
func TestIO(t *testing.T) {
	mangler := mangle.Mangle{Corpus: corpus, Secret: salta}
	for _, tt := range tests {
		r := strings.NewReader(tt.in)
		w := new(bytes.Buffer)
		err := mangler.MangleIO(r, w)
		if err != nil {
			t.Errorf("MangleIO(%q) error %q", tt.in, err)
		}
		out := w.String()
		if out != tt.out {
			t.Errorf("MangleIO(%q) => %q, want %q", tt.in, out, tt.out)
		}
	}
}

// Tests HTML based interface with basic strings.
func TestHTMLBasic(t *testing.T) {
	mangler := mangle.Mangle{Corpus: corpus, Secret: salta}
	for _, tt := range tests {
		r := strings.NewReader(tt.in)
		w := new(bytes.Buffer)
		err := mangler.MangleHTML(r, w)
		if err != nil {
			t.Errorf("MangleHTML(%q) error %q", tt.in, err)
		}
		out := w.String()
		if out != tt.out {
			t.Errorf("MangleHTML(%q) => %q, want %q", tt.in, out, tt.out)
		}
	}
}

var markuptests = []struct {
	in  string
	out string
}{
	// Full page tests.
	{"<html><head><title>A Simple HTML Example</title></head><body><h2>HTML is Easy To Learn</h2><p>Welcome!</p></body></html>", "<html><head><title>B Pppppp KKKK Uuuuuuu</title></head><body><h2>KKKK ee Jjjj Ee Nnnnn</h2><p>Uuuuuuu!</p></body></html>"},
	{"<!doctype html><title>Short HTML5</title>", "<!doctype html><title>Mmmmm MMMMM</title>"},
	// Snippet tests.
	// TODO: Would be nice to be able to preserve select tag/attribute combinations (e.g. a:href).
	{"<h2>HTML is Easy To Learn</h2><p>Welcome to the world of the <a href=\"http://www.w3.org/\">World Wide Web</a>.</p>", "<h2>KKKK ee Jjjj Ee Nnnnn</h2><p>Uuuuuuu ff iii mmmmm ff iii <a href=\"http://www.w3.org/\">Nnnnn Llll Hhh</a>.</p>"},
	{"<article><header><h1>Blog post</h1></header><nav><ul><li><a href=\"..\">Next post</a></li></ul></nav><p>Some article content!</p></article>", "<article><header><h1>Jjjj kkkk</h1></header><nav><ul><li><a href=\"..\">Jjjj kkkk</a></li></ul></nav><p>Jjjj uuuuuuu sssssss!</p></article>"},
	// Embedded CSS and JS tests.
	// TODO: Would be nice to be able to whitelist these tags.
	{"<head><style>body {background-color:lightgray}</style></head><body><h1>This is a heading</h1></body>", "<head><style>kkkk {ttttttt   -nnnnn:uuuuuuu  }</style></head><body><h1>Llll ee a uuuuuuu</h1></body>"},
	{"<head><script type='text/javascript'>$(document).ready(function() {}}</script></head><body><h1>This is a heading</h1></body>", "<head><script type='text/javascript'>$(sssssss ).ooooo(sssssss () {}}</script></head><body><h1>Llll ee a uuuuuuu</h1></body>"},
}

// Tests HTML based interface with markup.
func TestHTMLMarkup(t *testing.T) {
	mangler := mangle.Mangle{Corpus: corpus, Secret: salta}
	for _, tt := range markuptests {
		r := strings.NewReader(tt.in)
		w := new(bytes.Buffer)
		err := mangler.MangleHTML(r, w)
		if err != nil {
			t.Errorf("MangleHTML(%q) error %q", tt.in, err.Error())
		}
		out := w.String()
		if out != tt.out {
			t.Errorf("MangleHTML(%q) => %q, want %q", tt.in, out, tt.out)
		}
	}
}

// Tests that output is dependent on the user defined salt.
func TestSalts(t *testing.T) {
	in := "The quick brown fox jumps over the lazy dog"
	correcta := "Iii ooooo ooooo iii ooooo jjjj iii llll hhh"
	correctb := "Ggg nnnnn nnnnn hhh mmmmm jjjj ggg llll iii"
	manglera := mangle.Mangle{Corpus: corpus, Secret: salta}
	outa := manglera.MangleString(in)
	if outa != correcta {
		t.Errorf("MangleString(%s) => %q, want %q", in, outa, correcta)
	}
	manglerb := mangle.Mangle{Corpus: corpus, Secret: saltb}
	outb := manglerb.MangleString(in)
	if outb != correctb {
		t.Errorf("MangleString(%s) => %q, want %q", in, outb, correctb)
	}
}
