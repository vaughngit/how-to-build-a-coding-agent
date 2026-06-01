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

// defaultBedrockModel is a known-real us-east-1 inference-profile ID used when
// neither -model nor BEDROCK_MODEL is set. Override it to point at the exact
// Claude model you've enabled (e.g. Opus 4.6). List available IDs with:
//
//	aws bedrock list-inference-profiles --region us-east-1 \
//	  --query "inferenceProfileSummaries[].inferenceProfileId"
const defaultBedrockModel = "us.anthropic.claude-opus-4-6-v1"

// bedrockConfig holds optional settings loaded from a JSON config file so you
// don't have to export environment variables. The default path is ./bedrock.json
// (override with -config or the BEDROCK_CONFIG env var). Matching env vars
// (AWS_PROFILE, AWS_REGION, BEDROCK_MODEL) still take precedence when set.
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

// resolveBedrockSettings merges the optional config file with environment
// variables (env wins) and returns the AWS profile, region, and model to use.
func resolveBedrockSettings(path string) (profile, region, model string) {
	if path == "" {
		path = "bedrock.json"
	}
	var c bedrockConfig
	if data, err := os.ReadFile(path); err == nil {
		if jsonErr := json.Unmarshal(data, &c); jsonErr != nil {
			log.Fatalf("failed to parse %s: %v", path, jsonErr)
		}
	}
	profile = firstNonEmpty(os.Getenv("AWS_PROFILE"), c.Profile)
	region = firstNonEmpty(os.Getenv("AWS_REGION"), c.Region)
	model = firstNonEmpty(os.Getenv("BEDROCK_MODEL"), c.Model, defaultBedrockModel)
	return
}

// newBedrockClient builds an Anthropic client routed through Amazon Bedrock,
// using the resolved profile/region and the standard AWS credential chain
// (SSO profiles, static access keys, or instance roles). No ANTHROPIC_API_KEY.
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
	// With AWS SSO, LoadDefaultConfig sets a bearer-token provider for the SSO
	// OIDC token; the Bedrock helper would prefer it over SigV4 and send it as a
	// Bedrock API key (403 "Invalid API Key format"). Clear it to force SigV4.
	// A real Bedrock API key in AWS_BEARER_TOKEN_BEDROCK is still honored.
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

	// Resolve AWS profile / region / model from an optional config file so they
	// don't have to be exported; matching environment variables still win.
	profile, region, modelID := resolveBedrockSettings(*configPath)
	if *model != "" {
		modelID = *model // -model flag overrides config file and env
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
