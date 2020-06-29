package client

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/howeyc/gopass"
)

// inputReader provides functionality to read and parse information from stdin.
var inputReader *bufio.Reader

// Ensures that the input reader is always ready for reading from stdin.
func readyInputReader() {
	if inputReader == nil {
		inputReader = bufio.NewReader(os.Stdin)
	}
}

// Prompt the user for input on the command line and optionally supply a default
// response that is returned if the user simply hits enter without entering anything.
// Note that the prompt will be formatted with a ": " at the end of the line.
func Prompt(prompt, defaultResponse string) string {
	if defaultResponse != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultResponse)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	readyInputReader()
	text, err := inputReader.ReadString('\n')
	if err != nil {
		panic(err)
	}

	text = strings.Replace(text, "\n", "", -1)
	if text == "" {
		return defaultResponse
	}

	return text
}

// PromptPassword reads the password from the user with or without confirmation. Text
// input by the user is masked for password entry.
func PromptPassword(prompt string, confirm, allowEmpty bool) (string, error) {
	readyInputReader()
	pw := readPassword(prompt)

	if pw == "" && !allowEmpty {
		return "", errors.New("password cannot be empty")
	}

	if confirm {
		cpw := readPassword(prompt + " (confirm)")
		if pw != cpw {
			return "", errors.New("passwords do not match")
		}
	}

	return pw, nil
}

func readPassword(prompt string) string {
	fmt.Printf("%s: ", prompt)
	pass, err := gopass.GetPasswdMasked()
	if err != nil {
		panic(err)
	}
	return string(pass)
}

// Confirm the users intention with a Y/n prompt. Can optionally supply a number of
// retries to get confirmation depending on how strict the confirmation parsing is.
// True is returned if the user specifies yes, false if no.
func Confirm(prompt string, retries int, caseSensitive bool) bool {
	if retries < 0 {
		return false
	}

	readyInputReader()
	fmt.Printf("%s [Y/n]: ", prompt)
	text, err := inputReader.ReadString('\n')
	if err != nil {
		panic(err)
	}

	text = strings.Replace(text, "\n", "", -1)
	switch text[0] {
	case 'Y':
		return true
	case 'y':
		if !caseSensitive {
			return true
		}
	case 'n':
		return false
	case 'N':
		if !caseSensitive {
			return false
		}
	}

	fmt.Printf("could not understand input %q\n", text)
	return Confirm(prompt, retries-1, caseSensitive)
}
