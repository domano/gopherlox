package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
)

var hadError bool

func main() {
	if len(os.Args) > 2 {
		slog.Info("Usage: gopherlox [script]")
		os.Exit(64)
	}
	if len(os.Args) == 2 {
		runFile(os.Args[1])
	}
	runPrompt()
}

func runFile(path string) {
	var file, err = os.Open(path)
	if err != nil {
		slog.Error(fmt.Errorf("error while running file: %w", err).Error())
		return
	}
	run(file)
}

func runPrompt() {
	var reader = bufio.NewReader(os.Stdin)
	for {
		run(reader)
	}
}

var isReadingLiteral bool
var currentLiteral []byte

func run(source io.Reader) {
	scanner := bufio.NewScanner(source)
	var start, line int
	print("> ")
	for scanner.Scan() {
		start = 0
		lineString := scanner.Bytes()
		for current := 0; current < len(lineString); current++ {
			if current >= len(lineString) {
				errMsg(line, "Unexpected end of file")
				return
			}
			if isReadingLiteral {
				advance, err := charsUntilStringEnd(lineString[start:])
				if errors.Is(err, ErrStringLineBreak) {
					currentLiteral = append(currentLiteral, lineString[start:]...)
					break
				}
				if err != nil {
					errMsg(line, err.Error())
					return
				}
				currentLiteral = append(currentLiteral, lineString[start:start+advance-1]...)
				Tokens = append(Tokens, Token{Type: STRING, Lexeme: fmt.Sprintf("\"%s\"", currentLiteral), Literal: string(currentLiteral), Line: line})
				start += advance
				current += advance
				isReadingLiteral = false
			}
			token, n, err := scanToken(lineString, start, current+1, line)
			if errors.Is(err, ErrComment) {
				break
			}
			if errors.Is(err, ErrUnkownToken) {
				continue
			}
			if errors.Is(err, ErrWhiteSpace) {
				start++
				continue
			}
			if errors.Is(err, ErrString) {
				start++
				currentLiteral = make([]byte, 0)
				isReadingLiteral = true
				advance, err := charsUntilStringEnd(lineString[start:])
				if errors.Is(err, ErrStringLineBreak) {
					currentLiteral = append(currentLiteral, lineString[start:]...)
					break
				}
				if err != nil {
					errMsg(line, err.Error())
					return
				}
				currentLiteral = append(currentLiteral, lineString[start:start+advance-1]...) // dont include the " at the end
				Tokens = append(Tokens, Token{Type: STRING, Lexeme: fmt.Sprintf("\"%s\"", currentLiteral), Literal: string(currentLiteral), Line: line})
				start += advance
				current += advance
				isReadingLiteral = false
				continue
			}
			if errors.Is(err, ErrNumber) {
				currentLiteral = make([]byte, 0)
				advance, err := charsUntilNumberEnd(lineString[start:])
				if err != nil {
					errMsg(line, err.Error())
					return
				}
				currentLiteral = append(currentLiteral, lineString[start:start+advance]...)
				Tokens = append(Tokens, Token{Type: NUMBER, Lexeme: string(currentLiteral), Literal: string(currentLiteral), Line: line})
				start += advance
				current += advance
				continue
			}
			Tokens = append(Tokens, token)
			start += n
			if start >= len(lineString) {
				break
			}
		}
		line++
		for _, token := range Tokens {
			slog.Info(token.String())
		}
		print("> ")
	}

}

func errMsg(line int, msg string) {
	slog.Error(msg, "line", line)
	hadError = true
}

var (
	ErrUnkownToken = errors.New("unknown token")
	ErrComment     = errors.New("comment")
	ErrWhiteSpace  = errors.New("white space")
	ErrString      = errors.New("string start")
	ErrNumber      = errors.New("number")
)

