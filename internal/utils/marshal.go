package utils

import (
	"log/slog"
	"reflect"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

const (
	blockDefinition    = "BlockDefinition"
	hcl                = "hcl"
	traversalSeparator = "."
	noRefTag           = "noref"
)

type BlockDefinition struct {
	Type   string
	Name   string
	Labels []string
}

func Marshal(v any) ([]byte, error) {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	marshal(val, &hclTags{}, &bodyStruct{rootBody}, 0)

	return f.Bytes(), nil
}

func marshal(val reflect.Value, tag *hclTags, body *bodyStruct, depth int) {
	switch val.Kind() {
	case reflect.String:
		noRef := findNoRef(tag.labels)
		if !noRef {
			if strings.Contains(val.String(), traversalSeparator) {
				body.traversal(tag.type_, val.String())
				break
			}
		}
		body.string(tag.type_, val.String())

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		body.int(tag.type_, val.Int())

	case reflect.Float32, reflect.Float64:
		body.float(tag.type_, val.Float())

	case reflect.Struct:

		isBlock, tags := blockDefinitionLookUp(val)
		switch isBlock {
		case true:
			block := hclwrite.NewBlock(tags.type_, append([]string{tags.name}, tags.labels...))
			for i := 0; i < val.NumField(); i++ {

				currentField := val.Field(i)
				if currentField.Kind() == reflect.Ptr {
					currentField = currentField.Elem()
				}

				if val.Type().Field(i).Name == blockDefinition {
					continue
				}

				fieldTags := val.Type().Field(i).Tag.Get(hcl)
				marshal(currentField, getTags(fieldTags), &bodyStruct{block.Body()}, depth)
			}
			body.AppendBlock(block)
		case false:
			body.SetAttributeRaw(tag.type_, hclwrite.Tokens{
				&hclwrite.Token{
					Type:  hclsyntax.TokenOBrace,
					Bytes: []byte("{"),
				},
			})
			for i := 0; i < val.NumField(); i++ {
				currentField := val.Field(i)
				if currentField.Kind() == reflect.Ptr {
					currentField = currentField.Elem()
				}

				fieldTags := val.Type().Field(i).Tag.Get(hcl)
				marshal(currentField, getTags(fieldTags), body, depth+1)
			}

			body.AppendUnstructuredTokens(hclwrite.Tokens{
				&hclwrite.Token{
					Type:  hclsyntax.TokenCBrace,
					Bytes: []byte("}"),
				},
				&hclwrite.Token{
					Type:  hclsyntax.TokenNewline,
					Bytes: []byte("\n"),
				},
			})
		}

	case reflect.Slice:
		body.SetAttributeRaw(tag.type_, hclwrite.Tokens{
			&hclwrite.Token{
				Type:  hclsyntax.TokenOBrace,
				Bytes: []byte("["),
			},
		})

		depth++
		noRef := findNoRef(tag.labels)
		for j := 0; j < val.Len(); j++ {
			elem := val.Index(j)

			switch elem.Kind() {
			case reflect.Struct:

				indentTokens := indent(depth)
				body.AppendUnstructuredTokens(indentTokens)
				body.AppendUnstructuredTokens(hclwrite.Tokens{
					&hclwrite.Token{
						Type:  hclsyntax.TokenCBrace,
						Bytes: []byte("{"),
					},
					&hclwrite.Token{
						Type:  hclsyntax.TokenNewline,
						Bytes: []byte("\n"),
					},
				})

				depth++
				for k := 0; k < elem.NumField(); k++ {
					indentTokens = indent(depth)
					body.AppendUnstructuredTokens(indentTokens)

					currentField := elem.Field(k)
					if currentField.Kind() == reflect.Ptr {
						currentField = currentField.Elem()
					}

					fieldTags := elem.Type().Field(k).Tag.Get(hcl)
					marshal(currentField, getTags(fieldTags), body, depth)
				}
				depth--

				indentTokens = indent(depth)
				body.AppendUnstructuredTokens(indentTokens)
				body.AppendUnstructuredTokens(hclwrite.Tokens{
					&hclwrite.Token{
						Type:  hclsyntax.TokenCBrace,
						Bytes: []byte("}"),
					},
					&hclwrite.Token{
						Type:  hclsyntax.TokenNewline,
						Bytes: []byte("\n"),
					},
				})

			default:
				if !noRef {
					if strings.Contains(elem.String(), traversalSeparator) {
						body.AppendUnstructuredTokens(stringToTraversal(elem.String()))
						break
					}
				}
				body.AppendUnstructuredTokens(stringToBytes(elem.String()))
			}
		}
		depth--

		indentTokens := space(depth + 1)
		body.AppendUnstructuredTokens(indentTokens)
		body.AppendUnstructuredTokens(hclwrite.Tokens{
			&hclwrite.Token{
				Type:  hclsyntax.TokenCBrace,
				Bytes: []byte("]"),
			},
			&hclwrite.Token{
				Type:  hclsyntax.TokenNewline,
				Bytes: []byte("\n"),
			},
		})

	default:
	}
}

type bodyStruct struct {
	*hclwrite.Body
}

func (h *bodyStruct) traversal(tag, value string) {
	h.SetAttributeRaw(
		tag,
		hclwrite.Tokens{
			&hclwrite.Token{
				Type:  hclsyntax.TokenIdent,
				Bytes: []byte(value),
			},
		},
	)
}

func (h *bodyStruct) string(tag, value string) {
	h.SetAttributeValue(tag, cty.StringVal(value))
}

func (h *bodyStruct) int(tag string, value int64) {
	h.SetAttributeValue(tag, cty.NumberIntVal(value))
}

func (h *bodyStruct) float(tag string, value float64) {
	h.SetAttributeValue(tag, cty.NumberFloatVal(value))
}

type hclTags struct {
	type_  string
	name   string
	labels []string
}

func getTags(currentTags string) *hclTags {
	splitTags := strings.Split(currentTags, ",")
	t := &hclTags{
		type_: splitTags[0],
	}
	if len(splitTags) > 0 {
		t.labels = splitTags[1:]
		for i, tag := range t.labels {
			t.labels[i] = strings.ToLower(tag)
		}
	}
	return t
}

func indent(depth int) hclwrite.Tokens {
	tokens := hclwrite.Tokens{}
	for i := 0; i < depth; i++ {
		tokens = append(tokens, &hclwrite.Token{
			Type:  hclsyntax.TokenIdent,
			Bytes: []byte("\t"),
		})
	}
	return tokens
}

func stringToBytes(s string) hclwrite.Tokens {
	tokens := hclwrite.Tokens{}
	tokens = append(tokens, &hclwrite.Token{
		Bytes: []byte("\""),
		Type:  hclsyntax.TokenOQuote,
	})
	tokens = append(tokens, &hclwrite.Token{
		Bytes: []byte(s),
	})
	tokens = append(tokens, &hclwrite.Token{
		Bytes: []byte("\""),
		Type:  hclsyntax.TokenCQuote,
	})

	return tokens
}

func stringToTraversal(s string) hclwrite.Tokens {
	tokens := hclwrite.Tokens{}
	tokens = append(tokens, &hclwrite.Token{
		Bytes: []byte(s),
	})
	return tokens
}

func space(depth int) hclwrite.Tokens {
	tokens := hclwrite.Tokens{}
	for i := 0; i < depth; i++ {
		tokens = append(tokens, &hclwrite.Token{
			Bytes: []byte(" "),
		})
	}
	return tokens
}

func parseBlockDefinition(val reflect.Value) *hclTags {
	t := &hclTags{labels: make([]string, 0)}

	for j := 0; j < val.NumField(); j++ {
		fieldName := val.Type().Field(j).Name
		field := val.Field(j)
		switch fieldName {
		case "Type":
			t.type_ = field.String()
		case "Name":
			t.name = field.String()
		case "Labels":
			slice, ok := field.Interface().([]string)
			if !ok {
				slog.Error("")
			}
			t.labels = append(t.labels, slice...)
		}
	}

	return t
}

func blockDefinitionLookUp(val reflect.Value) (bool, *hclTags) {
	for i := 0; i < val.NumField(); i++ {
		currentField := val.Field(i)
		if currentField.Kind() == reflect.Ptr {
			currentField = currentField.Elem()
		}

		if val.Type().Field(i).Name == blockDefinition {
			return true, parseBlockDefinition(currentField)
		}
	}
	return false, nil
}

func findNoRef(labels []string) bool {
	for _, label := range labels {
		if strings.Contains(label, noRefTag) {
			return true
		}
	}
	return false
}
