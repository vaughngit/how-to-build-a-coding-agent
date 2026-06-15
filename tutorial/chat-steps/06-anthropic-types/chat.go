package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
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
	conversation := []anthropic.MessageParam{}

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

		userMessage := anthropic.NewUserMessage(anthropic.NewTextBlock(userInput))
		conversation = append(conversation, userMessage)

		assistantMessage, response, err := a.runInference(ctx, conversation)
		if err != nil {
			return err
		}
		conversation = append(conversation, assistantMessage)

		fmt.Printf("\u001b[93mClaude\u001b[0m: %s\n", response)
	}

	return nil
}

func (a *Agent) runInference(ctx context.Context, conversation []anthropic.MessageParam) (anthropic.MessageParam, string, error) {
	_ = ctx

	response := fmt.Sprintf("stub response; conversation has %d Anthropic message(s)", len(conversation))
	assistantMessage := anthropic.NewAssistantMessage(anthropic.NewTextBlock(response))
	return assistantMessage, response, nil
}
