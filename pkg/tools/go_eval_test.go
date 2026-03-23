package tools

import (
	"context"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGoEvalTool_BasicEval(t *testing.T) {
	evalTool := NewGoEvalTool("/tmp")

	args := map[string]any{
		"code": `
package main

import "fmt"

func main() {
	fmt.Print("Hello from basic eval!")
}
`,
	}

	res := evalTool.Execute(context.Background(), args)
	assert.False(t, res.IsError)
	assert.Contains(t, res.ForLLM, "Hello from basic eval!")
}

func TestGoEvalTool_WithInjectedBindings(t *testing.T) {
	evalTool := NewGoEvalTool("/tmp/test_workspace")

	bindings := make(map[string]reflect.Value)
	bindings["Workspace"] = reflect.ValueOf("/tmp/test_workspace")
	bindings["HTTPClient"] = reflect.ValueOf(&http.Client{Timeout: 5 * time.Second})

	evalTool.SetBindings(bindings)

	args := map[string]any{
		"code": `
package main

import (
	"fmt"
	"jane/env"
)

func main() {
	fmt.Print("Workspace:", env.Workspace)
}
`,
	}

	res := evalTool.Execute(context.Background(), args)
	assert.False(t, res.IsError, "Execution should not fail: %v", res.ForLLM)
	assert.Contains(t, res.ForLLM, "Workspace:/tmp/test_workspace")
}

func TestGoEvalTool_Timeout(t *testing.T) {
	evalTool := NewGoEvalTool("/tmp")
	evalTool.timeout = 100 * time.Millisecond

	args := map[string]any{
		"code": `
package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Print("Start sleeping...")
	time.Sleep(2 * time.Second)
	fmt.Print("Done sleeping!")
}
`,
	}

	res := evalTool.Execute(context.Background(), args)
	assert.True(t, res.IsError)
	assert.Contains(t, res.ForLLM, "Execution timed out")
	assert.Contains(t, res.ForLLM, "Start sleeping...")
	assert.NotContains(t, res.ForLLM, "Done sleeping!")
}

func TestGoEvalTool_ErrorPropagation(t *testing.T) {
	evalTool := NewGoEvalTool("/tmp")

	args := map[string]any{
		"code": `
package main

import "fmt"

func main() {
	invalid_code
}
`,
	}

	res := evalTool.Execute(context.Background(), args)
	assert.True(t, res.IsError)
	assert.True(t, strings.Contains(res.ForLLM, "Execution failed:") || strings.Contains(res.ForLLM, "expected '=='"))
}
