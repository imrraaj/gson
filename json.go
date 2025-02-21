package json

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
)

// JSON BNF Grammar
// https://gist.github.com/EndingCredits/c12a9a99f87fd34d81df86b5588d3f1d

/**
 *
 * GLOBAL CONSTANTS
 *
 */
const TOK_END = -1
const (
	TOK_NUMBER = iota
	TOK_STRING
	TOK_BOOL

	TOK_OPEN_SQ_PARAN
	TOK_CLOSE_SQ_PARAN
	TOK_OPEN_CURLY_PARAN
	TOK_CLOSE_CURLY_PARAN

	TOK_COMMA
	TOK_COLON
)

/**
 *
 * Lexer Implementation
 *
 */

type Token struct {
	Key   int
	Value string
}

var literal_tokens = []Token{
	{Value: "{", Key: TOK_OPEN_CURLY_PARAN},
	{Value: "}", Key: TOK_CLOSE_CURLY_PARAN},
	{Value: "[", Key: TOK_OPEN_SQ_PARAN},
	{Value: "]", Key: TOK_CLOSE_SQ_PARAN},
	{Value: ":", Key: TOK_COLON},
	{Value: ",", Key: TOK_COMMA},
	{Value: "true", Key: TOK_BOOL},
	{Value: "false", Key: TOK_BOOL},
	{Value: "null", Key: TOK_BOOL},
}

type Lexer struct {
	Input  string
	Cursor int
}

func (l *Lexer) Chop(n int) {
	for range n {
		if l.Cursor >= len(l.Input) {
			panic("Chop Invalid")
		}
		l.Cursor = l.Cursor + 1
	}
}
func (l *Lexer) Next() Token {

	t := Token{
		Key:   TOK_END,
		Value: "",
	}

	for l.Cursor < len(l.Input) && unicode.IsSpace(rune(l.Input[l.Cursor])) {
		l.Chop(1)
	}

	if l.Cursor >= len(l.Input) {
		return t
	}

	for _, lt := range literal_tokens {
		tokenLength := len(lt.Value)
		if l.Cursor+tokenLength <= len(l.Input) && string(l.Input[l.Cursor:l.Cursor+tokenLength]) == lt.Value {
			t.Key = lt.Key
			t.Value = lt.Value
			l.Chop(tokenLength)
			return t
		}
	}

	if l.Input[l.Cursor] == '"' {
		l.Cursor++
		var prev byte
		var strBuilder strings.Builder

		for l.Cursor < len(l.Input) {
			current := l.Input[l.Cursor]

			if current == '\\' && prev != '\\' {
				prev = current
				l.Cursor++
				continue
			}

			if current == '"' && prev != '\\' {
				break
			}

			strBuilder.WriteByte(current)
			prev = current
			l.Cursor++
		}
		t.Key = TOK_STRING
		t.Value = strBuilder.String()
		l.Chop(1) // chop ending quote
		return t
	}

	if unicode.IsDigit(rune(l.Input[l.Cursor])) {
		t.Key = TOK_NUMBER
		tokenStart := l.Cursor

		for l.Cursor < len(l.Input) && unicode.IsDigit(rune(l.Input[l.Cursor])) {
			l.Chop(1)
		}
		t.Value = string(l.Input[tokenStart:l.Cursor])
		return t
	}

	l.Cursor = l.Cursor + 1
	fmt.Printf("ERROR: Invalid character %c", l.Input[l.Cursor])
	return t
}

func NewLexer(input string) Lexer {
	return Lexer{
		Input:  input,
		Cursor: 0,
	}
}

/**
 *
 * Parser Implementation
 *
 */

type Parser struct {
	tokens []Token
	idx    int
}

func (p *Parser) peek() Token {
	if p.idx < len(p.tokens) {
		return p.tokens[p.idx]
	}
	return Token{}
}

func (p *Parser) consume() {
	fmt.Printf("consumed %s\n", p.tokens[p.idx].Value)
	if p.idx < len(p.tokens) {
		p.idx++
	}
}

