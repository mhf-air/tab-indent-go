package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

/*
description:
  transforms the line-leading spaces into tabs

usage:
  go run a.go -input=b.vue -inplace

limitation:
    also transforms the line-leading spaces in multiline strings if that is not
  what you want
*/

var (
	flagTabWidth  int
	flagInputFile string
	flagInplace   bool
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `transforms the line-leading spaces into tabs

usage:
`)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
limitation:
    also transforms the line-leading spaces in multiline strings if that is not
  what you want
`)
	}

	setupFlag()
	buf := run()
	if flagInplace {
		ioutil.WriteFile(flagInputFile, buf, 0666)
	} else {
		fmt.Printf("%s", string(buf))
	}
}

func setupFlag() {
	tabwidth := flag.Int("tabwidth", 2, "how many spaces a tab corresponds, must be nonnegative")
	defer func() {
		if *tabwidth < 0 {
			panic("tabwidth must be nonnegative")
		}
		flagTabWidth = *tabwidth
	}()

	inputFile := flag.String("input", "", "input file name, if not set, will input from stdin")
	defer func() {
		flagInputFile = *inputFile
	}()

	inplace := flag.Bool("inplace", false, "edit in place, if not set, will output to stdout")
	defer func() {
		flagInplace = *inplace
	}()

	flag.Parse()
}

func run() []byte {
	var f io.Reader
	var err error

	if flagInputFile == "" {
		f = os.Stdin
	} else {
		f, err = os.Open(flagInputFile)
		ck(err)
	}
	buf, err := ioutil.ReadAll(f)
	ck(err)
	buflen := len(buf)

	atLineStart := true
	inSpace := false
	inTab := false

	spaceCount := 0
	start := 0

	r := make([]byte, 0, buflen)
	for i := 0; i < buflen; i++ {
		ch := buf[i]
		if atLineStart {
			start = i
			if ch == '\n' {
				r = append(r, '\n')
				continue
			}

			atLineStart = false
			if ch == '\t' {
				inTab = true
				continue
			}
			inTab = false

			if ch != ' ' {
				inSpace = false
				continue
			}

			r = append(r, buf[start:i]...)
			spaceCount = 1
			inSpace = true

		} else if inTab { // for cases like "\t\t    "
			if ch == '\n' {
				r = append(r, '\n')
				atLineStart = true
				continue
			}

			if ch != '\t' {
				inTab = false
				if ch != ' ' {
					continue
				}
				r = append(r, buf[start:i]...)
				spaceCount = 1
				inSpace = true
			}

		} else if inSpace {
			if ch == ' ' {
				spaceCount++
				continue
			}

			if ch == '\n' {
				r = append(r, '\n')
				atLineStart = true
				continue
			}

			tabCount := spaceCount / flagTabWidth
			for j := 0; j < tabCount; j++ {
				r = append(r, '\t')
			}

			start = i
			inSpace = false

		} else {
			if ch == '\n' {
				r = append(r, buf[start:i+1]...)
				atLineStart = true
			}
		}
	}
	if buflen > 0 && buf[buflen-1] != '\n' {
		r = append(r, buf[start:]...)
	}

	return r
}

func ck(err error) {
	if err != nil {
		panic(err)
	}
}
