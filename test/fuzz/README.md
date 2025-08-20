# KubeStellar Fuzzing Tests

This directory contains fuzzing tests for KubeStellar components to help identify potential bugs, crashes, and security vulnerabilities through automated testing with random inputs.

## Overview

Fuzzing is a testing technique that involves providing random, unexpected, or invalid data as input to a computer program to find bugs, crashes, or security vulnerabilities. The fuzzing tests in this directory use Go's built-in fuzzing support (introduced in Go 1.18) to test various KubeStellar components.

## Fuzzing Tests

### 1. FuzzJSONPathParsing
Tests the JSONPath parser with various inputs to ensure it handles edge cases and malformed queries correctly.

**Target Components:**
- `pkg/jsonpath/` - JSONPath parsing and evaluation

**Seed Corpus:**
- Basic JSONPath expressions like `$.spec.name`, `$.metadata.labels`
- Array access patterns like `$.spec.containers[0].name`
- Quoted string access like `$.metadata.annotations["key"]`

### 2. FuzzCRDValidation
Tests CRD validation with various YAML inputs to ensure proper handling of KubeStellar custom resources.

**Target Components:**
- `pkg/crd/` - Custom Resource Definition validation
- `api/control/v1alpha1/` - KubeStellar API types

**Seed Corpus:**
- Valid BindingPolicy, StatusCollector, and Binding YAML configurations
- Various combinations of these resources

### 3. FuzzLabelParsing
Tests label parsing functionality to ensure proper handling of Kubernetes label selectors.

**Target Components:**
- `pkg/util/labels.go` - Label parsing utilities

**Seed Corpus:**
- Valid label strings like `app=test`, `environment=production`
- Various label key-value combinations

### 4. FuzzAPIGroupParsing
Tests API group parsing functionality to ensure proper handling of API group restrictions.

**Target Components:**
- `pkg/util/resources.go` - API group parsing utilities

**Seed Corpus:**
- Empty strings, single groups, comma-separated groups
- Various API group combinations

### 5. FuzzJSONValueValidation
Tests JSON value validation to ensure proper type checking and validation.

**Target Components:**
- `api/control/v1alpha1/` - Value type definitions

**Seed Corpus:**
- Various JSON value types (string, number, bool, object, null)
- Different value representations

## Running Fuzzing Tests

### Local Development

Run all fuzzing tests with default duration (10 seconds):
```bash
make test-fuzz
```

Run fuzzing tests with short duration (5 seconds):
```bash
make test-fuzz-short
```

Run a specific fuzzing test:
```bash
cd test/fuzz
go test -fuzz=FuzzJSONPathParsing -fuzztime=30s
```

### Continuous Integration

The fuzzing tests are integrated into the CI pipeline and will run automatically on pull requests and merges.

**CI Fuzzing Workflow:**
- The `.github/workflows/cifuzz.yml` workflow runs all Go fuzzers in this directory on every push and pull request (except for documentation/OWNERS/MAINTAINERS changes).
- The workflow uses native Go fuzzing (Go 1.18+) to build and run the fuzzers for 2 minutes each.
- If any crash or bug is found, the workflow will fail and upload crash artifacts for maintainers to review.
- Maintainers should review any failed fuzzing jobs and examine the uploaded artifacts to diagnose and fix bugs.
- To add a new fuzzer to CI, simply add a new `Fuzz*` function in `kubestellar_fuzz_test.go` and add a corresponding step in the GitHub Actions workflow.

**Interpreting Results:**
- If the CI fuzzing job fails, check the 'fuzzing-artifacts' in the workflow run for crash details and input that triggered the bug.
- Fix the bug or update the fuzzer as needed, then re-run the workflow to verify the fix.

**Verification:**
- Run `./verify-setup.sh` in this directory to verify your local fuzzing setup
- The script checks Go version, dependencies, and runs a quick test


## Test Data

The `testdata/` directory contains seed corpus files that provide initial test cases for the fuzzers:

- `binding_policy.yaml` - Sample BindingPolicy configurations
- `status_collector.yaml` - Sample StatusCollector configurations  
- `binding.yaml` - Sample Binding configurations

## Adding New Fuzzing Tests

To add a new fuzzing test:

1. Create a new function with the prefix `Fuzz` in `kubestellar_fuzz_test.go`
2. Add seed corpus data using `f.Add()` calls
3. Implement the fuzzing logic in the `f.Fuzz()` callback
4. Add test data files to `testdata/` if needed
5. Update `oss_fuzz_build.sh` to include the new fuzzer
6. Update this README with documentation

Example:
```go
func FuzzNewComponent(f *testing.F) {
    // Add seed corpus
    f.Add("seed1")
    f.Add("seed2")
    
    f.Fuzz(func(t *testing.T, input string) {
        // Fuzzing logic here
        if result, err := processInput(input); err != nil {
            // Expected for invalid inputs
            return
        }
        
        // Validate result
        if result == nil {
            t.Errorf("Unexpected nil result")
        }
    })
}
```

## Best Practices

1. **Seed Corpus**: Always provide meaningful seed corpus data to guide the fuzzer
2. **Error Handling**: Distinguish between expected errors (invalid inputs) and unexpected errors (bugs)
3. **Validation**: Include basic validation logic to catch potential issues
4. **Performance**: Keep fuzzing tests efficient to allow for longer running times
5. **Documentation**: Document what each fuzzing test targets and why

## Troubleshooting

### Common Issues

1. **Import Errors**: Ensure all required packages are imported
2. **Type Errors**: Check that types match the expected interfaces
3. **Build Errors**: Verify that the fuzzing test compiles correctly

### Debugging

To debug a fuzzing test that's failing:

1. Run with verbose output: `go test -fuzz=FuzzTestName -v`
2. Use a longer fuzz time: `go test -fuzz=FuzzTestName -fuzztime=60s`
3. Check the generated test corpus in the `testdata/fuzz/` directory

## Contributing

When contributing fuzzing tests:

1. Follow the existing patterns and conventions
2. Ensure tests are deterministic and reproducible
3. Add appropriate documentation
4. Test locally before submitting
5. Consider the impact on CI build times

## References

- [Go Fuzzing Documentation](https://go.dev/doc/fuzz/)
- [Go Fuzzing Tutorial](https://go.dev/doc/tutorial/fuzz.html)
- [OSS-Fuzz Documentation](https://google.github.io/oss-fuzz/)
- [KubeStellar Architecture](https://kubestellar.io/docs/) 