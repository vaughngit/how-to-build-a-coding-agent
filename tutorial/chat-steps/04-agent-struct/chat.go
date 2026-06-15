package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	getUserMessage := func() (string, bool) {
		if !scanner.Scan() {
			return "", false
		}
		return scanner.Text(), true
	}

	agent := NewAgent(getUserMessage)
	err := agent.Run()
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
}

func NewAgent(getUserMessage func() (string, bool)) *Agent {
	return &Agent{
		getUserMessage: getUserMessage,
	}
}

type Agent struct {
	getUserMessage func() (string, bool)
}

func (a *Agent) Run() error {
	fmt.Println("Chat with Claude (use 'ctrl-c' to quit)")

	for {
		fmt.Print("\u001b[94mYou\u001b[0m: ")
		userInput, ok := a.getUserMessage()
		if !ok {
			break
		}

		if userInput == "" {
			continue
		}

		fmt.Printf("\u001b[93mClaude\u001b[0m: You said %q\n", userInput)
	}

	return nil
}
