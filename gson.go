package gson

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

// JSON BNF Grammar: https://gist.github.com/EndingCredits/c12a9a99f87fd34d81df86b5588d3f1d

/** GLOBAL CONSTANTS */
const tok_end = -1
const (
	tok_number = iota
	tok_string
	tok_bool
	tok_null

	tok_open_sq_paran
	tok_close_sq_paran
	tok_open_curly_paran
	tok_close_curly_paran

	tok_comma
	tok_colon
)

/**
 * lexer Implementation
 */
type token struct {
	key   int
	value string
}

var literal_tokens = []token{
	{value: "{", key: tok_open_curly_paran},
	{value: "}", key: tok_close_curly_paran},
	{value: "[", key: tok_open_sq_paran},
	{value: "]", key: tok_close_sq_paran},
	{value: ":", key: tok_colon},
	{value: ",", key: tok_comma},
	{value: "true", key: tok_bool},
	{value: "false", key: tok_bool},
	{value: "null", key: tok_null},
}

type lexer struct {
	input  string
	cursor int
}

func (l *lexer) chop(n int) {
	for range n {
		if l.cursor >= len(l.input) {
			panic("chop Invalid")
		}
		l.cursor = l.cursor + 1
	}
}
func (l *lexer) next() token {

	t := token{
		key:   tok_end,
		value: "",
	}

	for l.cursor < len(l.input) && unicode.IsSpace(rune(l.input[l.cursor])) {
		l.chop(1)
	}

	if l.cursor >= len(l.input) {
		return t
	}

	for _, lt := range literal_tokens {
		tokenLength := len(lt.value)
		if l.cursor+tokenLength <= len(l.input) && string(l.input[l.cursor:l.cursor+tokenLength]) == lt.value {
			t.key = lt.key
			t.value = lt.value
			l.chop(tokenLength)
			return t
		}
	}

	if l.input[l.cursor] == '"' {
		l.cursor++
		var prev byte
		var strBuilder strings.Builder

		for l.cursor < len(l.input) {
			current := l.input[l.cursor]

			if current == '\\' && prev != '\\' {
				prev = current
				l.cursor++
				continue
			}

			if current == '"' && prev != '\\' {
				break
			}

			strBuilder.WriteByte(current)
			prev = current
			l.cursor++
		}
		t.key = tok_string
		t.value = strBuilder.String()
		l.chop(1) // chop ending quote
		return t
	}

	if unicode.IsDigit(rune(l.input[l.cursor])) {
		t.key = tok_number
		tokenStart := l.cursor

		for l.cursor < len(l.input) && unicode.IsDigit(rune(l.input[l.cursor])) {
			l.chop(1)
		}

		if l.cursor < len(l.input) && rune(l.input[l.cursor]) == '.' {
			l.chop(1)

			for l.cursor < len(l.input) && unicode.IsDigit(rune(l.input[l.cursor])) {
				l.chop(1)
			}
		}
		t.value = string(l.input[tokenStart:l.cursor])
		fmt.Println(t.value)
		return t
	}

	l.cursor = l.cursor + 1
	fmt.Printf("ERROR: Invalid character %c", l.input[l.cursor])
	return t
}

func newlexer(input string) lexer {
	return lexer{
		input:  input,
		cursor: 0,
	}
}

/**
 *
 * parser Implementation
 *
 */

type parser struct {
	tokens []token
	idx    int
}

func (p *parser) peek() token {
	if p.idx < len(p.tokens) {
		return p.tokens[p.idx]
	}
	return token{}
}

func (p *parser) consume() {
	if p.idx < len(p.tokens) {
		p.idx++
	}
}

func (p *parser) parseLiteral(expected string) error {
	if p.idx >= len(p.tokens) {
		return fmt.Errorf("expected '%s', but got end of input", expected)
	}

	currenttoken := p.tokens[p.idx]
	if currenttoken.value != expected {
		return fmt.Errorf("expected '%s', but got '%s'", expected, currenttoken.value)
	}
	p.consume()
	return nil
}

