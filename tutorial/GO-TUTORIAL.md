# Understanding Go through `chat.go`

A beginner's walkthrough of the Go language, using this project's `chat.go` as the
example.

The best hands-on path is `tutorial/BUILD-CHAT-BY-HAND.md`: it starts with a blank
practice file and walks you through typing each section yourself.

Alongside that workbook, `tutorial/chat-steps/` contains a ladder of complete,
runnable `chat.go` files. Each step works on its own, then the next step adds one new
idea until you arrive at the full file in this repo. After that ladder, the rest of
this document acts as a reference for the Go concepts used by the finished program.

No prior Go knowledge assumed. If you've written JavaScript, Python, or C, you'll
recognize a lot.

---

## Start here: build `chat.go` one working file at a time

If you want the by-hand learning flow, open:

```text
tutorial/BUILD-CHAT-BY-HAND.md
```

That guide tells you how to create `tutorial/workbench/chat.go` from a blank file,
type each stage yourself, run it, and compare your work against the finished snapshots.

If you want to inspect the completed examples, each folder below contains a complete
file named `chat.go`. Run a step from the repo root like this:

```bash
go run ./tutorial/chat-steps/01-hello
```

For the interactive steps, you can type into the prompt, or pipe a message in:

```bash
printf 'hello\n' | go run ./tutorial/chat-steps/06-anthropic-types
```

The ladder:

| Step | What it adds | Run it |
| --- | --- | --- |
| `01-hello` | Minimum runnable Go file | `go run ./tutorial/chat-steps/01-hello` |
| `02-read-one-message` | Read one line from stdin | `go run ./tutorial/chat-steps/02-read-one-message` |
| `03-chat-loop` | Keep prompting until input ends | `go run ./tutorial/chat-steps/03-chat-loop` |
| `04-agent-struct` | Move behavior into an `Agent` type | `go run ./tutorial/chat-steps/04-agent-struct` |
| `05-conversation-state` | Track conversation history and call `runInference` | `go run ./tutorial/chat-steps/05-conversation-state` |
| `06-anthropic-types` | Use Anthropic SDK message types, still with fake replies | `go run ./tutorial/chat-steps/06-anthropic-types` |
| `07-flags-logging` | Add `-verbose` and logging | `go run ./tutorial/chat-steps/07-flags-logging -verbose` |
| `08-config-resolution` | Add config/env/model resolution | `go run ./tutorial/chat-steps/08-config-resolution` |
| `09-bedrock-client` | Build the Bedrock-backed Anthropic client | `go run ./tutorial/chat-steps/09-bedrock-client` |
| `10-final-chat` | The full current `chat.go` with the real API call | `go run ./tutorial/chat-steps/10-final-chat` |

The first nine steps return stubbed assistant responses. That is intentional: you can
practice the Go shape without needing working AWS Bedrock credentials yet. Step 10 is
the real program, so it needs the same Bedrock access and config as the root `chat.go`.

If you like learning by copying, start with `tutorial/chat-steps/01-hello/chat.go`,
then open the next folder and compare what changed. The file stays alive the whole
time; it just grows more capable.

---

## 0. The 30-second mental model of Go

- **Compiled.** You run `go build` to turn `.go` source into a single native binary.
  `go run chat.go` just compiles-then-runs in one step.
- **Statically typed.** Every variable has a type, known at compile time. The compiler
  catches a lot of mistakes before the program ever runs.
- **Opinionated and small.** One way to format code (`gofmt`), no `while`, no
  exceptions (errors are just values you return), no classes (structs + methods instead).
- **"Unused = error."** An imported package or a local variable you don't use is a
  *compile error*, not a warning. This feels strict at first; it keeps code clean.

---

## 1. The minimum to make a Go file "work"

A runnable Go program can be a single file. The first tutorial snapshot uses this
same shape: `tutorial/chat-steps/01-hello/chat.go`.

