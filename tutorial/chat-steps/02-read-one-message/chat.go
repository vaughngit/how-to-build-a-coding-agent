package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("\u001b[94mYou\u001b[0m: ")
	if !scanner.Scan() {
		return
	}

	userInput := scanner.Text()
	fmt.Printf("\u001b[93mClaude\u001b[0m: You said %q\n", userInput)
}
