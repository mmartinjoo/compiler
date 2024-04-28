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

	asm := tokensToAsm(tokens)

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

func tokensToAsm(tokens []Token) string {
	asm := "section .text\r\n\tglobal _start\r\n\r\n_start:\r\n"

	for i, token := range tokens {
		if token.Type == Return {
			if i+1 < len(tokens) && tokens[i+1].Type == IntLiteral {
				returnVal := tokens[i+1].Value
				if i+2 < len(tokens) && tokens[i+2].Type == Semi {
					asm += "\tmov eax, 1\r\n\tmov ebx, " + *returnVal + "\r\n\tint 80h"
				}
			}
		}
	}

	return asm
}

func write(str string) {
	file, err := os.Create("./out.asm")
	defer file.Close()

	if err != nil {
		panic(err)
	}

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	_, err = writer.WriteString(str)

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
