package main

import (
	"embed"
	"os"
	"path/filepath"
)

//go:embed templates/*
var embeddedTemplates embed.FS

// Template names and their custom override names
const (
	TemplateHTML        = "default.html"
	TemplateDiffline    = "diffline.html"
	TemplateCSS         = "default.css"
	TemplateJS          = "default.js"
	TemplateShowHideIcon = "default_show_hide_icon.svg"
)

// customOverrideNames maps default template names to their custom override names
var customOverrideNames = map[string]string{
	TemplateHTML:         "custom.html",
	TemplateDiffline:     "custom_diffline.html",
	TemplateCSS:          "custom.css",
	TemplateJS:           "custom.js",
	TemplateShowHideIcon: "custom_show_hide_icon.svg",
}

// LoadTemplate loads a template by name, checking for user overrides first
// It looks in $XDG_CONFIG_HOME/cleandiff/ for custom overrides before
// falling back to the embedded default templates.
func LoadTemplate(name string) (string, error) {
	// Check for custom override
	if customName, ok := customOverrideNames[name]; ok {
		customPath := filepath.Join(getConfigDir(), customName)
		if content, err := os.ReadFile(customPath); err == nil {
			return string(content), nil
		}
		// Fall through to embedded template if custom doesn't exist
	}

	// Load from embedded templates
	content, err := embeddedTemplates.ReadFile("templates/" + name)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// MustLoadTemplate loads a template and panics if it fails
// Use this for templates that must exist (embedded defaults)
func MustLoadTemplate(name string) string {
	content, err := LoadTemplate(name)
	if err != nil {
		panic("failed to load template " + name + ": " + err.Error())
	}
	return content
}

// LoadAllTemplates loads all the templates needed for rendering
// Returns css, js, showHideIcon, htmlTemplate, difflineTemplate
func LoadAllTemplates() (css, js, showHideIcon, htmlTemplate, difflineTemplate string, err error) {
	css, err = LoadTemplate(TemplateCSS)
	if err != nil {
		return
	}

	js, err = LoadTemplate(TemplateJS)
	if err != nil {
		return
	}

	showHideIcon, err = LoadTemplate(TemplateShowHideIcon)
	if err != nil {
		return
	}

	htmlTemplate, err = LoadTemplate(TemplateHTML)
	if err != nil {
		return
	}

	difflineTemplate, err = LoadTemplate(TemplateDiffline)
	if err != nil {
		return
	}

	return
}