func scanToken(data []byte, start, end, line int) (Token, int, error) {
	switch string(data[start:end]) {
	case "(":
		return Token{Type: LEFT_PAREN, Lexeme: "(", Literal: "(", Line: line}, end - start, nil
	case ")":
		return Token{Type: RIGHT_PAREN, Lexeme: ")", Literal: ")", Line: line}, end - start, nil
	case "{":
		return Token{Type: LEFT_BRACE, Lexeme: "{", Literal: "{", Line: line}, end - start, nil
	case "}":
		return Token{Type: RIGHT_BRACE, Lexeme: "}", Literal: "}", Line: line}, end - start, nil
	case ",":
		return Token{Type: COMMA, Lexeme: ",", Literal: ",", Line: line}, end - start, nil
	case ".":
		return Token{Type: DOT, Lexeme: ".", Literal: ".", Line: line}, end - start, nil
	case "-":
		return Token{Type: MINUS, Lexeme: "-", Literal: "-", Line: line}, end - start, nil
	case "+":
		return Token{Type: PLUS, Lexeme: "+", Literal: "+", Line: line}, end - start, nil
	case ";":
		return Token{Type: SEMICOLON, Lexeme: ";", Literal: ";", Line: line}, end - start, nil
	case "/":
		token, n, err := scanToken(data, start, end+1, line)
		if !errors.Is(err, ErrUnkownToken) { // could be a comment
			return token, n, err
		}
		return Token{Type: SLASH, Lexeme: "/", Literal: "/", Line: line}, end - start, nil
	case "//":
		return Token{}, 0, ErrComment
	case "*":
		return Token{Type: STAR, Lexeme: "*", Literal: "*", Line: line}, end - start, nil
	case "!":
		token, n, err := scanToken(data, start, end+1, line)
		if !errors.Is(err, ErrUnkownToken) {
			return token, n, err
		}
		return Token{Type: BANG, Lexeme: "!", Literal: "!", Line: line}, end - start, nil
	case "=":
		token, n, err := scanToken(data, start, end+1, line)
		if !errors.Is(err, ErrUnkownToken) {
			return token, n, err
		}
		return Token{Type: EQUAL, Lexeme: "=", Literal: "=", Line: line}, end - start, nil
	case "!=":
		return Token{Type: BANG_EQUAL, Lexeme: "!=", Literal: "!=", Line: line}, end - start, nil
	case "==":
		return Token{Type: EQUAL_EQUAL, Lexeme: "==", Literal: "==", Line: line}, end - start, nil
	case ">":
		token, n, err := scanToken(data, start, end+1, line)
		if !errors.Is(err, ErrUnkownToken) {
			return token, n, err
		}
		return Token{Type: GREATER, Lexeme: ">", Literal: ">", Line: line}, end - start, nil
	case ">=":
		return Token{Type: GREATER_EQUAL, Lexeme: ">=", Literal: ">=", Line: line}, end - start, nil
	case "<":
		token, n, err := scanToken(data, start, end+1, line)
		if !errors.Is(err, ErrUnkownToken) {
			return token, n, err
		}
		return Token{Type: LESS, Lexeme: "<", Literal: "<", Line: line}, end - start, nil
	case "<=":
		return Token{Type: LESS_EQUAL, Lexeme: "<=", Literal: "<=", Line: line}, end - start, nil
	case "var":
		return Token{Type: VAR, Lexeme: "var", Literal: "var", Line: line}, end - start, nil
	case "class":
		return Token{Type: CLASS, Lexeme: "class", Literal: "class", Line: line}, end - start, nil
	case "super":
		return Token{Type: SUPER, Lexeme: "super", Literal: "super", Line: line}, end - start, nil
	case "this":
		return Token{Type: THIS, Lexeme: "this", Literal: "this", Line: line}, end - start, nil
	case "true":
		return Token{Type: TRUE, Lexeme: "true", Literal: "true", Line: line}, end - start, nil
	case "false":
		return Token{Type: FALSE, Lexeme: "false", Literal: "false", Line: line}, end - start, nil
	case "nil":
		return Token{Type: NIL, Lexeme: "nil", Literal: "nil", Line: line}, end - start, nil
	case "if":
		return Token{Type: IF, Lexeme: "if", Literal: "if", Line: line}, end - start, nil
	case "else":
		return Token{Type: ELSE, Lexeme: "else", Literal: "else", Line: line}, end - start, nil
	case "while":
		return Token{Type: WHILE, Lexeme: "while", Literal: "while", Line: line}, end - start, nil
	case "for":
		return Token{Type: FOR, Lexeme: "for", Literal: "for", Line: line}, end - start, nil
	case "fun":
		return Token{Type: FUN, Lexeme: "fun", Literal: "fun", Line: line}, end - start, nil
	case "return":
		return Token{Type: RETURN, Lexeme: "return", Literal: "return", Line: line}, end - start, nil
	case "print":
		return Token{Type: PRINT, Lexeme: "print", Literal: "print", Line: line}, end - start, nil
	case "\"":
		return Token{}, 0, ErrString
	case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
		return Token{}, 0, ErrNumber
	case " ", "\r", "\t":
		return Token{}, 0, ErrWhiteSpace
	default:
		errMsg(line, fmt.Sprintf("Unexpected char '%s'", string(data)))
		return Token{}, 0, ErrUnkownToken
	}
}

var ErrStringLineBreak = errors.New("string line break")

func charsUntilStringEnd(data []byte) (int, error) {
	for current := range data {
		if data[current] == '"' {
			return current + 1, nil
		}
	}
	return 0, ErrStringLineBreak
}

func charsUntilNumberEnd(data []byte) (int, error) {
	var dot bool
	for current := range data {
		if data[current] == '.' {
			if dot {
				return current, nil
			}
			dot = true
			continue
		}
		if data[current] <= '0' && data[current] >= '9' {
			return current, nil
		}
		if data[current] == ' ' || data[current] == '\t' || data[current] == '\r' {
			return current, nil
		}
	}
	return len(data), nil
}
