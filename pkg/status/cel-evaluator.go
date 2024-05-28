package status

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types/ref"

	"k8s.io/apimachinery/pkg/runtime"
)

// celEvaluator is a struct that holds the CEL environment
// and provides a method to evaluate an expression with an unstructured object
// as the context.
type celEvaluator struct {
	env *cel.Env
}

// NewCELEvaluator initializes the CEL environment.
func newCELEvaluator() (*celEvaluator, error) {
	env, err := cel.NewEnv(
		cel.Declarations(
			decls.NewVar("obj", decls.NewMapType(decls.String, decls.Dyn)),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create CEL environment: %v", err)
	}

	return &celEvaluator{env: env}, nil
}

// preprocessExpression replaces occurrences of $ with obj.
func (e *celEvaluator) preprocessExpression(expression string) string {
	return strings.ReplaceAll(expression, "$", "obj")
}

// CheckExpression checks if an expression is valid.
func (e *celEvaluator) CheckExpression(expression string) error {
	expression = e.preprocessExpression(expression)

	ast, issues := e.env.Parse(expression)
	if issues != nil && issues.Err() != nil {
		return fmt.Errorf("failed to parse expression: %w", issues.Err())
	}

	_, issues = e.env.Check(ast)
	if issues != nil && issues.Err() != nil {
		return fmt.Errorf("failed to check expression: %w", issues.Err())
	}

	return nil
}

// Evaluate takes an expression and a Kubernetes raw object, and returns the
// evaluation of the expression with the object as the context.
func (e *celEvaluator) Evaluate(expression string, rawObj *runtime.RawExtension) (ref.Val, error) {
	expression = e.preprocessExpression(expression)

	ast, issues := e.env.Parse(expression)
	if issues != nil && issues.Err() != nil {
		return nil, fmt.Errorf("failed to parse expression: %w", issues.Err())
	}

	checked, issues := e.env.Check(ast)
	if issues != nil && issues.Err() != nil {
		return nil, fmt.Errorf("failed to check expression: %w", issues.Err())
	}

	// create the program
	prog, err := e.env.Program(checked)
	if err != nil {
		return nil, fmt.Errorf("failed to create program: %w", err)
	}

	// unmarshal the raw JSON data into a map
	var objMap map[string]interface{}
	err = json.Unmarshal(rawObj.Raw, &objMap)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal raw object: %w", err)
	}

	// evaluate the expression with the unstructured object
	result, _, err := prog.Eval(map[string]interface{}{
		"obj": objMap,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate expression: %w", err)
	}

	return result, nil
}