func (p *Parser) parseLiteral(expected string) error {
	if p.idx >= len(p.tokens) {
		return fmt.Errorf("expected '%s', but got end of input", expected)
	}

	currentToken := p.tokens[p.idx]
	if currentToken.Value != expected {
		return fmt.Errorf("expected '%s', but got '%s'", expected, currentToken.Value)
	}
	p.consume()
	return nil
}

func (p *Parser) parseNumber() (int, error) {
	if p.idx >= len(p.tokens) {
		return -1, fmt.Errorf("expected string, but got end of input")
	}

	if p.tokens[p.idx].Key != TOK_NUMBER {
		return -1, fmt.Errorf("expected string, but got '%s'", p.tokens[p.idx].Value)
	}
	value := p.tokens[p.idx].Value
	p.consume()
	return strconv.Atoi(value)
}

func (p *Parser) parseString() (string, error) {
	if p.idx >= len(p.tokens) {
		return "", fmt.Errorf("expected string, but got end of input")
	}

	if p.tokens[p.idx].Key != TOK_STRING {
		return "", fmt.Errorf("expected string, but got '%s'", p.tokens[p.idx].Value)
	}
	value := p.tokens[p.idx].Value
	p.consume()
	return value, nil
}

func (p *Parser) parseBool() (bool, error) {
	if p.idx >= len(p.tokens) {
		return false, fmt.Errorf("expected string, but got end of input")
	}

	if p.tokens[p.idx].Key != TOK_BOOL {
		return false, fmt.Errorf("expected string, but got '%s'", p.tokens[p.idx].Value)
	}
	value := p.tokens[p.idx].Value
	p.consume()
	return strconv.ParseBool(value)
}

func (p *Parser) parsePrimitive() (interface{}, error) {
	// <number> | <string> | <boolean>
	if err := p.peek(); err.Key == TOK_NUMBER {
		return p.parseNumber()
	} else if err := p.peek(); err.Key == TOK_STRING {
		return p.parseString()
	} else if err := p.peek(); err.Key == TOK_BOOL {
		return p.parseBool()
	}
	return nil, fmt.Errorf("ERROR: No valid primitive found")
}

func (p *Parser) parseMember() (map[string]interface{}, error) {
	// <string> ': ' <json>
	key, err := p.parseString()
	if err != nil {
		return nil, err
	}

	if err = p.parseLiteral(":"); err != nil {
		return nil, err
	}

	value, err := p.Parse()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{key: value}, nil
}

func (p *Parser) parseObject() (map[string]interface{}, error) {
	// '{' [ <member> *(', ' <member>) ] '}'

	if err := p.parseLiteral("{"); err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	firstElement := true
	for {
		currentToken := p.peek()
		if currentToken.Value == "}" {
			break
		}

		if !firstElement {

			if err := p.parseLiteral(","); err != nil {
				if currentToken.Value != "}" {
					return nil, err
				}
				break
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

func (p *Parser) parseArray() ([]interface{}, error) {
	// <array> ::= '[' [ <json> *(', ' <json>) ] ']'

	if err := p.parseLiteral("["); err != nil {
		return nil, err
	}

	var result []interface{}
	firstElement := true
	for {
		currentToken := p.peek()
		if currentToken.Value == "]" {
			break
		}

		if !firstElement {
			if err := p.parseLiteral(","); err != nil {
				if currentToken.Value != "}" {
					return nil, err
				}
				break
			}
		}

		value, err := p.Parse()
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

func (p *Parser) parseContainer() (interface{}, error) {
	// <container> ::= <object> | <array>
	objMap, err := p.parseObject()
	if err == nil {
		return objMap, nil
	}

	arrList, err := p.parseArray()
	if err == nil {
		return arrList, nil
	}
	return nil, fmt.Errorf("ERROR: No object or array found")
}

func (p *Parser) Parse() (interface{}, error) {
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

	return nil, fmt.Errorf("unsupported token type: %s", p.tokens[p.idx].Value)
}

func NewParser(input string) Parser {
	l := NewLexer(input)
	tokens := make([]Token, 0)
	for {
		t := l.Next()
		if t.Key == TOK_END {
			break
		}
		tokens = append(tokens, t)
	}
	return Parser{
		tokens: tokens,
	}
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
	p := NewParser(string(data))
	val, err := p.Parse()
	fmt.Println(val)
}
