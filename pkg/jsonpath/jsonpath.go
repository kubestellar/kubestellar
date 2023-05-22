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

package jsonpath

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/tidwall/sjson"
)

func ParsePath(pathStr string) (Path, error) {
	if !strings.HasPrefix(pathStr, "$.") {
		return Path{}, errors.New("syntax error: does not start with '$.'")
	}
	return Path{pathStr[2:]}, nil
}

type Path struct {
	sjsonPath string
}

func Update(data map[string]any, replacements ...Replacement) (map[string]any, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	dataStr := string(dataBytes)
	for _, repl := range replacements {
		dataStr, err = sjson.Set(dataStr, repl.Path.sjsonPath, repl.Value)
		if err != nil {
			return nil, fmt.Errorf("error on replacing %s: %w", repl.Path.sjsonPath, err)
		}
	}
	resultMap := map[string]any{}
	err = json.Unmarshal([]byte(dataStr), &resultMap)
	return resultMap, err
}

type Replacement struct {
	Path  Path
	Value any
}
