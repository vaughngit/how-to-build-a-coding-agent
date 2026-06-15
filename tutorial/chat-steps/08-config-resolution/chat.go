package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
)

const defaultBedrockModel = "us.anthropic.claude-opus-4-6-v1"

type bedrockConfig struct {
	Profile string `json:"aws_profile"`
	Region  string `json:"aws_region"`
	Model   string `json:"model"`
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func resolveBedrockSettings(path string) (profile, region, model string) {
	explicit := path != ""
	if path == "" {
		path = "bedrock.json"
	}
	var c bedrockConfig
	switch data, err := os.ReadFile(path); {
	case err == nil:
		if jsonErr := json.Unmarshal(data, &c); jsonErr != nil {
			log.Fatalf("failed to parse %s: %v", path, jsonErr)
		}
	case explicit:
		log.Fatalf("config file %q could not be read: %v", path, err)
	}
	profile = firstNonEmpty(os.Getenv("AWS_PROFILE"), c.Profile)
	region = firstNonEmpty(os.Getenv("AWS_REGION"), c.Region)
	model = firstNonEmpty(os.Getenv("BEDROCK_MODEL"), c.Model, defaultBedrockModel)
	return
}

func main() {
	verbose := flag.Bool("verbose", false, "enable verbose logging")
	model := flag.String("model", "", "Bedrock model / inference-profile ID (overrides config file and BEDROCK_MODEL)")
	configPath := flag.String("config", os.Getenv("BEDROCK_CONFIG"), "path to a Bedrock JSON config file (default ./bedrock.json)")
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

	profile, region, modelID := resolveBedrockSettings(*configPath)
	if *model != "" {
		modelID = *model
	}
	if *verbose {
		log.Printf("Resolved settings (profile=%q region=%q model=%s)", profile, region, modelID)
	}

	scanner := bufio.NewScanner(os.Stdin)
	getUserMessage := func() (string, bool) {
		if !scanner.Scan() {
			return "", false
		}
		return scanner.Text(), true
	}

	agent := NewAgent(getUserMessage, *verbose, modelID)
	err := agent.Run(ctx)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
}

func NewAgent(getUserMessage func() (string, bool), verbose bool, model string) *Agent {
	return &Agent{
		getUserMessage: getUserMessage,
		verbose:        verbose,
		model:          model,
	}
}

type Agent struct {
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

		assistantMessage, response, err := a.runInference(ctx, conversation)
		if err != nil {
			if a.verbose {
				log.Printf("Error during inference: %v", err)
			}
			return err
		}
		conversation = append(conversation, assistantMessage)

		fmt.Printf("\u001b[93mClaude\u001b[0m: %s\n", response)
	}

	if a.verbose {
		log.Println("Chat session ended")
	}
	return nil
}

func (a *Agent) runInference(ctx context.Context, conversation []anthropic.MessageParam) (anthropic.MessageParam, string, error) {
	_ = ctx

	if a.verbose {
		log.Printf("Pretending to call Claude with model %s, conversation length: %d", a.model, len(conversation))
	}

	response := fmt.Sprintf("stub response from %s; conversation has %d Anthropic message(s)", a.model, len(conversation))
	assistantMessage := anthropic.NewAssistantMessage(anthropic.NewTextBlock(response))
	return assistantMessage, response, nil
}
