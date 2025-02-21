package main

import (
	"fmt"
	"os"
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

func (p *Parser) parseNumber() error {
	if p.idx >= len(p.tokens) {
		return fmt.Errorf("expected string, but got end of input")
	}

	if p.tokens[p.idx].Key != TOK_NUMBER {
		return fmt.Errorf("expected string, but got '%s'", p.tokens[p.idx].Value)
	}
	p.consume()
	return nil
}

func (p *Parser) parseString() error {
	if p.idx >= len(p.tokens) {
		return fmt.Errorf("expected string, but got end of input")
	}

	if p.tokens[p.idx].Key != TOK_STRING {
		return fmt.Errorf("expected string, but got '%s'", p.tokens[p.idx].Value)
	}
	p.consume()
	return nil
}

func (p *Parser) parseBool() error {
	if p.idx >= len(p.tokens) {
		return fmt.Errorf("expected string, but got end of input")
	}

	if p.tokens[p.idx].Key != TOK_BOOL {
		return fmt.Errorf("expected string, but got '%s'", p.tokens[p.idx].Value)
	}
	p.consume()
	return nil
}

func (p *Parser) parsePrimitive() error {
	// <number> | <string> | <boolean>
	err := p.parseNumber()
	if err == nil {
		return nil
	}

	err = p.parseString()
	if err == nil {
		return nil
	}
	err = p.parseBool()
	if err == nil {
		return nil
	}
	return fmt.Errorf("ERROR: No number, string or boolean found")
}

func (p *Parser) parseMember() error {
	// <string> ': ' <json>
	err := p.parseString()
	if err != nil {
		return err
	}

	err = p.parseLiteral(":")
	if err != nil {
		return err
	}

	err = p.Parse()
	if err != nil {
		return err
	}

	return nil
}

func (p *Parser) parseObject() error {
	// '{' [ <member> *(', ' <member>) ] '}'

	err := p.parseLiteral("{")
	if err != nil {
		return err
	}

	firstElement := true
	for {
		currentToken := p.peek()
		if currentToken.Value == "}" {
			break
		}

		if !firstElement {
			err = p.parseLiteral(",")
			if err != nil {
				if currentToken.Value != "}" {
					return err
				}
				break
			}
		}

		err = p.parseMember()
		if err != nil {
			return err
		}

		firstElement = false
	}

	err = p.parseLiteral("}")
	if err != nil {
		return err
	}
	return nil
}

func (p *Parser) parseArray() error {
	// <array> ::= '[' [ <json> *(', ' <json>) ] ']'

	err := p.parseLiteral("[")
	if err != nil {
		return err
	}

	firstElement := true
	for {
		currentToken := p.peek()
		if currentToken.Value == "]" {
			break
		}

		if !firstElement {
			err = p.parseLiteral(",")
			if err != nil {
				// If comma is missing but we still have tokens, it could be an error
				if currentToken.Value != "]" {
					return err
				}
				break // Allow missing comma before closing bracket
			}
		}

		err = p.Parse()
		if err != nil {
			return err
		}

		firstElement = false
	}

	err = p.parseLiteral("]")
	if err != nil {
		return err
	}
	return nil
}

func (p *Parser) parseContainer() error {
	// <container> ::= <object> | <array>
	err := p.parseObject()
	if err == nil {
		return nil
	}

	err = p.parseArray()
	if err == nil {
		return nil
	}
	return fmt.Errorf("ERROR: No object or array found")
}

func (p *Parser) Parse() error {
	if p.idx >= len(p.tokens) {
		return fmt.Errorf("unexpected end of input")
	}

	err := p.parsePrimitive()
	if err == nil {
		return nil
	}
	err = p.parseContainer()
	if err == nil {
		return nil
	}

	return fmt.Errorf("unsupported token type: %s", p.tokens[p.idx].Value)
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
	p.Parse()
}
