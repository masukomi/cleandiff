package main

import (
	"bytes"
	"fmt"
	"html/template"
	"regexp"
	"strings"

	"github.com/dacharyc/tokendiff"
)

// Regexp to detect lines with changes (insertions or deletions)
var changedLineRegexp = regexp.MustCompile(`<inserted>.*</inserted>|<deleted>.*</deleted>`)

// HTMLData holds the data for rendering the main HTML template
type HTMLData struct {
	FileName     string
	CSS          template.CSS
	JS           template.JS
	ShowHideIcon template.HTML
	Diff         template.HTML
}

// DiffLineData holds the data for rendering a single diff line
type DiffLineData struct {
	LineNumberPair string
	RawText        template.HTML // Use template.HTML to prevent double-escaping
	IsUnchanged    bool          // True if line has no insertions or deletions
}

// escapeHTMLEntities replaces < and > with their HTML entities
func escapeHTMLEntities(s string) string {
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

// ProcessDiff takes two file contents and returns the processed diff lines
func ProcessDiff(file1Content, file2Content, difflineTemplate string) (string, error) {
	// Escape HTML entities in both inputs
	file1Escaped := escapeHTMLEntities(file1Content)
	file2Escaped := escapeHTMLEntities(file2Content)

	// Configure tokendiff options as per instructions
	opts := tokendiff.Options{
		// Delimiters from instructions: \x0A%,;/:._{}[]()'\!|-=~><"\\
		// Note: \x0A is newline, which is handled by line-by-line processing
		Delimiters: `%,;/:._{}[]()'\!|-=~><"\`,
	}

	fmtOpts := tokendiff.FormatOptions{
		StartDelete:   "<deleted>",
		StopDelete:    "</deleted>",
		StartInsert:   "<inserted>",
		StopInsert:    "</inserted>",
		RepeatMarkers: true,
		// We don't use ShowLineNumbers because we get line numbers from LineDiffResult
		ShowLineNumbers: false,
	}

	// Run the line-by-line diff
	output := tokendiff.DiffLineByLine(file1Escaped, file2Escaped, opts, fmtOpts, "best", 0.5)

	// Parse the diffline template
	tmpl, err := template.New("diffline").Parse(difflineTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse diffline template: %w", err)
	}

	// Process each line
	var diffLines []string
	for _, line := range output.Lines {
		// Format line number pair as "OldLineNum:NewLineNum"
		lineNumPair := fmt.Sprintf("%d:%d", line.OldLineNum, line.NewLineNum)

		// Check if line has any changes (insertions or deletions)
		changedLine := changedLineRegexp.MatchString(line.Output)

		// The Output field already has the formatted diff with <deleted>/<inserted> tags
		data := DiffLineData{
			LineNumberPair: lineNumPair,
			RawText:        template.HTML(line.Output),
			IsUnchanged:    !changedLine,
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return "", fmt.Errorf("failed to execute diffline template: %w", err)
		}

		diffLines = append(diffLines, buf.String())
	}

	// Join all lines with newlines
	return strings.Join(diffLines, "\n"), nil
}

// GenerateHTML generates the complete HTML output for the diff
func GenerateHTML(file1Content, file2Content, fileName string) (string, error) {
	// Load all templates
	css, js, showHideIcon, htmlTemplate, difflineTemplate, err := LoadAllTemplates()
	if err != nil {
		return "", fmt.Errorf("failed to load templates: %w", err)
	}

	// Process the diff
	diff, err := ProcessDiff(file1Content, file2Content, difflineTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to process diff: %w", err)
	}

	// Parse the HTML template
	tmpl, err := template.New("html").Parse(htmlTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML template: %w", err)
	}

	// Prepare the data
	data := HTMLData{
		FileName:     fileName,
		CSS:          template.CSS(css),
		JS:           template.JS(js),
		ShowHideIcon: template.HTML(showHideIcon),
		Diff:         template.HTML(diff),
	}

	// Execute the template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute HTML template: %w", err)
	}

	return buf.String(), nil
}
