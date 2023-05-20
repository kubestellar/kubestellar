/*
Copyright 2022 The KCP Authors.

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
	"os"

	"k8s.io/client-go/pkg/version"
)

func main() {
	args := os.Args
	vi := version.Get()
	switch len(args) {
	case 1:
		enc, err := json.Marshal(vi)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(enc))
		return
	case 2:
		field := args[1]
		var value string
		switch field {
		case "buildDate":
			value = vi.BuildDate
		case "gitCommit":
			value = vi.GitCommit
		case "gitTreeState":
			value = vi.GitTreeState
		case "platform":
			value = vi.Platform
		default:
			fmt.Fprintf(os.Stderr, "Invalid component requested: %q\n", field)
			goto Usage
		}
		fmt.Println(value)
		return
	default:
		fmt.Fprintf(os.Stderr, "%s: too many arguments\n", args[0])
	}
Usage:
	fmt.Fprintf(os.Stderr, "Usage: %s [buildDate|gitCommit|gitTreeState|platform]\n", args[0])
	os.Exit(1)
}
