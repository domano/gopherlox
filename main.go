package main

import (
	"bufio"
	"io"
	"log/slog"
	"os"
	"strings"
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
	var bytes, err = os.ReadFile(path)
	if err != nil {
		slog.Error("Error while running rile: %w", err)
		return
	}
	run(string(bytes))
}

func runPrompt() {
	var reader = bufio.NewReader(os.Stdin)
	for {
		print("> ")
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			slog.Error("Error while reading input: %w", err)
			return
		}

		if err == io.EOF {
			break
		}
		run(line)

	}
}

func run(source string) {
	scanner := bufio.NewScanner(strings.NewReader(source))

	// TODO we can add the lexing step here via a split function in the scanner

	for {
		if !scanner.Scan() {
			break
		}
		slog.Info(scanner.Text())
	}
}

func errMsg(line int, msg string) {
	slog.Error("Error at line %d: %s", line, msg)
	hadError = true
}
