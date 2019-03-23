package main

import (
	"bufio"
	"fmt"
	"os"
)

// StdinLineReader ...
type StdinLineReader struct {
	scanner *bufio.Scanner
}

// NewStdinLineReader news a
func NewStdinLineReader() *StdinLineReader {
	r := &StdinLineReader{}
	r.scanner = bufio.NewScanner(os.Stdin)
	return r
}

// MustReadLine must read a line.
func (r *StdinLineReader) MustReadLine(prompt string) string {
	str, err := r.ReadLine(prompt)
	if err != nil {
		panic(err)
	}
	return str
}

// ReadLine reads a line.
func (r *StdinLineReader) ReadLine(prompt string) (str string, err error) {
	if prompt != "" {
		fmt.Print(prompt, " ")
	}
	if !r.scanner.Scan() {
		return "", r.scanner.Err()
	}
	str = r.scanner.Text()
	err = r.scanner.Err()
	return
}
