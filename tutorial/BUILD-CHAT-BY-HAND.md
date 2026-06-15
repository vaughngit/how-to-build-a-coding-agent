# Build `chat.go` By Hand

This is the workbook version of the tutorial.

The finished examples live in `tutorial/chat-steps/`. This guide is different: you
will create a blank `chat.go`, type each stage yourself, run it, and only use the
finished examples as an answer key when you get stuck.

The goal is not to memorize every line. The goal is to feel the file grow from a tiny
working program into the full Bedrock chat app.

## What You Need

You need:

- This repo open in a terminal.
- Go installed.
- A code editor.

Check Go:

```bash
go version
```

On this machine, Go is already installed. This repo's `go.mod` says:

```text
go 1.24.2
```

Any compatible newer Go version should work. If `go version` says the command is
missing, install Go first. On macOS, either use Homebrew:

```bash
brew install go
```

or install it from the official Go download page.

You do not need AWS or Bedrock credentials until the final step. Steps 1-9 use fake
assistant replies on purpose.

## Create Your Blank Practice File

From the repo root:

```bash
mkdir -p tutorial/workbench
touch tutorial/workbench/chat.go
```

Open `tutorial/workbench/chat.go` in your editor. It should be blank.

Run it once before adding code:

```bash
go run ./tutorial/workbench
```

It should fail because an empty Go file is not a program yet. That is fine. The first
thing Go wants from you is a `package` line.

After each step, compare your work against the completed version:

```bash
diff -u tutorial/chat-steps/01-hello/chat.go tutorial/workbench/chat.go
```

Change the step folder number as you go.

## Step 1: Make the Smallest Chat-Shaped Program

Open your blank `tutorial/workbench/chat.go` and type:

```go
package main

import "fmt"

func main() {
	fmt.Println("Chat with Claude")
}
```

Run it:

```bash
go run ./tutorial/workbench
```

Expected output:

```text
Chat with Claude
```

What you just built:

- `package main` tells Go this folder builds an executable program.
- `import "fmt"` gives you formatted printing.
- `func main()` is where the program starts.

Answer key:

```bash
diff -u tutorial/chat-steps/01-hello/chat.go tutorial/workbench/chat.go
```

Ask deeper questions here if `package main`, imports, or `func main()` still feel
mysterious.

## Step 2: Read One Message From the User

Now edit the same file. Keep `package main`, but replace the import and `main` body so
the file reads from stdin.

Your imports should become:

```go
import (
	"bufio"
	"fmt"
	"os"
)
```

Your `main` should become:

```go
func main() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("\u001b[94mYou\u001b[0m: ")
	if !scanner.Scan() {
		return
	}

	userInput := scanner.Text()
	fmt.Printf("\u001b[93mClaude\u001b[0m: You said %q\n", userInput)
}
```

Run it:

```bash
go run ./tutorial/workbench
```

Type a message and press return.

What changed:

- `bufio.NewScanner(os.Stdin)` creates a line reader.
- `scanner.Scan()` waits for one line of input.
- `scanner.Text()` gets the text the user typed.
- `%q` prints a quoted string, which makes whitespace easier to see.

Answer key:

```bash
diff -u tutorial/chat-steps/02-read-one-message/chat.go tutorial/workbench/chat.go
```

## Step 3: Turn One Message Into a Chat Loop

Now wrap the prompt/read/respond sequence in a `for` loop.

Inside `main`, after creating the scanner, add the startup message:

```go
fmt.Println("Chat with Claude (use 'ctrl-c' to quit)")
```

Then use this loop:

```go
for {
	fmt.Print("\u001b[94mYou\u001b[0m: ")
	if !scanner.Scan() {
		break
	}

	userInput := scanner.Text()
	if userInput == "" {
		continue
	}

	fmt.Printf("\u001b[93mClaude\u001b[0m: You said %q\n", userInput)
}
```

Run it:

```bash
go run ./tutorial/workbench
```

Try:

- Typing a normal message.
- Pressing return on an empty line.
- Pressing `ctrl-c` to quit.

What changed:

- `for { ... }` is Go's endless loop.
- `break` exits the loop.
- `continue` skips to the next loop turn.

Answer key:

```bash
diff -u tutorial/chat-steps/03-chat-loop/chat.go tutorial/workbench/chat.go
```

## Step 4: Move the Chat Behavior Into an `Agent`

Now the file starts to look like the real architecture.

The job of `main` becomes setup:

- Create the scanner.
- Create a `getUserMessage` function.
- Create an `Agent`.
- Call `agent.Run()`.

Replace `main` with:

```go
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
```

Then add these below `main`:

```go
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
```

Run it:

```bash
go run ./tutorial/workbench
```

The behavior should feel the same as step 3. The structure is different.

What changed:

