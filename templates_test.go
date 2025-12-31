package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadEmbeddedTemplates(t *testing.T) {
	tests := []struct {
		name        string
		templateName string
		wantContain string
	}{
		{
			name:         "load default HTML",
			templateName: TemplateHTML,
			wantContain:  "<html>",
		},
		{
			name:         "load diffline HTML",
			templateName: TemplateDiffline,
			wantContain:  "<diffline",
		},
		{
			name:         "load CSS",
			templateName: TemplateCSS,
			wantContain:  "background",
		},
		{
			name:         "load JS",
			templateName: TemplateJS,
			wantContain:  "function",
		},
		{
			name:         "load SVG",
			templateName: TemplateShowHideIcon,
			wantContain:  "<svg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := LoadTemplate(tt.templateName)
			if err != nil {
				t.Fatalf("LoadTemplate(%q) error = %v", tt.templateName, err)
			}

			if !strings.Contains(content, tt.wantContain) {
				t.Errorf("LoadTemplate(%q) missing expected content %q", tt.templateName, tt.wantContain)
			}
		})
	}
}

func TestLoadTemplateNotFound(t *testing.T) {
	_, err := LoadTemplate("nonexistent.html")
	if err == nil {
		t.Error("LoadTemplate() expected error for nonexistent template, got nil")
	}
}

func TestLoadAllTemplates(t *testing.T) {
	css, js, showHideIcon, htmlTemplate, difflineTemplate, err := LoadAllTemplates()
	if err != nil {
		t.Fatalf("LoadAllTemplates() error = %v", err)
	}

	if css == "" {
		t.Error("LoadAllTemplates() css is empty")
	}
	if js == "" {
		t.Error("LoadAllTemplates() js is empty")
	}
	if showHideIcon == "" {
		t.Error("LoadAllTemplates() showHideIcon is empty")
	}
	if htmlTemplate == "" {
		t.Error("LoadAllTemplates() htmlTemplate is empty")
	}
	if difflineTemplate == "" {
		t.Error("LoadAllTemplates() difflineTemplate is empty")
	}
}

func TestMustLoadTemplate(t *testing.T) {
	// Should not panic for valid templates
	content := MustLoadTemplate(TemplateHTML)
	if content == "" {
		t.Error("MustLoadTemplate() returned empty content")
	}
}

func TestMustLoadTemplatePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustLoadTemplate() did not panic for nonexistent template")
		}
	}()

	MustLoadTemplate("nonexistent.html")
}

func TestTemplateOverride(t *testing.T) {
	// Create a temporary XDG_CONFIG_HOME
	tmpDir, err := os.MkdirTemp("", "cleandiff-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create the cleandiff config directory
	configDir := filepath.Join(tmpDir, "cleandiff")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Create a custom CSS file
	customCSS := "/* custom css */"
	customCSSPath := filepath.Join(configDir, "custom.css")
	if err := os.WriteFile(customCSSPath, []byte(customCSS), 0644); err != nil {
		t.Fatalf("Failed to write custom CSS: %v", err)
	}

	// Set XDG_CONFIG_HOME
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

	// Load the CSS template - should get the custom one
	content, err := LoadTemplate(TemplateCSS)
	if err != nil {
		t.Fatalf("LoadTemplate(%q) error = %v", TemplateCSS, err)
	}

	if content != customCSS {
		t.Errorf("LoadTemplate(%q) = %q, want %q (custom override)", TemplateCSS, content, customCSS)
	}
}

func TestAllOverrideMappings(t *testing.T) {
	// Verify all override mappings are defined
	expectedMappings := map[string]string{
		TemplateHTML:         "custom.html",
		TemplateDiffline:     "custom_diffline.html",
		TemplateCSS:          "custom.css",
		TemplateJS:           "custom.js",
		TemplateShowHideIcon: "custom_show_hide_icon.svg",
	}

	for template, expectedCustom := range expectedMappings {
		customName, ok := customOverrideNames[template]
		if !ok {
			t.Errorf("Missing override mapping for %q", template)
			continue
		}
		if customName != expectedCustom {
			t.Errorf("Override mapping for %q = %q, want %q", template, customName, expectedCustom)
		}
	}
}
