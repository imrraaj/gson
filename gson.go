package gson

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
)

// JSON BNF Grammar: https://gist.github.com/EndingCredits/c12a9a99f87fd34d81df86b5588d3f1d

/** GLOBAL CONSTANTS */
const TOK_END = -1
const (
	TOK_NUMBER = iota
	TOK_STRING
	TOK_BOOL
	TOK_NULL

	TOK_OPEN_SQ_PARAN
	TOK_CLOSE_SQ_PARAN
	TOK_OPEN_CURLY_PARAN
	TOK_CLOSE_CURLY_PARAN

	TOK_COMMA
	TOK_COLON
)

/**
 * lexer Implementation
 */
type token struct {
	key   int
	value string
}

var literal_tokens = []token{
	{value: "{", key: TOK_OPEN_CURLY_PARAN},
	{value: "}", key: TOK_CLOSE_CURLY_PARAN},
	{value: "[", key: TOK_OPEN_SQ_PARAN},
	{value: "]", key: TOK_CLOSE_SQ_PARAN},
	{value: ":", key: TOK_COLON},
	{value: ",", key: TOK_COMMA},
	{value: "true", key: TOK_BOOL},
	{value: "false", key: TOK_BOOL},
	{value: "null", key: TOK_NULL},
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
		key:   TOK_END,
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
		t.key = TOK_STRING
		t.value = strBuilder.String()
		l.chop(1) // chop ending quote
		return t
	}

	if unicode.IsDigit(rune(l.input[l.cursor])) {
		t.key = TOK_NUMBER
		tokenStart := l.cursor

		for l.cursor < len(l.input) && unicode.IsDigit(rune(l.input[l.cursor])) {
			l.chop(1)
		}
		t.value = string(l.input[tokenStart:l.cursor])
		return t
	}

	l.cursor = l.cursor + 1
	fmt.Printf("ERROR: Invalid character %c", l.input[l.cursor])
	return t
}

func Newlexer(input string) lexer {
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

func (p *parser) parseNumber() (int, error) {
	if p.idx >= len(p.tokens) {
		return -1, fmt.Errorf("expected string, but got end of input")
	}

	if p.tokens[p.idx].key != TOK_NUMBER {
		return -1, fmt.Errorf("expected string, but got '%s'", p.tokens[p.idx].value)
	}
	value := p.tokens[p.idx].value
	p.consume()
	return strconv.Atoi(value)
}

func (p *parser) parseString() (string, error) {
	if p.idx >= len(p.tokens) {
		return "", fmt.Errorf("expected string, but got end of input")
	}

	if p.tokens[p.idx].key != TOK_STRING {
		return "", fmt.Errorf("expected string, but got '%s'", p.tokens[p.idx].value)
	}
	value := p.tokens[p.idx].value
	p.consume()
	return value, nil
}

func (p *parser) parseBool() (bool, error) {
	if p.idx >= len(p.tokens) {
		return false, fmt.Errorf("expected string, but got end of input")
	}

	if p.tokens[p.idx].key != TOK_BOOL {
		return false, fmt.Errorf("expected string, but got '%s'", p.tokens[p.idx].value)
	}
	value := p.tokens[p.idx].value
	p.consume()
	return strconv.ParseBool(value)
}

func (p *parser) parseNull() (interface{}, error) {
	if p.idx >= len(p.tokens) {
		return -1, fmt.Errorf("expected string, but got end of input")
	}

	if p.tokens[p.idx].key != TOK_NULL {
		return -1, fmt.Errorf("expected string, but got '%s'", p.tokens[p.idx].value)
	}
	p.consume()
	return nil, nil
}

func (p *parser) parsePrimitive() (interface{}, error) {
	// <number> | <string> | <boolean> | <null>
	if err := p.peek(); err.key == TOK_NUMBER {
		return p.parseNumber()
	} else if err := p.peek(); err.key == TOK_STRING {
		return p.parseString()
	} else if err := p.peek(); err.key == TOK_BOOL {
		return p.parseBool()
	} else if err := p.peek(); err.key == TOK_NULL {
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
	l := Newlexer(input)
	tokens := make([]token, 0)
	for {
		t := l.next()
		if t.key == TOK_END {
			break
		}
		tokens = append(tokens, t)
	}
	p := parser{
		tokens: tokens,
	}
	return p.parse()
}

/**
 *
 * Main Function
 *
 */
func main() {

	args := os.Args
	if len(args) < 2 {
		fmt.Printf("USAGE: %s <filename>\n", args[0])
		return
	}

	filename := args[1]

	if filename == "" {
		fmt.Printf("ERROR: Please provide valid filename\n")
		return
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		panic("ERROR: Not able to read the file")
	}
	val, err := Parse(string(data))
	fmt.Println(val)
}