- `type Agent struct` creates a bundle of fields.
- `NewAgent` is a constructor function.
- `func (a *Agent) Run()` is a method attached to `Agent`.
- `getUserMessage` is a function stored inside the struct.

Answer key:

```bash
diff -u tutorial/chat-steps/04-agent-struct/chat.go tutorial/workbench/chat.go
```

This is a good place to ask about structs, methods, pointers, or why `main` got
smaller.

## Step 5: Add Context, Conversation State, and `runInference`

Now add `context` to your imports:

```go
import (
	"bufio"
	"context"
	"fmt"
	"os"
)
```

In `main`, create a context before creating the agent:

```go
ctx := context.Background()
```

Then call:

```go
err := agent.Run(ctx)
```

Change the `Run` method signature:

```go
func (a *Agent) Run(ctx context.Context) error {
```

At the top of `Run`, create conversation history:

```go
conversation := []string{}
```

When the user enters text, append it:

```go
conversation = append(conversation, "user: "+userInput)
```

Then call a new method:

```go
response, err := a.runInference(ctx, conversation)
if err != nil {
	return err
}
conversation = append(conversation, "assistant: "+response)

fmt.Printf("\u001b[93mClaude\u001b[0m: %s\n", response)
```

Finally, add `runInference` below `Run`:

```go
func (a *Agent) runInference(ctx context.Context, conversation []string) (string, error) {
	_ = ctx

	return fmt.Sprintf("stub response; conversation has %d message(s)", len(conversation)), nil
}
```

Run it:

```bash
go run ./tutorial/workbench
```

What changed:

- `context.Context` is passed through now, even though the fake function does not use it.
- `[]string{}` creates an empty slice.
- `append` adds conversation turns.
- `runInference` becomes the place where the future API call will live.
- `_ = ctx` tells Go, "I know this value is unused for now."

Answer key:

```bash
diff -u tutorial/chat-steps/05-conversation-state/chat.go tutorial/workbench/chat.go
```

## Step 6: Swap Plain Strings for Anthropic Message Types

Now you start using the same message types as the real app.

Add the Anthropic SDK import below the standard library imports:

```go
import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
)
```

Change the conversation slice:

```go
conversation := []anthropic.MessageParam{}
```

Replace the string append for user messages with:

```go
userMessage := anthropic.NewUserMessage(anthropic.NewTextBlock(userInput))
conversation = append(conversation, userMessage)
```

Change the inference call to receive an assistant message and printable text:

```go
assistantMessage, response, err := a.runInference(ctx, conversation)
if err != nil {
	return err
}
conversation = append(conversation, assistantMessage)

fmt.Printf("\u001b[93mClaude\u001b[0m: %s\n", response)
```

Change `runInference` to:

```go
func (a *Agent) runInference(ctx context.Context, conversation []anthropic.MessageParam) (anthropic.MessageParam, string, error) {
	_ = ctx

	response := fmt.Sprintf("stub response; conversation has %d Anthropic message(s)", len(conversation))
	assistantMessage := anthropic.NewAssistantMessage(anthropic.NewTextBlock(response))
	return assistantMessage, response, nil
}
```

Run it:

```bash
go run ./tutorial/workbench
```

What changed:

- The app now stores conversation history in the shape the Anthropic API expects.
- The reply is still fake, but it is fake data in the right type.
- This step creates a bridge from "learning Go" to "building the real chat app."

Answer key:

```bash
diff -u tutorial/chat-steps/06-anthropic-types/chat.go tutorial/workbench/chat.go
```

## Step 7: Add Flags and Verbose Logging

Add `flag` and `log` to the imports:

```go
import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
)
```

At the top of `main`, before creating the scanner, add:

```go
verbose := flag.Bool("verbose", false, "enable verbose logging")
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
```

Pass `*verbose` into the agent:

```go
agent := NewAgent(getUserMessage, *verbose)
```

Update the constructor and struct:

```go
func NewAgent(getUserMessage func() (string, bool), verbose bool) *Agent {
	return &Agent{
		getUserMessage: getUserMessage,
		verbose:        verbose,
	}
}

type Agent struct {
	getUserMessage func() (string, bool)
	verbose        bool
}
```

Add logs in `Run`:

```go
if a.verbose {
	log.Println("Starting chat session")
}
```

Inside the loop, log input, skipped empty messages, inference errors, and the end of
the session. Use the answer key if you want the exact placements.

Add this inside `runInference`:

```go
if a.verbose {
	log.Printf("Pretending to call Claude, conversation length: %d", len(conversation))
}
```

Run it:

```bash
go run ./tutorial/workbench -verbose
```

What changed:

- `flag.Bool` returns a pointer, so you read it with `*verbose`.
- `log` is for diagnostic output.
- The app now has a quiet mode and a verbose mode.

Answer key:

