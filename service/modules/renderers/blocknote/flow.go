package blocknote

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type Block struct {
	ID    string         `json:"id"`
	Type  string         `json:"type"`
	Props map[string]any `json:"props"`

	Content  *Content `json:"content"`
	Children []*Block `json:"children"`
}

type Inline struct {
	Type   string         `json:"type"`
	Text   string         `json:"text"`
	Styles map[string]any `json:"styles"`
	// ...
}

type Content struct {
	Inlines []*Inline
	Object  any
}

type ContentObjectType struct {
	Type string `json:"type"`
}

type TableContent struct {
	ContentObjectType
	ColumnWidths []*int `json:"columnWidths"`
	HeaderRows   int    `json:"headerRows"`
	HeaderCols   int    `json:"headerCols"`
	Rows         []struct {
		Cells []*Block `json:"cells"`
	} `json:"rows"`
}

func (c *Content) UnmarshalJSON(data []byte) error {
	if data[0] == '[' {
		dec := json.NewDecoder(bytes.NewReader(data))
		dec.DisallowUnknownFields()
		return dec.Decode(&c.Inlines)
	}
	if data[0] == '{' {
		dec := json.NewDecoder(bytes.NewReader(data))
		ty := ContentObjectType{}
		if err := dec.Decode(&ty); err != nil {
			return err
		}
		switch ty.Type {
		case `tableContent`:
			dec := json.NewDecoder(bytes.NewReader(data))
			dec.UseNumber()
			dec.DisallowUnknownFields()
			typed := TableContent{}
			if err := dec.Decode(&typed); err != nil {
				return fmt.Errorf(`error decoding table content: %w`, err)
			}
			c.Object = &typed
			return nil
		default:
			return fmt.Errorf(`unknown blocknote content: %s`, ty.Type)
		}
	}
	return fmt.Errorf(`cannot decode blocknote content: %s`, string(data))
}
