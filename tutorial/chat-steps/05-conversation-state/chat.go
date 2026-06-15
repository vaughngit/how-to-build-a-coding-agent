package main

import (
	"bufio"
	"context"
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

	ctx := context.Background()

	agent := NewAgent(getUserMessage)
	err := agent.Run(ctx)
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

func (a *Agent) Run(ctx context.Context) error {
	conversation := []string{}

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

		conversation = append(conversation, "user: "+userInput)

		response, err := a.runInference(ctx, conversation)
		if err != nil {
			return err
		}
		conversation = append(conversation, "assistant: "+response)

		fmt.Printf("\u001b[93mClaude\u001b[0m: %s\n", response)
	}

	return nil
}

func (a *Agent) runInference(ctx context.Context, conversation []string) (string, error) {
	_ = ctx

	return fmt.Sprintf("stub response; conversation has %d message(s)", len(conversation)), nil
}
