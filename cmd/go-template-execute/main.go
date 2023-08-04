/*
Copyright 2022 The KubeStellar Authors.

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
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	flag "github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

type Decoder interface {
	Decode(val interface{}) error
}

func main() {
	var dataYAML, dataJSON string
	dataFieldsYAML := []string{}
	dataFieldsJSON := []string{}
	flag.StringVar(&dataYAML, "data-yaml", dataYAML, "YAML data to read, literal or '-' for stdin or '@pathname' for file contents")
	flag.StringVar(&dataJSON, "data-json", dataJSON, "JSON data to read, literal or '-' for stdin or '@pathname' for file contents")
	flag.StringArrayVar(&dataFieldsYAML, "data-field-yaml", dataFieldsYAML, "TopField=data (literal or dash or @pathname)")
	flag.StringArrayVar(&dataFieldsJSON, "data-field-json", dataFieldsJSON, "TopField=data (literal or dash or @pathname)")
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "A template pathname must appear on the command line")
		os.Exit(1)
	}

	templatePath := flag.Arg(0)
	var err error
	data := map[string]any{}
	err = readFields(yaml.NewDecoder, dataFieldsYAML, data)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(20)
	}
	err = readFields(json.NewDecoder, dataFieldsJSON, data)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(21)
	}
	if dataJSON != "" {
		err = readData(json.NewDecoder, dataJSON, &data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read a JSON data object from %q: %v\n", dataJSON, err)
			os.Exit(31)
		}
	}
	if dataYAML != "" {
		err = readData(yaml.NewDecoder, dataYAML, &data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read a YAML data object from %q: %v\n", dataYAML, err)
			os.Exit(31)
		}
	}
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse template file %q: %v\n", templatePath, err)
		os.Exit(60)
	}
	err = tmpl.Execute(os.Stdout, data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to execute template %q on %v: %v\n", templatePath, data, err)
		os.Exit(80)
	}
}

var errEmpty = errors.New("empty data spec")

func readFields[ADecoder Decoder](decoderFactory func(io.Reader) ADecoder, dataFields []string, dest map[string]any) error {
	for _, dataSection := range dataFields {
		splitPos := strings.Index(dataSection, "=")
		if splitPos < 0 {
			return fmt.Errorf("no equals sign in data field %q", dataSection)
		}
		sectionName := dataSection[:splitPos]
		sectionPath := dataSection[splitPos+1:]
		var sectionData any
		err := readData(decoderFactory, sectionPath, &sectionData)
		if err != nil {
			return fmt.Errorf("failed to read data field %q: %w", dataSection, err)
		}
		dest[sectionName] = sectionData
	}
	return nil
}

func readData[ADecoder Decoder](decoderFactory func(io.Reader) ADecoder, dataArg string, dest interface{}) error {
	return openData(dataArg, func(reader io.Reader) error {
		decoder := decoderFactory(reader)
		return decoder.Decode(dest)
	})
}

func openData(dataArg string, kont func(io.Reader) error) error {
	if len(dataArg) == 0 {
		return errEmpty
	} else if dataArg == "-" {
		return kont(os.Stdin)
	} else if dataArg[0] == '@' {
		dataPath := dataArg[1:]
		dataFile, err := os.Open(dataPath)
		if err != nil {
			return err
		}
		defer dataFile.Close()
		return kont(dataFile)
	} else {
		return kont(strings.NewReader(dataArg))
	}
}
