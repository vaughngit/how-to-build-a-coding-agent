# Runnable `chat.go` Ladder

For the by-hand workbook, start with:

```text
tutorial/BUILD-CHAT-BY-HAND.md
```

That guide has you create a blank `tutorial/workbench/chat.go`, type each stage
yourself, then use these folders as the answer key.

Each folder contains a complete `chat.go`. Run any step from the repo root:

```bash
go run ./tutorial/chat-steps/03-chat-loop
```

For interactive steps, you can also pipe input:

```bash
printf 'hello\n' | go run ./tutorial/chat-steps/06-anthropic-types
```

## Steps

| Step | Adds | Still avoids |
| --- | --- | --- |
| `01-hello` | Minimum runnable Go file | stdin, structs, SDKs |
| `02-read-one-message` | `bufio.Scanner`, `os.Stdin`, one user message | loops |
| `03-chat-loop` | Repeating prompt and empty-message skip | custom types |
| `04-agent-struct` | `Agent`, constructor, method receiver | conversation history |
| `05-conversation-state` | `context.Context`, `runInference`, in-memory history | Anthropic SDK types |
| `06-anthropic-types` | `anthropic.MessageParam`, user/assistant message constructors | flags, AWS |
| `07-flags-logging` | `flag`, `log`, `-verbose` | config files |
| `08-config-resolution` | JSON config, env vars, model selection | real Bedrock client |
| `09-bedrock-client` | AWS config loading and Bedrock-backed client construction | real API call |
| `10-final-chat` | The full current root `chat.go` | nothing; requires working Bedrock access |

The first nine steps use stubbed assistant replies so they are cheap to run while you
learn the file. Step 10 is the actual Bedrock chat program.