func (p *parser) parseNumber() (float64, error) {
	if p.idx >= len(p.tokens) {
		return -1, fmt.Errorf("expected number, but got end of input")
	}

	if p.tokens[p.idx].key != tok_number {
		return -1, fmt.Errorf("expected number, but got '%s'", p.tokens[p.idx].value)
	}
	value := p.tokens[p.idx].value
	p.consume()
	return strconv.ParseFloat(value, 64)
}

func (p *parser) parseString() (string, error) {
	if p.idx >= len(p.tokens) {
		return "", fmt.Errorf("expected string, but got end of input")
	}

	if p.tokens[p.idx].key != tok_string {
		return "", fmt.Errorf("expected string, but got '%s'", p.tokens[p.idx].value)
	}
	value := p.tokens[p.idx].value
	p.consume()
	return value, nil
}

func (p *parser) parseBool() (bool, error) {
	if p.idx >= len(p.tokens) {
		return false, fmt.Errorf("expected boolean, but got end of input")
	}

	if p.tokens[p.idx].key != tok_bool {
		return false, fmt.Errorf("expected boolean, but got '%s'", p.tokens[p.idx].value)
	}
	value := p.tokens[p.idx].value
	p.consume()
	return strconv.ParseBool(value)
}

func (p *parser) parseNull() (interface{}, error) {
	if p.idx >= len(p.tokens) {
		return -1, fmt.Errorf("expected string, but got end of input")
	}

	if p.tokens[p.idx].key != tok_null {
		return -1, fmt.Errorf("expected string, but got '%s'", p.tokens[p.idx].value)
	}
	p.consume()
	return nil, nil
}

func (p *parser) parsePrimitive() (interface{}, error) {
	// <number> | <string> | <boolean> | <null>
	if err := p.peek(); err.key == tok_number {
		return p.parseNumber()
	} else if err := p.peek(); err.key == tok_string {
		return p.parseString()
	} else if err := p.peek(); err.key == tok_bool {
		return p.parseBool()
	} else if err := p.peek(); err.key == tok_null {
		return p.parseNull()
	}
	return nil, fmt.Errorf("ERROR: No valid primitive found")
}

func (p *parser) parseMember() (map[string]interface{}, error) {
	// <string> ': ' <json>
	key, err := p.parseString()
	if err != nil {
		return nil, err
	}

	if err = p.parseLiteral(":"); err != nil {
		return nil, err
	}

	value, err := p.parse()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{key: value}, nil
}

func (p *parser) parseObject() (map[string]interface{}, error) {
	// '{' [ <member> *(', ' <member>) ] '}'

	if p.idx >= len(p.tokens) {
		return nil, fmt.Errorf("unexpected end of input")
	}

	if err := p.parseLiteral("{"); err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	firstElement := true
	for {
		currenttoken := p.peek()
		if currenttoken.value == "}" {
			break
		}

		if !firstElement {
			if err := p.parseLiteral(","); err != nil {
				if currenttoken.value != "}" {
					return nil, fmt.Errorf("Expected \",\" or \"}\" \n")
				} else {
					return nil, fmt.Errorf("Expected \",\" \n")
				}
			}
		}

		member, err := p.parseMember()
		if err != nil {
			return nil, err
		}
		for k, v := range member {
			result[k] = v
		}

		firstElement = false
	}

	if err := p.parseLiteral("}"); err != nil {
		return nil, err
	}
	return result, nil
}

func (p *parser) parseArray() ([]interface{}, error) {
	// <array> ::= '[' [ <json> *(', ' <json>) ] ']'

	if p.idx >= len(p.tokens) {
		return nil, fmt.Errorf("unexpected end of input")
	}

	if err := p.parseLiteral("["); err != nil {
		return nil, err
	}

	var result []interface{}
	firstElement := true
	for {
		currenttoken := p.peek()
		if currenttoken.value == "]" {
			break
		}

		if !firstElement {
			if err := p.parseLiteral(","); err != nil {
				if currenttoken.value != "}" {
					return nil, err
				}
				break
			}
		}

		value, err := p.parse()
		if err != nil {
			return nil, err
		}

		result = append(result, value)
		firstElement = false
	}

	if err := p.parseLiteral("]"); err != nil {
		return nil, err
	}
	return result, nil
}