Start with this complete baseline:

```go
// chat.go
package main

import "fmt"

func main() {
	fmt.Println("Chat with Claude")
}
```

Save that as `chat.go`, then run it from the same folder:

```bash
go run chat.go
```

Expected output:

```text
Chat with Claude
```

That file contains the three pieces a runnable Go program needs:

- **`package main`** — A package is a folder of `.go` files compiled together. The
  special name `main` means "this builds into an executable." Any other name (e.g.
  `package util`) builds into a *library* that other code imports, with no entry point.
- **`func main()`** — When you run the binary, Go calls `main()`. No `main`, no program.
- **`import`** — Pulls in code from the standard library or external modules. You can
  only use what you import, and you must use everything you import.

`chat.go` has the same three pieces, though the real file puts helper functions between
the imports and `main`. The top of the file gives us the package and imports:

```go
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
```

Then, farther down, `chat.go` supplies the entry point:

```go
func main() {
	verbose := flag.Bool("verbose", false, "enable verbose logging")
	model := flag.String("model", "", "Bedrock model / inference-profile ID (overrides config file and BEDROCK_MODEL)")
	configPath := flag.String("config", os.Getenv("BEDROCK_CONFIG"), "path to a Bedrock JSON config file (default ./bedrock.json)")
	flag.Parse()

	...
}
```

Two groups separated by a blank line, by convention: **standard library** first
(`fmt`, `os`, …), then **external** packages (the `github.com/...` ones). `gofmt`
keeps each group sorted.

### Where do the external imports come from? `go.mod`

A Go *module* is a project with a `go.mod` file at its root. Ours says:

```
module chat
go 1.24.2
require (
	github.com/anthropics/anthropic-sdk-go v1.26.0
	...
)
```

`require` lists external dependencies and their exact versions. `go mod tidy` reads
your `import` lines and updates `go.mod`/`go.sum` to match. This is Go's equivalent of
`package.json` (Node) or `requirements.txt` (Python).

> **A note about this repo specifically.** Normally all `.go` files in one folder are
> the *same* package and compile together. This workshop bends that rule: `chat.go`,
> `read.go`, `bash_tool.go`, etc. are each a *complete, standalone program* that each
> declares `package main` and its own `func main()`. That's why the `Makefile` builds
> them one at a time (`go build -o chat chat.go`) instead of all at once — two `main`
> functions in one build would collide. So treat each `.go` file here as its own mini-app.
> The tutorial ladder uses one folder per snapshot for the same reason: each folder can
> have its own complete `chat.go` without colliding with the others.

---

## 2. Functions

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

Anatomy of `func name(params) returnType { body }`:

- **`firstNonEmpty`** — the name. Lowercase first letter = **private** to this package
  (more on that in §10). Uppercase = visible to other packages.
- **`vals ...string`** — a **variadic** parameter: zero or more `string` arguments,
  received as a slice. Called like `firstNonEmpty(a, b, c)`.
- **`string`** after the `)` — the **return type**. This function hands back one string.

### Multiple return values

Go functions can return several values. This is everywhere in Go, especially the
`(result, error)` pair:

```go
cfg, err := config.LoadDefaultConfig(ctx, opts...)
```

`LoadDefaultConfig` returns *two* values; we capture both. The `opts...` "spreads" a
slice back into variadic arguments (the mirror image of `vals ...string`).

### Named return values

```go
func resolveBedrockSettings(path string) (profile, region, model string) {
	...
	profile = firstNonEmpty(...)
	region  = firstNonEmpty(...)
	model   = firstNonEmpty(...)
	return                       // "naked" return: hands back profile, region, model
}
```

Here the three returns are *named* (`profile, region, model string`). They act like
pre-declared variables; a bare `return` sends back their current values. Handy for
documenting what comes out of a function.

---

## 3. Variables, types, and constants

