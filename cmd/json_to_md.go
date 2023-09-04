// Copyright 2023 Chris Wheeler

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// 	http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"slices"
	"strconv"
	"strings"
)

// jsonToMdFile converts JSON bytes to markdown, prints the markdown to a file or
// stdout, and returns the destination's name. If the destination name is "-", output
// goes to stdout. If the destination's name is empty, a file is created with a unique
// name based on the given JSON. Only an empty destination name will be changed from
// what is given before being returned.
func jsonToMdFile(jsonBytes []byte, destName, tmplName, tmplStr string, statusRanges [][]int, confirmReplaceExistingFile bool) (string, error) {
	collection, err := parseCollection(jsonBytes)
	if err != nil {
		return "", err
	}
	filterResponsesByStatus(collection, statusRanges)

	collectionName := collection["info"].(map[string]any)["name"].(string)
	destFile, destName, err := getDestFile(destName, collectionName, confirmReplaceExistingFile)
	if err != nil {
		return "", err
	}
	if destName != "-" {
		// destFile is not os.Stdout
		defer destFile.Close()
	}

	if err = executeTemplate(destFile, collection, tmplName, tmplStr); err != nil {
		destFile.Close()
		os.Remove(destName)
		return "", err
	}

	return destName, nil
}

// parseCollection converts a collection from a slice of bytes of JSON to a map.
func parseCollection(jsonBytes []byte) (map[string]any, error) {
	var collection map[string]any
	if err := json.Unmarshal(jsonBytes, &collection); err != nil {
		return nil, err
	}
	if collection["info"].(map[string]any)["schema"] != "https://schema.getpostman.com/json/collection/v2.1.0/collection.json" {
		return nil, fmt.Errorf("Unknown JSON schema. When exporting from Postman, export as Collection v2.1.0")
	}

	return collection, nil
}

// parseStatusRanges converts a string of status ranges to a slice of slices of
// integers. The slice may be nil, but any inner slices each have two elements: the
// start and end of the range. Examples: "200", "200-299", "200-299,400-499", "200-200".
func parseStatusRanges(statusesStr string) ([][]int, error) {
	if len(statusesStr) == 0 {
		return nil, nil
	}
	statusRangeStrs := strings.Split(statusesStr, ",")
	statusRanges := make([][]int, len(statusRangeStrs))
	for i, statusRangeStr := range statusRangeStrs {
		startAndEnd := strings.Split(statusRangeStr, "-")
		if len(startAndEnd) > 2 {
			return nil, fmt.Errorf("Invalid status format. There should be zero or one dashes in %s", statusRangeStr)
		}
		start, err := strconv.Atoi(startAndEnd[0])
		if err != nil {
			return nil, fmt.Errorf("Invalid status range format. Expected an integer, got %q", startAndEnd[0])
		}
		end := start
		if len(startAndEnd) > 1 {
			end, err = strconv.Atoi(startAndEnd[1])
			if err != nil {
				return nil, fmt.Errorf("Invalid status range format. Expected an integer, got %q", startAndEnd[1])
			}
		}
		statusRanges[i] = make([]int, 2)
		statusRanges[i][0] = start
		statusRanges[i][1] = end
	}

	return statusRanges, nil
}

// filterResponsesByStatus removes all sample responses with status codes outside the
// given range(s). If no status ranges are given, the collection remains unchanged.
func filterResponsesByStatus(collection map[string]any, statusRanges [][]int) {
	if statusRanges == nil || len(statusRanges) == 0 {
		return
	}
	endpoints := collection["item"].([]any)
	for _, endpointAny := range endpoints {
		endpoint := endpointAny.(map[string]any)
		responses := endpoint["response"].([]any)
		for j := len(responses) - 1; j >= 0; j-- {
			response := responses[j].(map[string]any)
			inRange := false
			for _, statusRange := range statusRanges {
				code := int(response["code"].(float64))
				if code >= statusRange[0] && code <= statusRange[1] {
					inRange = true
					break
				}
			}
			if !inRange {
				responses = slices.Delete(responses, j, j+1)
				endpoint["response"] = responses
			}
		}
	}
}

// getDestFile gets the destination file and its name. If the given destination name is
// "-", the destination file is os.Stdout. If the given destination name is empty, a new
// file is created with a name based on the collection name and the returned name will
// be different from the given one. If the given destination name refers to an existing
// file and confirmation to replace an existing file is not given, an error is returned.
// Any returned file is open.
func getDestFile(destName, collectionName string, confirmReplaceExistingFile bool) (*os.File, string, error) {
	if destName == "-" {
		return os.Stdout, destName, nil
	}
	if len(destName) == 0 {
		fileName := FormatFileName(collectionName)
		if len(fileName) == 0 {
			fileName = "collection"
		}
		destName = CreateUniqueFileName(fileName, ".md")
	} else if FileExists(destName) && !confirmReplaceExistingFile {
		return nil, "", fmt.Errorf("File %q already exists. Run the command again with the --replace flag to confirm replacing it.", destName)
	}
	destFile, err := os.Create(destName)
	if err != nil {
		return nil, "", fmt.Errorf("os.Create: %s", err)
	}
	return destFile, destName, nil
}

// executeTemplate uses a template and FuncMap to convert the collection to markdown and
// saves to the given destination file. The destination file is not closed.
func executeTemplate(destFile *os.File, collection map[string]any, tmplName, tmplStr string) error {
	tmpl, err := template.New(tmplName).Funcs(funcMap).Parse(tmplStr)
	if err != nil {
		return fmt.Errorf("Template parsing error: %s", err)
	}

	return tmpl.Execute(destFile, collection)
}
