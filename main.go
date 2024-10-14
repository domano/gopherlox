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
		for current := 1; current <= len(lineString); current++ {
			if current > len(lineString) {
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
				currentLiteral = append(currentLiteral, lineString[start:start+advance]...)
				Tokens = append(Tokens, Token{Type: STRING, Lexeme: fmt.Sprintf("\"%s\"", currentLiteral), Literal: string(currentLiteral), Line: line})
				start += advance
				current += advance
				isReadingLiteral = false
			}
			token, n, err := scanToken(lineString, start, current, line)
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
				currentLiteral = append(currentLiteral, lineString[start:start+advance]...) // dont include the " at the end
				Tokens = append(Tokens, Token{Type: STRING, Lexeme: fmt.Sprintf("\"%s\"", currentLiteral), Literal: string(currentLiteral), Line: line})
				start += advance
				current = start
				isReadingLiteral = false
				continue
			}
			if errors.Is(err, ErrIdentifier) {
				currentLiteral = make([]byte, 0)
				advance, err := charsUntilStringEnd(lineString[start:])
				if err != nil {
					errMsg(line, err.Error())
					return
				}
				currentLiteral = append(currentLiteral, lineString[start:start+advance]...)
				Tokens = append(Tokens, identifier(currentLiteral, line))
				start += advance
				current = start
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
				current = start
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
	ErrIdentifier  = errors.New("identifier")
)

func scanToken(data []byte, start, end, line int) (Token, int, error) {
	var str string = string(data[start:end])
	switch str {
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
	case "\"":
		return Token{}, 0, ErrString
	case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
		return Token{}, 0, ErrNumber
	case " ", "\r", "\t":
		return Token{}, 0, ErrWhiteSpace
	default:
		if isAlpha(data[start]) {
			return Token{}, 0, ErrIdentifier
		}
		errMsg(line, fmt.Sprintf("Unexpected char '%s'", string(data)))
		return Token{}, 0, ErrUnkownToken
	}
}

func identifier(data []byte, line int) Token {
	switch string(data) {
	case "var":
		return Token{Type: VAR, Lexeme: "var", Literal: "var", Line: line}
	case "class":
		return Token{Type: CLASS, Lexeme: "class", Literal: "class", Line: line}
	case "super":
		return Token{Type: SUPER, Lexeme: "super", Literal: "super", Line: line}
	case "this":
		return Token{Type: THIS, Lexeme: "this", Literal: "this", Line: line}
	case "true":
		return Token{Type: TRUE, Lexeme: "true", Literal: "true", Line: line}
	case "false":
		return Token{Type: FALSE, Lexeme: "false", Literal: "false", Line: line}
	case "nil":
		return Token{Type: NIL, Lexeme: "nil", Literal: "nil", Line: line}
	case "if":
		return Token{Type: IF, Lexeme: "if", Literal: "if", Line: line}
	case "else":
		return Token{Type: ELSE, Lexeme: "else", Literal: "else", Line: line}
	case "while":
		return Token{Type: WHILE, Lexeme: "while", Literal: "while", Line: line}
	case "for":
		return Token{Type: FOR, Lexeme: "for", Literal: "for", Line: line}
	case "fun":
		return Token{Type: FUN, Lexeme: "fun", Literal: "fun", Line: line}
	case "return":
		return Token{Type: RETURN, Lexeme: "return", Literal: "return", Line: line}
	case "print":
		return Token{Type: PRINT, Lexeme: "print", Literal: "print", Line: line}
	default:
		return Token{Type: IDENTIFIER, Lexeme: string(data), Literal: string(data), Line: line}
	}
}

func isAlphaNumeric(data byte) bool {
	return (data >= 'a' && data <= 'z') || (data >= 'A' && data <= 'Z') || (data >= '0' && data <= '9') || data == '_'
}
func isAlpha(data byte) bool {
	return (data >= 'a' && data <= 'z') || (data >= 'A' && data <= 'Z') || data == '_'
}

func isDigit(data byte) bool {
	return data >= '0' && data <= '9'
}

var ErrStringLineBreak = errors.New("string line break")

func charsUntilStringEnd(data []byte) (int, error) {
	for current := range data {
		if !isAlphaNumeric(data[current]) {
			return current, nil
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
		if !isDigit(data[current]) {
			return current, nil
		}
	}
	return len(data), nil
}
