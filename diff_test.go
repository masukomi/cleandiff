package main

import (
	"strings"
	"testing"
)

func TestEscapeHTMLEntities(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no special characters",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "less than",
			input:    "a < b",
			expected: "a &lt; b",
		},
		{
			name:     "greater than",
			input:    "a > b",
			expected: "a &gt; b",
		},
		{
			name:     "both",
			input:    "<html>",
			expected: "&lt;html&gt;",
		},
		{
			name:     "multiple occurrences",
			input:    "if (a < b && c > d)",
			expected: "if (a &lt; b && c &gt; d)",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeHTMLEntities(tt.input)
			if result != tt.expected {
				t.Errorf("escapeHTMLEntities(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestProcessDiff(t *testing.T) {
	difflineTemplate := `<diffline><linum>{{ .LineNumberPair }}</linum><raw>{{ .RawText }}</raw></diffline>`

	tests := []struct {
		name        string
		file1       string
		file2       string
		wantContain []string
	}{
		{
			name:  "identical files",
			file1: "hello world",
			file2: "hello world",
			wantContain: []string{
				"<linum>1:1</linum>",
				"hello world",
			},
		},
		{
			name:  "simple change",
			file1: "hello",
			file2: "world",
			wantContain: []string{
				"<deleted>",
				"</deleted>",
				"<inserted>",
				"</inserted>",
			},
		},
		{
			name:  "multiline",
			file1: "line1\nline2",
			file2: "line1\nline2",
			wantContain: []string{
				"<linum>1:1</linum>",
				"<linum>2:2</linum>",
			},
		},
		{
			name:  "html entities escaped",
			file1: "<div>",
			file2: "<span>",
			wantContain: []string{
				"&lt;", // < gets escaped to &lt;
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ProcessDiff(tt.file1, tt.file2, difflineTemplate)
			if err != nil {
				t.Fatalf("ProcessDiff() error = %v", err)
			}

			for _, want := range tt.wantContain {
				if !strings.Contains(result, want) {
					t.Errorf("ProcessDiff() result missing %q\nGot: %s", want, result)
				}
			}
		})
	}
}

func TestGenerateHTML(t *testing.T) {
	tests := []struct {
		name        string
		file1       string
		file2       string
		fileName    string
		wantContain []string
	}{
		{
			name:     "basic HTML structure",
			file1:    "old content",
			file2:    "new content",
			fileName: "test.txt",
			wantContain: []string{
				"<html>",
				"</html>",
				"<title>test.txt</title>",
				"<style>",
				"</style>",
				"<script",
				"</script>",
				"<diff>",
				"</diff>",
			},
		},
		{
			name:     "filename in output",
			file1:    "a",
			file2:    "b",
			fileName: "myfile.go",
			wantContain: []string{
				"myfile.go",
			},
		},
		{
			name:     "diff markers present",
			file1:    "old",
			file2:    "new",
			fileName: "test.txt",
			wantContain: []string{
				"<deleted>",
				"<inserted>",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateHTML(tt.file1, tt.file2, tt.fileName)
			if err != nil {
				t.Fatalf("GenerateHTML() error = %v", err)
			}

			for _, want := range tt.wantContain {
				if !strings.Contains(result, want) {
					t.Errorf("GenerateHTML() result missing %q", want)
				}
			}
		})
	}
}