```go
explicit := path != ""        // short declaration: type inferred (bool here)
var c bedrockConfig           // explicit declaration; c starts as the "zero value"
const defaultBedrockModel = "us.anthropic.claude-opus-4-6-v1"
```

- **`:=`** — "declare and assign," with the type **inferred** from the right-hand side.
  Only usable *inside* a function. `explicit := path != ""` makes `explicit` a `bool`.
- **`var c bedrockConfig`** — declares `c` with an explicit type and no initial value.
  Go gives it the **zero value**: `0` for numbers, `""` for strings, `false` for bools,
  `nil` for pointers/slices/maps, and an all-zeroed struct for structs. There is no
  "undefined" in Go.
- **`const`** — a compile-time constant. `defaultBedrockModel` can never be reassigned.

Top-level (outside any function) you must use `var`, `const`, `func`, or `type` — you
cannot use `:=` there. That's why `defaultBedrockModel` uses `const`.

Common types you'll see in `chat.go`: `string`, `bool`, `int64`, `error`, plus the
composite types in the next sections.

---

## 4. Structs and methods — your "anomalous functions" explained

### A struct is a bundle of fields (like a record / a class without behavior)

```go
type Agent struct {
	client         *anthropic.Client
	getUserMessage func() (string, bool)
	verbose        bool
	model          string
}
```

`type Agent struct { ... }` defines a new type named `Agent` with four fields. Notice a
field can even be a **function** (`getUserMessage func() (string, bool)` — a function
that takes nothing and returns a string and a bool).

You can also see a simpler struct with **JSON tags**:

```go
type bedrockConfig struct {
	Profile string `json:"aws_profile"`
	Region  string `json:"aws_region"`
	Model   string `json:"model"`
}
```

The back-tick text (`` `json:"aws_profile"` ``) is a **struct tag** — metadata. It tells
`json.Unmarshal` that the JSON key `aws_profile` maps to the Go field `Profile`. That's
how `bedrock.json` turns into a Go value.

### A method is a function attached to a type

This is the "anomalous function with a star agent" you spotted:

```go
func (a *Agent) Run(ctx context.Context) error {
	...
}
```

Read it left to right:

- **`func (a *Agent)`** — the **receiver**. This makes `Run` a *method on* `Agent`.
  Inside the body, `a` refers to the specific `Agent` you called it on (like `this` or
  `self` in other languages). The `*` means the receiver is a **pointer** (see §5).
- **`Run`** — the method name.
- **`(ctx context.Context)`** — one parameter named `ctx` of type `context.Context`.
- **`error`** — the return type.

You call it with dot syntax: `agent.Run(ctx)`. `runInference` is the same idea — another
method on `Agent`:

```go
func (a *Agent) runInference(ctx context.Context, conversation []anthropic.MessageParam) (*anthropic.Message, error) {
```

So "structs hold the data, methods are the behavior." Together they're Go's stand-in for
classes — but explicitly, with no inheritance.

---

## 5. Pointers — what the `*` and `&` mean

A **pointer** is a value that holds the *address* of another value, instead of a copy.

- **`*Agent`** — "a pointer to an `Agent`."
- **`&client`** — "take the address of `client`."
- **`*p`** — "follow the pointer to the value it points at" (dereference).

In `chat.go`:

```go
client := anthropic.NewClient(...)        // client is an anthropic.Client (a value)
agent := NewAgent(&client, ...)           // pass its ADDRESS into NewAgent
```

and `NewAgent` accepts `client *anthropic.Client` — a pointer.

**Why bother?** Two reasons, both visible here:

1. **Avoid copying.** Passing a big struct by value copies the whole thing. Passing
   `*Agent` passes a small address.
2. **Allow mutation.** A method with a pointer receiver (`func (a *Agent) ...`) can
   modify the original `Agent`. A value receiver (`func (a Agent)`) would only see a
   copy. By convention, if *any* method needs a pointer receiver, they all use one — so
   `Run` and `runInference` both take `*Agent`.

