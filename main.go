package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"unicode"
)

func main() {
	args := os.Args

	if len(args) != 2 {
		err := errors.New("incorrect number of arguments.\r\nUsage: main.go <input_file>")
		panic(err)
	}

	path := args[1]

	stat, err := os.Stat(path)

	if stat == nil || err != nil {
		err := errors.New("input file does not exist")
		panic(err)
	}

	file, err := os.Open(path)
	defer file.Close()

	if err != nil {
		panic(err)
	}

	buf := make([]byte, 1024)
	reader := bufio.NewReader(file)

	for {
		_, err := reader.Read(buf)

		if err == io.EOF {
			break
		}

		if err != nil {
			panic(err)
		}
	}

	t := Tokenizer{
		src: &buf,
		idx: 0,
	}

	tokens, err := t.tokenize()

	if err != nil {
		panic(err)
	}

	parser := Parser{
		tokens: &tokens,
		idx:    0,
	}

	tree, err := parser.parse()

	if err != nil {
		panic(err)
	}

	generator := Generator{root: tree}
	asm := generator.generate()

	write(asm)

	cmd := exec.Command("nasm", "-o", "out.o", "out.asm")

	_, err = cmd.Output()

	if err != nil {
		fmt.Println("Error executing command:", err)
		return
	}
}

type TokenType string

const (
	Return     TokenType = "return"
	IntLiteral TokenType = "int_literal"
	Semi       TokenType = "semicolon"
)

type Token struct {
	Type  TokenType
	Value *string
}

func write(asm *string) {
	file, err := os.Create("./out.asm")
	defer file.Close()

	if err != nil {
		panic(err)
	}

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	_, err = writer.WriteString(*asm)

	if err != nil {
		panic(err)
	}
}

type Tokenizer struct {
	src *[]byte
	idx int
}

func (t *Tokenizer) peak() (*uint8, error) {
	if t.idx >= len(*t.src) {
		return nil, errors.New("out of bounds")
	}

	char := (*t.src)[t.idx]

	return &char, nil
}

func (t *Tokenizer) consume() *uint8 {
	char := (*t.src)[t.idx]

	t.idx++

	return &char
}

func (t *Tokenizer) tokenize() ([]Token, error) {
	buf := make([]byte, 0)

	tokens := make([]Token, 0)

	for {
		p, err := t.peak()

		if err != nil {
			break
		}

		if unicode.IsLetter(rune(*p)) {
			for {
				p, err = t.peak()

				if err != nil {
					break
				}

				if !unicode.IsLetter(rune(*p)) {
					break
				}

				c := t.consume()
				buf = append(buf, *c)
			}

			if string(buf) == "return" {
				token := Token{Type: Return, Value: nil}
				tokens = append(tokens, token)
				buf = buf[:0]

				continue
			} else {
				return nil, errors.New("Invalid token: " + string(buf))
			}
		} else if unicode.IsNumber(rune(*p)) {
			for {
				p, err = t.peak()

				if err != nil {
					break
				}

				if !unicode.IsNumber(rune(*p)) {
					break
				}

				c := t.consume()
				buf = append(buf, *c)
			}

			str := string(buf)
			token := Token{Type: IntLiteral, Value: &str}
			tokens = append(tokens, token)
			buf = buf[:0]
			continue
		} else if rune(*p) == ';' {
			t.consume()
			token := Token{Type: Semi, Value: nil}
			tokens = append(tokens, token)
			buf = buf[:0]
			continue
		} else if unicode.IsSpace(rune(*p)) {
			t.consume()
			continue
		} else if *p == 0 {
			break
		} else {
			return nil, errors.New(fmt.Sprintf("invalid character at position %d: %c", t.idx, rune(*p)))
		}
	}

	return tokens, nil
}

type Parser struct {
	tokens *[]Token
	idx    int
}

func (p *Parser) peak() (*Token, error) {
	if p.idx >= len(*p.tokens) {
		return nil, errors.New("out of bounds")
	}

	char := (*p.tokens)[p.idx]

	return &char, nil
}

func (p *Parser) consume() *Token {
	token := (*p.tokens)[p.idx]

	p.idx++

	return &token
}

type NodeReturn struct {
	expr *NodeExpr
}

type NodeExpr struct {
	intLiteral *Token
}

func (p *Parser) parse() (*NodeReturn, error) {
	var returnNode *NodeReturn

	for {
		peak, err := p.peak()

		if err != nil {
			break
		}

		if peak.Type == Return {
			p.consume()
			expr, err := p.parseExpr()

			if err != nil {
				return nil, err
			}

			returnNode = &NodeReturn{expr: expr}

			newPeak, err := p.peak()

			if err != nil {
				panic(err)
			}

			if newPeak.Type == Semi {
				p.consume()
			} else {
				return nil, errors.New("invalid expression: '" + *newPeak.Value + "'")
			}
		}
	}

	p.idx = 0

	return returnNode, nil
}

func (p *Parser) parseExpr() (*NodeExpr, error) {
	peak, err := p.peak()

	if err != nil {
		return nil, err
	}

	if peak.Type == IntLiteral {
		return &NodeExpr{
			intLiteral: p.consume(),
		}, nil
	} else {
		return nil, errors.New("invalid expression")
	}
}

type Generator struct {
	root *NodeReturn
}

func (g *Generator) generate() *string {
	asm := "section .text\r\n\tglobal _start\r\n\r\n_start:\r\n"

	asm += "\tmov eax, 1\r\n"
	asm += "\tmov ebx, " + *g.root.expr.intLiteral.Value + "\r\n"
	asm += "\tint 80h"

	return &asm
}
