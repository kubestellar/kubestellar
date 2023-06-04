/*
Copyright 2023 The KubeStellar Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/kcp-dev/edge-mc/pkg/jsonpath"
)

func main() {
	args := os.Args[1:]
	if len(args)%2 != 0 {
		fmt.Fprint(os.Stderr, "Usage: path replacement path replacement ... <input >output\n")
		os.Exit(1)
	}
	var data jsonpath.JSONValue
	inputBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read input: %v\n", err)
		os.Exit(10)
	}
	err = json.Unmarshal(inputBytes, &data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse input: %v\n", err)
		os.Exit(20)
	}
	for index := 0; index < len(args); index += 2 {
		pathStr := args[index]
		replacementStr := args[index+1]
		parsedPath, err := jsonpath.ParseString(pathStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse %dth JSONPath: %v\n", index/2, err)
			os.Exit(30)
		}
		var replacementVal any
		err = json.Unmarshal([]byte(replacementStr), &replacementVal)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse %dth replacement: %v\n", index/2, err)
			os.Exit(31)
		}
		data = jsonpath.Apply(data, parsedPath, true, func(any) any { return replacementVal })
	}
	outputBytes, err := json.Marshal(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal result: %v\n", err)
		os.Exit(80)
	}
	os.Stdout.Write(outputBytes)
	os.Stdout.Write([]byte{'\n'})
}
