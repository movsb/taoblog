package post_translators

import (
	"bytes"
	"os/exec"
	"strings"
)

type MarkdownTranslator struct {
	PostTranslator
}

func (me *MarkdownTranslator) Translate(source string, base string) (string, error) {
	var err error

	cmd := exec.Command(
		"node",
		"./service/modules/post_translators/markdown.js",
		base,
	)

	cmd.Stdin = strings.NewReader(source)

	strWriter := bytes.NewBuffer(nil)
	cmd.Stdout = strWriter

	if err = cmd.Run(); err != nil {
		return "", err
	}

	return strWriter.String(), nil
}
