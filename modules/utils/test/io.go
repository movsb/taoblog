package test_utils

import (
	"io"
	"os"

	"gopkg.in/yaml.v2"
)

func MustLoadCasesFromYaml[T any](path string) []*T {
	fp, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer fp.Close()

	return MustLoadCasesFromYamlReader[T](fp)
}

func MustLoadCasesFromYamlReader[T any](r io.Reader) []*T {
	var t []*T

	if err := yaml.NewDecoder(r).Decode(&t); err != nil {
		panic(err)
	}

	return t
}
