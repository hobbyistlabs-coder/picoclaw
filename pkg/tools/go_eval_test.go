package tools

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

func TestGoEvalBindings(t *testing.T) {
	tool := NewGoEvalTool("/tmp")

	// Create a mock binding
	bindings := map[string]reflect.Value{
		"TestVal": reflect.ValueOf("HelloYaegi"),
	}
	tool.SetBindings(bindings)

	// Execute a script that imports jane/env and prints the TestVal
	script := `
package main
import (
	"fmt"
	"jane/env"
)

func Run() {
	fmt.Print(env.TestVal)
}
`
	args := map[string]any{
		"code": script + "\n\nfunc init() { Run() }",
	}

	result := tool.Execute(context.Background(), args)

	if result.IsError {
		t.Fatalf("Expected successful execution, got error: %v (ForLLM: %s)", result.Err, result.ForLLM)
	}

	if !strings.Contains(result.ForLLM, "HelloYaegi") {
		t.Errorf("Expected output to contain 'HelloYaegi', got: %s", result.ForLLM)
	}
}
