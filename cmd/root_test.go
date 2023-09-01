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
	"os"
	"testing"
)

func TestExportDefaultTemplate(t *testing.T) {
	if FileExists("collection.tmpl") {
		t.Errorf("FileExists(\"collection.tmpl\") = true, want false")
		return
	}
	fileName := exportDefaultTemplate()
	if fileName != "collection.tmpl" {
		t.Errorf("exportDefaultTemplate() = %q, want \"collection.tmpl\"", fileName)
	}
	if !FileExists(fileName) {
		t.Errorf("FileExists(%q) = false, want true", fileName)
	}
	os.Remove(fileName)
}

func TestArgsFunc(t *testing.T) {
	tests := []struct {
		name  string
		input []string
	}{
		{"[]string{\"api.json\"}", []string{"api.json"}},
		{"[]string{\"a.json\", \"out.txt\"}", []string{"a.json", "out.txt"}},
		{"[]string{\"-\"}", []string{"-"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := argsFunc(nil, test.input)
			if err != nil {
				t.Errorf("argsFunc(nil, %q) = %q, want nil", test.name, err)
			}
		})
	}
}

func TestArgsFuncWithInvalidArgs(t *testing.T) {
	tests := []struct {
		name  string
		input []string
	}{
		{"nil", nil},
		{"[]string{\"a.json\", \"b\", \"c\"}", []string{"a.json", "b", "c"}},
		{"[]string{\"file.txt\"}", []string{"file.txt"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := argsFunc(nil, test.input)
			if err == nil {
				t.Errorf("argsFunc(nil, %q) = nil, want non-nil error", test.name)
			}
		})
	}
}

func TestArgsFuncGetTemplate(t *testing.T) {
	GetTemplate = true
	err := argsFunc(nil, nil)
	if err != nil {
		t.Errorf("argsFunc(nil, nil) = %q, want nil", err)
	}
	GetTemplate = false
}

func TestArgsFuncWithCustomTmplPath(t *testing.T) {
	CustomTmplPath = "custom.tmpl"
	err := argsFunc(nil, []string{"api.json"})
	if err != nil {
		t.Errorf("argsFunc(nil, []string{\"api.json\"}) = %q, want nil", err)
	}
	CustomTmplPath = ""
}

func TestArgsFuncWithInvalidCustomTmplPath(t *testing.T) {
	CustomTmplPath = "custom.template"
	err := argsFunc(nil, []string{"api.json"})
	if err == nil {
		t.Errorf("argsFunc(nil, []string{\"api.json\"}) = nil, want non-nil error")
	}
	CustomTmplPath = ""
}

func TestLoadTmplDefault(t *testing.T) {
	tmplName, tmplStr, err := loadTmpl("")
	if tmplName != defaultTmplName {
		t.Errorf("loadTmpl(\"\") returned template name %q, want %q", tmplName, defaultTmplName)
	}
	assertNoDiff(t, tmplStr, defaultTmplStr, "\r\n")
	if err != nil {
		t.Errorf("loadTmpl(\"\") returned error %q, want nil error", err)
	}
}

func TestLoadTmplCustom(t *testing.T) {
	customTmplPath := "../samples/custom.tmpl"
	ansName, ansTmplStr, err := loadTmpl(customTmplPath)
	wantName := "custom.tmpl"
	customBytes, err := os.ReadFile(customTmplPath)
	if err != nil {
		t.Error(err)
		return
	}
	wantTmplStr := string(customBytes)

	if ansName != wantName {
		t.Errorf("loadTmpl(\"../samples/custom.tmpl\") returned template name %q, want %q", ansName, wantName)
	}
	assertNoDiff(t, ansTmplStr, wantTmplStr, "\r\n")
	if err != nil {
		t.Errorf("loadTmpl(\"../samples/custom.tmpl\") returned error %q, want nil error", err)
	}
}

func TestLoadTmplNonexistent(t *testing.T) {
	tmplName, tmplStr, err := loadTmpl("nonexistent.tmpl")
	if err == nil {
		t.Errorf("loadTmpl(\"nonexistent.tmpl\") = (%q, len %d template, nil), want non-nil error", tmplName, len(tmplStr))
	}
}
