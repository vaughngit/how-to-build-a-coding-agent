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
	"github.com/anthropics/anthropic-sdk-go/bedrock"
	"github.com/aws/aws-sdk-go-v2/config"
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

func newBedrockClient(ctx context.Context, profile, region string) anthropic.Client {
	var opts []func(*config.LoadOptions) error
	if region != "" {
		opts = append(opts, config.WithRegion(region))
	}
	if profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}
	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}
	cfg.BearerAuthTokenProvider = nil
	return anthropic.NewClient(bedrock.WithConfig(cfg))
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
	client := newBedrockClient(ctx, profile, region)
	if *verbose {
		log.Printf("Anthropic (Bedrock) client initialized (profile=%q region=%q model=%s)", profile, region, modelID)
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
	_ = a.client

	if a.verbose {
		log.Printf("Bedrock client is wired; pretending to call Claude with model %s", a.model)
	}

	response := fmt.Sprintf("stub response from %s; the real API call is added in step 10", a.model)
	assistantMessage := anthropic.NewAssistantMessage(anthropic.NewTextBlock(response))
	return assistantMessage, response, nil
}