func (p *parser) parseContainer() (interface{}, error) {
	// <container> ::= <object> | <array>

	if p.idx >= len(p.tokens) {
		return nil, fmt.Errorf("unexpected end of input")
	}

	currenttoken := p.peek()
	if currenttoken.value == "{" {
		objMap, err := p.parseObject()
		if err == nil {
			return objMap, nil
		}
	}
	if currenttoken.value == "[" {
		arrList, err := p.parseArray()
		if err == nil {
			return arrList, nil
		}
	}

	return nil, fmt.Errorf("ERROR: No object or array found")
}

func (p *parser) parse() (interface{}, error) {
	if p.idx >= len(p.tokens) {
		return nil, fmt.Errorf("unexpected end of input")
	}

	primitive, err := p.parsePrimitive()
	if err == nil {
		return primitive, nil
	}
	container, err := p.parseContainer()
	if err == nil {
		return container, nil
	}

	return nil, fmt.Errorf("unsupported token type")
}

func Parse(input string) (interface{}, error) {
	l := newlexer(input)
	tokens := make([]token, 0)
	for {
		t := l.next()
		if t.key == tok_end {
			break
		}
		tokens = append(tokens, t)
	}
	p := parser{
		tokens: tokens,
	}
	return p.parse()
}

func Stringify(input interface{}) (string, error) {
	var builder strings.Builder
	err := stringifyValue(input, &builder)
	if err != nil {
		return "", err
	}
	return builder.String(), nil
}

func stringifyValue(in interface{}, builder *strings.Builder) error {

	if in == nil {
		builder.WriteString("null")
		return nil
	}

	val := reflect.ValueOf(in)

	switch val.Kind() {
	case reflect.String:
		builder.WriteString(`"` + val.String() + `"`)

	case reflect.Int:
		builder.WriteString(strconv.Itoa(int(val.Int())))

	case reflect.Bool:
		builder.WriteString(strconv.FormatBool(val.Bool()))

	case reflect.Slice, reflect.Array:
		builder.WriteString("[")
		for i := 0; i < val.Len(); i++ {
			if i > 0 {
				builder.WriteString(",")
			}
			err := stringifyValue(val.Index(i).Interface(), builder)
			if err != nil {
				return err
			}
		}
		builder.WriteString("]")
		return nil
	case reflect.Map:
		builder.WriteString("{")
		keys := val.MapKeys()
		for i, keyVal := range keys {
			if i > 0 {
				builder.WriteString(",")
			}

			key := keyVal.Interface()

			keyString, ok := key.(string)
			if !ok {
				return fmt.Errorf("map keys must be strings")
			}

			builder.WriteString(`"` + escapeString(keyString) + `":`)
			err := stringifyValue(val.MapIndex(keyVal).Interface(), builder)
			if err != nil {
				return err
			}
		}
		builder.WriteString("}")
		return nil

	default:
		return fmt.Errorf("unsupported type: %v", val.Kind())
	}
	return nil
}

func escapeString(s string) string {
	var builder strings.Builder
	for _, r := range s {
		switch r {
		case '\\':
			builder.WriteString(`\\`)
		case '"':
			builder.WriteString(`\"`)
		case '\n':
			builder.WriteString(`\n`)
		case '\r':
			builder.WriteString(`\r`)
		case '\t':
			builder.WriteString(`\t`)
		case '\b':
			builder.WriteString(`\b`)
		case '\f':
			builder.WriteString(`\f`)
		default:
			if r < 32 {
				builder.WriteString(fmt.Sprintf(`\u%04X`, int(r)))
			} else {
				builder.WriteRune(r)
			}
		}
	}
	return builder.String()
}
