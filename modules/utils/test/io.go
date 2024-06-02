package test_utils

import (
	"os"

	"gopkg.in/yaml.v2"
)

func MustLoadCasesFromYaml[T any](path string) []*T {
	var t []*T

	fp, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer fp.Close()

	if yaml.NewDecoder(fp).Decode(&t); err != nil {
		panic(err)
	}

	return t
}
