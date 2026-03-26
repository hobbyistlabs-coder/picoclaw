1. **Understand the Goal**: The objective is to evolve JANE (PicoClaw) into an active, autonomous agent by providing features that enable autonomous skill acquisition and tool-calling capabilities. From the memory, we have the constraint: "The `go_eval` tool (using Yaegi) enables autonomous task execution by exposing internal agent bindings (such as Workspace, HTTPClient, and BrowserActionTool) to dynamic scripts via a synthetic `jane/env` package. Bindings are injected using `SetBindings()` with values wrapped in `reflect.ValueOf()`. When injecting synthetic packages into Yaegi via `i.Use(exports)`, provide a fallback path for package resolution (e.g., `"jane/env/env": bindings` alongside `"jane/env": bindings`) to prevent 'unable to find source related to' import errors."
2. **Implement SetBindings on GoEvalTool**:
   - In `pkg/tools/go_eval.go`, add a `bindings map[string]reflect.Value` to `GoEvalTool`.
   - Add a `SetBindings(bindings map[string]any)` method that populates `GoEvalTool.bindings` by wrapping each value in `reflect.ValueOf()`.
   - Update `Execute()` to register a synthetic package `jane/env` (and fallback `jane/env/env`) with Yaegi using `i.Use()` so that dynamically evaluated scripts can import `"jane/env"` and use the injected bindings.
3. **Inject Bindings into GoEvalTool in the Main Loop**:
   - In `pkg/agent/loop_init.go` (or wherever `NewGoEvalTool` is called and registered), after creating `GoEvalTool`, construct a map of bindings containing relevant active context objects (e.g., `Workspace`, `HTTPClient`, `BrowserActionTool` if enabled).
   - Call `goEvalTool.SetBindings(bindings)`.
4. **Testing**:
   - Update `pkg/tools/go_eval_test.go` to properly test that `SetBindings` allows scripts to import and access injected objects.
   - Run `go test ./pkg/tools` to verify.
5. **Pre-commit**: Complete pre-commit steps to ensure proper testing, verification, review, and reflection are done.