```bash
diff -u tutorial/chat-steps/07-flags-logging/chat.go tutorial/workbench/chat.go
```

## Step 8: Add Config, Env Vars, and Model Selection

This is the largest pure-Go step. You are adding the code that decides which AWS
profile, region, and Bedrock model to use.

Add imports:

```go
	"encoding/json"
```

Near the top of the file, after imports, add:

```go
const defaultBedrockModel = "us.anthropic.claude-opus-4-6-v1"

type bedrockConfig struct {
	Profile string `json:"aws_profile"`
	Region  string `json:"aws_region"`
	Model   string `json:"model"`
}
```

Add this helper:

```go
func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
```

Add this resolver:

```go
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
```

In `main`, add flags:

```go
model := flag.String("model", "", "Bedrock model / inference-profile ID (overrides config file and BEDROCK_MODEL)")
configPath := flag.String("config", os.Getenv("BEDROCK_CONFIG"), "path to a Bedrock JSON config file (default ./bedrock.json)")
```

After creating `ctx`, resolve settings:

```go
profile, region, modelID := resolveBedrockSettings(*configPath)
if *model != "" {
	modelID = *model
}
if *verbose {
	log.Printf("Resolved settings (profile=%q region=%q model=%s)", profile, region, modelID)
}
```

Pass `modelID` into the agent:

```go
agent := NewAgent(getUserMessage, *verbose, modelID)
```

Add `model string` to the constructor and struct. Then update the fake response to
include the model:

```go
response := fmt.Sprintf("stub response from %s; conversation has %d Anthropic message(s)", a.model, len(conversation))
```

Run it:

```bash
go run ./tutorial/workbench -verbose
```

What changed:

- Struct tags map JSON keys to Go field names.
- Env vars override config-file values.
- The `-model` flag overrides both.
- The program now has the same settings flow as the final app, but still no real API call.

Answer key:

```bash
diff -u tutorial/chat-steps/08-config-resolution/chat.go tutorial/workbench/chat.go
```

## Step 9: Build the Bedrock Client, But Keep Fake Replies

Now add the AWS and Bedrock imports:

```go
	"github.com/anthropics/anthropic-sdk-go/bedrock"
	"github.com/aws/aws-sdk-go-v2/config"
```

Add this function after `resolveBedrockSettings`:

```go
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
```

In `main`, create the client after resolving settings:

```go
client := newBedrockClient(ctx, profile, region)
if *verbose {
	log.Printf("Anthropic (Bedrock) client initialized (profile=%q region=%q model=%s)", profile, region, modelID)
}
```

Pass the client into the agent:

```go
agent := NewAgent(&client, getUserMessage, *verbose, modelID)
```

Update the constructor and struct to store it:

```go
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
```

Inside the fake `runInference`, add:

```go
_ = a.client
```

Run it:

```bash
go run ./tutorial/workbench -verbose
```

What changed:

- The real Bedrock-backed client now exists.
- The app still does not call the model, so this remains a lower-risk checkpoint.
- If this step fails, the issue is likely AWS config/profile/region setup, not chat logic.

Answer key:

```bash
diff -u tutorial/chat-steps/09-bedrock-client/chat.go tutorial/workbench/chat.go
```

## Step 10: Replace the Fake Inference With the Real API Call

This is the point where the program becomes the real `chat.go`.

Change `Run` so it calls:

```go
message, err := a.runInference(ctx, conversation)
if err != nil {
	if a.verbose {
		log.Printf("Error during inference: %v", err)
	}
	return err
}
conversation = append(conversation, message.ToParam())
```

Then print each text content block:

```go
for _, content := range message.Content {
	switch content.Type {
	case "text":
		fmt.Printf("\u001b[93mClaude\u001b[0m: %s\n", content.Text)
	}
}
```

Replace `runInference` with:

```go
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
```

Run it:

```bash
go run ./tutorial/workbench -verbose
```

This step requires working Bedrock access. If it fails, keep the error. It is useful.
Good questions at this point are usually about credentials, region, model IDs, or the
shape of `anthropic.MessageNewParams`.

Final answer key:

```bash
diff -u tutorial/chat-steps/10-final-chat/chat.go tutorial/workbench/chat.go
```

If there is no diff, you built the same file by hand.

## How To Use This With Me

After each step, bring me one of these:

- The command you ran.
- The error message.
- The part that felt weird.
- A diff against the answer key.

Good prompts:

- "Explain why this import is needed."
- "Why does `flag.Bool` give me a pointer?"
- "Why is `Run` a method but `NewAgent` is a function?"
- "What is `context.Context` doing here?"
- "Why does the fake inference step come before the real API call?"
- "Read my diff and tell me what I misunderstood."

Do not rush the final Bedrock call. The middle steps are where the Go model starts to
stick.
