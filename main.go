package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"unicode"
)

func main() {
	file, err := os.Open("./test.m")
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

	tokens := tokenize(string(buf))

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

func tokenize(str string) []Token {
	buf := make([]byte, 0)

	tokens := make([]Token, 0)

	i := 0

	for i < len(str) {
		c := rune(str[i])

		if unicode.IsLetter(c) {
			for unicode.IsLetter(rune(str[i])) {
				buf = append(buf, str[i])
				i++
			}

			if string(buf) == "return" {
				token := Token{Type: Return, Value: nil}
				tokens = append(tokens, token)
				buf = buf[:0]
			}
		} else if unicode.IsNumber(rune(c)) {
			for unicode.IsNumber(rune(str[i])) {
				buf = append(buf, str[i])
				i++
			}

			val := string(buf)
			token := Token{Type: IntLiteral, Value: &val}
			tokens = append(tokens, token)
			buf = buf[:0]
		} else if rune(c) == ';' {
			token := Token{Type: Semi, Value: nil}
			tokens = append(tokens, token)
			buf = buf[:0]
			i++
		} else {
			i++
		}
	}

	return tokens
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