### The constructor pattern: `NewAgent`

Go has no `new MyClass()` keyword for your own types. The idiom is a plain function named
`NewX` that builds and returns the struct:

```go
func NewAgent(client *anthropic.Client, getUserMessage func() (string, bool), verbose bool, model string) *Agent {
	return &Agent{
		client:         client,
		getUserMessage: getUserMessage,
		verbose:        verbose,
		model:          model,
	}
}
```

`&Agent{ ... }` creates an `Agent` and immediately takes its address, so the function
returns a `*Agent`. The `field: value` form is a **struct literal** — it sets fields by
name (order doesn't matter, and any omitted field gets its zero value).

---

## 6. Control flow

### `for` is the *only* loop — and it's also the `while`

Go deliberately has no `while` or `do/while`. `for` covers all of it:

```go
for {                 // no condition  -> infinite loop (this is the chat loop)
	...
}

for i := 0; i < n; i++ {   // classic C-style counting loop
	...
}

for cond {            // condition only -> this is your "while" loop
	...
}

for _, v := range vals {   // range loop -> iterate a slice/map/string/channel
	...
}
```

The heart of `chat.go` is the bare `for { }` — an intentional infinite loop that keeps
chatting until you hit something that stops it:

```go
for {
	fmt.Print("[94mYou[0m: ")
	userInput, ok := a.getUserMessage()
	if !ok {
		break          // jump OUT of the loop (e.g. you pressed Ctrl-D / EOF)
	}
	if userInput == "" {
		continue       // skip the rest, jump back to the top of the loop
	}
	...
}
```

- **`break`** exits the loop entirely.
- **`continue`** abandons this iteration and starts the next one.

### `if` (note: no parentheses, braces required)

```go
if userInput == "" {
	continue
}
```

Go `if` needs no `()` around the condition but always needs `{ }`. It can also run a
short statement first — you'll see this all over Go:

```go
if data, err := os.ReadFile(path); err == nil {
	// data and err only exist inside this if/else
}
```

### `switch`

```go
for _, content := range message.Content {
	switch content.Type {
	case "text":
		fmt.Printf("[93mClaude[0m: %s\n", content.Text)
	}
}
```

Cleaner than a chain of `if/else`. Go's `switch` **does not fall through** by default —
no `break` needed at the end of each `case`. (Claude's reply can contain several
"content blocks"; here we print the ones whose `Type` is `"text"`.)

There's also a clever switch in `resolveBedrockSettings` with *no* value after `switch` —
each `case` is a boolean test, which reads nicely as "first true branch wins":

```go
switch data, err := os.ReadFile(path); {
case err == nil:
	// parse it
case explicit:
	// the user named a file that we couldn't read -> hard error
}
```

---

## 7. Slices (Go's dynamic arrays)

```go
conversation := []anthropic.MessageParam{}     // empty slice of MessageParam
...
conversation = append(conversation, userMessage)   // grow it by one
...
len(conversation)                                   // its length
```

- **`[]T`** is a *slice* of `T` — a growable, ordered list (an array has a fixed size;
  you'll almost always use slices).
- **`append(s, x)`** returns a new, possibly-larger slice with `x` added. You must assign
  the result back: `s = append(s, x)`.
- **`len(s)`** gives the count.

The conversation with Claude is just a slice of messages that grows by one each turn:
your message gets appended, then Claude's reply gets appended (`message.ToParam()`), so
the model always sees the full history.

> Maps (`map[K]V`, Go's dictionaries/hash tables) don't appear in `chat.go`, but they're
> the other workhorse collection — e.g. `counts := map[string]int{}`.

---

## 8. Interfaces (and the two you already use: `error` and `context.Context`)

An **interface** is a named set of method signatures. *Any* type that has those methods
"satisfies" the interface automatically — there's no `implements` keyword. This is Go's
main form of polymorphism.

The most important interface in all of Go is `error`, which is just:

```go
type error interface {
	Error() string
}
```

Anything with an `Error() string` method *is* an `error`. That's why error handling is
"just values" (next section).

`context.Context` (the `ctx` threaded through `Run` and `runInference`) is also an
interface. A `Context` carries cancellation signals and deadlines across function and
network-call boundaries. `context.Background()` in `main` creates the root, empty one;
everything downstream receives it so a future Ctrl-C or timeout could cancel the
in-flight API call. You mostly just **pass `ctx` along** as the first argument.

---

## 9. Error handling — the `value, err :=` pattern

Go has no exceptions. Functions that can fail return an `error` as their *last* return
value. You check it immediately. This is *the* most recognizable Go idiom:

```go
message, err := a.runInference(ctx, conversation)
if err != nil {
	return err          // hand the problem up to the caller
}
// ...safe to use `message` here...
```

`err != nil` means "something went wrong." A `nil` error means success. You'll write this
little `if err != nil { ... }` block constantly — that repetition is intentional; the
error path is always visible, never hidden.

Two shortcuts in `chat.go`:

- **`_`** the blank identifier — "I must receive this value but I don't need it." Used to
  ignore a return you don't care about, e.g. `for _, v := range vals` ignores the index.
- **`log.Fatalf(...)`** — prints a message **and then exits the program** (`os.Exit(1)`).
  Used in `newBedrockClient`/`resolveBedrockSettings` for unrecoverable startup problems
  where there's no point continuing.

---

## 10. Exported vs. unexported (capitalization is the access modifier)

Go has no `public`/`private` keywords. **The first letter decides:**

- **Uppercase** = exported (visible to other packages): `Agent`, `Run`, `NewAgent`,
  `Profile`.
- **lowercase** = unexported (private to this package): `firstNonEmpty`, `runInference`,
  the `client`/`model` struct fields.

This is also why the JSON struct fields are capitalized (`Profile`, `Region`, `Model`) —
`json.Unmarshal` lives in another package and can only see *exported* fields. The struct
tag then maps them to the lowercase JSON keys.

---

## 11. Putting it together: a guided tour of `chat.go`, top to bottom

| Lines | What it is | Concept |
|------:|------------|---------|
| 1 | `package main` | makes this an executable (§1) |
| 3–15 | `import ( ... )` | standard lib + external deps (§1) |
| 17–23 | `const defaultBedrockModel` | a compile-time constant (§3) |
| 25–33 | `type bedrockConfig struct` | a struct with JSON tags (§4) |
| 35–42 | `func firstNonEmpty(vals ...string)` | a variadic helper function (§2) |
| 44–66 | `func resolveBedrockSettings(...)` | named returns + value-less `switch` (§2, §6) |
| 68–89 | `func newBedrockClient(...)` | builds the API client; `(value, err)` checks (§9) |
| 91–133 | `func main()` | entry point — wires everything up |
| 135–142 | `func NewAgent(...) *Agent` | the constructor pattern (§5) |
| 144–149 | `type Agent struct` | the data the agent carries (§4) |
| 151–213 | `func (a *Agent) Run(...)` | the chat loop — a method (§4, §6) |
| 215–235 | `func (a *Agent) runInference(...)` | one API call to Claude — a method (§4) |

### What `main()` actually does, in order

1. **Define command-line flags** (`flag.Bool`, `flag.String`) and `flag.Parse()`. After
   parsing, `*verbose`, `*model`, `*configPath` hold the values. (They're pointers —
   that's why the `*` to read them.)
2. **Set up logging** based on `-verbose`.
3. **`ctx := context.Background()`** — make the root context (§8).
4. **Resolve settings** from `bedrock.json` / env / flags, then **build the client**.
5. **Create `getUserMessage`**, a small **closure** that reads one line of stdin:

   ```go
   scanner := bufio.NewScanner(os.Stdin)
   getUserMessage := func() (string, bool) {
       if !scanner.Scan() { return "", false }
       return scanner.Text(), true
   }
   ```
   A *closure* is an anonymous function that "captures" variables from around it — here
   it captures `scanner`. It returns `(line, true)` normally, or `("", false)` at
   end-of-input. That `false` is the `ok` that makes `Run`'s loop `break`.
6. **Build the agent and run it:** `agent := NewAgent(...)`, then `agent.Run(ctx)`.

### What one turn of the loop in `Run()` does

1. Print the `You:` prompt.
2. `getUserMessage()` → your line (or stop on EOF via `!ok`).
3. Wrap your text as a message and `append` it to `conversation`.
4. `runInference` sends the *whole* conversation to Claude and returns the reply.
5. `append` Claude's reply to `conversation` (so history accumulates).
6. Loop over the reply's content blocks and print the text ones.
7. Back to step 1.

`runInference` itself is just one call — `a.client.Messages.New(ctx, anthropic.MessageNewParams{ Model, MaxTokens, Messages })` — returning `(*anthropic.Message, error)`.

---

## 12. Try it yourself (exercises that teach the syntax)

Each of these is a small, safe edit to `chat.go`. Rebuild with `go run chat.go` after each.

1. **Change a constant.** Lower `MaxTokens: int64(1024)` to `256` and watch replies get
   cut off. (Touching a literal value — §3.)
2. **Add a field to the struct.** Add `turns int` to `Agent`, set it in `NewAgent` to
   `0`, do `a.turns++` each loop iteration, and print it in the prompt:
   `fmt.Printf("[%d] You: ", a.turns)`. (Structs + pointer-receiver mutation — §4, §5.)
3. **Add a command.** Near the top of the loop, after you read `userInput`, add:
   ```go
   if userInput == "/quit" {
       break
   }
   ```
   (String comparison + `break` — §6.)
4. **Use the error return.** Make `runInference` print something when `err != nil`
   before returning it. (The error idiom — §9.)
5. **Print more block types.** Add a `case "thinking":` (or a `default:`) to the `switch`
   over `content.Type` to see what else Claude sends back. (`switch` — §6.)

When something doesn't compile, read the error — Go's messages are unusually direct
("declared and not used", "undefined: X", "cannot use Y as Z"). The compiler is your
fastest teacher here.

---

## 13. One-page cheat sheet

```go
package main                      // executable; library packages use another name
import ( "fmt" )                  // must use everything you import

const Pi = 3.14                   // compile-time constant
var x int                         // declared, zero-valued (0)
y := 10                           // declare + infer type (inside funcs only)

func add(a, b int) int { return a + b }          // params + one return
func split(s string) (a, b string) { ... return } // named multiple returns
func sum(nums ...int) int { ... }                 // variadic

type Point struct { X, Y int }    // struct (data)
func (p *Point) Move(dx int) { p.X += dx }        // method (behavior), pointer receiver
p := &Point{X: 1, Y: 2}           // struct literal + address-of

if cond { } else { }              // braces required, no parens
for { } / for cond { } / for i,v := range s { }   // the only loop keyword
switch v { case 1: ...; default: ... }            // no fallthrough by default

val, err := mightFail()           // the (value, error) pair
if err != nil { return err }      // check it immediately
_ = ignored                       // blank identifier discards a value

s := []int{}                      // slice (dynamic array)
s = append(s, 42)                 // grow it (reassign!)
m := map[string]int{}             // map (dictionary)

*T   // pointer-to-T type        &v  // address of v        *p  // value at pointer p
Name // exported (public)        name // unexported (private)
```

---

*Generated as a learning aid for this repo. The canonical, friendlier language tour is
[go.dev/tour](https://go.dev/tour/), and the deeper reference is
[go.dev/doc/effective_go](https://go.dev/doc/effective_go).*
