package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/bedrock"
	"github.com/aws/aws-sdk-go-v2/config"
)

// defaultBedrockModel is a known-real us-east-1 inference-profile ID used when
// neither -model nor BEDROCK_MODEL is set. Override it to point at the exact
// Claude model you've enabled (e.g. Opus 4.6). List available IDs with:
//
//	aws bedrock list-inference-profiles --region us-east-1 \
//	  --query "inferenceProfileSummaries[].inferenceProfileId"
const defaultBedrockModel = "us.anthropic.claude-opus-4-6-v1"

func main() {
	verbose := flag.Bool("verbose", false, "enable verbose logging")
	model := flag.String("model", os.Getenv("BEDROCK_MODEL"), "Bedrock model / inference-profile ID")
	flag.Parse()

	if *verbose {
		log.SetOutput(os.Stderr)
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Println("Verbose logging enabled")
	} else {
		log.SetOutput(os.Stdout)
		log.SetFlags(0)
		log.SetPrefix("")
	}

	ctx := context.Background()

	// Route through Amazon Bedrock using the standard AWS credential chain
	// (AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY / AWS_REGION, profiles, SSO,
	// or instance roles). No ANTHROPIC_API_KEY is used.
	cfg, cfgErr := config.LoadDefaultConfig(ctx)
	if cfgErr != nil {
		log.Fatalf("failed to load AWS config: %v", cfgErr)
	}
	// With AWS SSO, LoadDefaultConfig sets a bearer-token provider for the SSO
	// OIDC token; the Bedrock helper would prefer it over SigV4 and send it as a
	// Bedrock API key (403 "Invalid API Key format"). Clear it to force SigV4.
	// A real Bedrock API key in AWS_BEARER_TOKEN_BEDROCK is still honored.
	cfg.BearerAuthTokenProvider = nil
	client := anthropic.NewClient(bedrock.WithConfig(cfg))
	if *verbose {
		log.Println("Anthropic (Bedrock) client initialized")
	}

	modelID := *model
	if modelID == "" {
		modelID = defaultBedrockModel
	}
	if *verbose {
		log.Printf("Using Bedrock model: %s", modelID)
	}

	scanner := bufio.NewScanner(os.Stdin)
	getUserMessage := func() (string, bool) {
		if !scanner.Scan() {
			return "", false
		}
		return scanner.Text(), true
	}

	agent := NewAgent(&client, getUserMessage, *verbose, modelID)
	err := agent.Run(ctx)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
}

func NewAgent(client *anthropic.Client, getUserMessage func() (string, bool), verbose bool, model string) *Agent {
	return &Agent{
		client:         client,
		getUserMessage: getUserMessage,
		verbose:        verbose,
		model:          model,
	}
}

type Agent struct {
	client         *anthropic.Client
	getUserMessage func() (string, bool)
	verbose        bool
	model          string
}

func (a *Agent) Run(ctx context.Context) error {
	conversation := []anthropic.MessageParam{}

	if a.verbose {
		log.Println("Starting chat session")
	}
	fmt.Println("Chat with Claude (use 'ctrl-c' to quit)")

	for {
		fmt.Print("\u001b[94mYou\u001b[0m: ")
		userInput, ok := a.getUserMessage()
		if !ok {
			if a.verbose {
				log.Println("User input ended, breaking from chat loop")
			}
			break
		}

		// Skip empty messages
		if userInput == "" {
			if a.verbose {
				log.Println("Skipping empty message")
			}
			continue
		}

		if a.verbose {
			log.Printf("User input received: %q", userInput)
		}

		userMessage := anthropic.NewUserMessage(anthropic.NewTextBlock(userInput))
		conversation = append(conversation, userMessage)

		if a.verbose {
			log.Printf("Sending message to Claude, conversation length: %d", len(conversation))
		}

		message, err := a.runInference(ctx, conversation)
		if err != nil {
			if a.verbose {
				log.Printf("Error during inference: %v", err)
			}
			return err
		}
		conversation = append(conversation, message.ToParam())

		if a.verbose {
			log.Printf("Received response from Claude with %d content blocks", len(message.Content))
		}

		for _, content := range message.Content {
			switch content.Type {
			case "text":
				fmt.Printf("\u001b[93mClaude\u001b[0m: %s\n", content.Text)
			}
		}
	}

	if a.verbose {
		log.Println("Chat session ended")
	}
	return nil
}

func (a *Agent) runInference(ctx context.Context, conversation []anthropic.MessageParam) (*anthropic.Message, error) {
	if a.verbose {
		log.Printf("Making API call to Claude with model: %s", a.model)
	}

	message, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(a.model),
		MaxTokens: int64(1024),
		Messages:  conversation,
	})

	if a.verbose {
		if err != nil {
			log.Printf("API call failed: %v", err)
		} else {
			log.Printf("API call successful, response received")
		}
	}

	return message, err
}
