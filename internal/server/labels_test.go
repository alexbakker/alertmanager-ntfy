package server

import (
	"strings"
	"testing"
	"text/template"

	"github.com/alexbakker/alertmanager-ntfy/internal/alertmanager"
	"github.com/alexbakker/alertmanager-ntfy/internal/config"
)

func TestRenderLabelsTemplate(t *testing.T) {
	tests := []struct {
		name         string
		templateStr  string
		labels       map[string]string
		expectedTags []string
	}{
		{
			name:         "nil template uses default behavior",
			templateStr:  "",
			labels:       map[string]string{"severity": "critical", "service": "api"},
			expectedTags: []string{"severity = critical", "service = api"},
		},
		{
			name:         "empty template returns no tags",
			templateStr:  "{{/* empty */}}",
			labels:       map[string]string{"severity": "critical"},
			expectedTags: []string{},
		},
		{
			name:         "custom format",
			templateStr:  "{{range $key, $value := .}}{{$key}}: {{$value}}, {{end}}",
			labels:       map[string]string{"severity": "critical"},
			expectedTags: []string{"severity: critical"},
		},
		{
			name:         "filter labels",
			templateStr:  "{{range $key, $value := .}}{{if ne $key \"internal\"}}{{$key}}={{$value}}, {{end}}{{end}}",
			labels:       map[string]string{"severity": "critical", "internal": "debug"},
			expectedTags: []string{"severity=critical"},
		},
		{
			name:         "capitalize function",
			templateStr:  "{{range $key, $value := .}}{{ capitalize $value }}, {{end}}",
			labels:       map[string]string{"env": "production"},
			expectedTags: []string{"Production"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Ntfy: &config.Ntfy{
					Notification: config.Notification{
						Templates: &config.Templates{},
					},
				},
			}

			if tt.templateStr != "" {
				tmpl, err := template.New("").Funcs(config.TemplateFuncs).Parse(tt.templateStr)
				if err != nil {
					t.Fatalf("Failed to parse template: %v", err)
				}
				cfg.Ntfy.Notification.Templates.Labels = (*config.Template)(tmpl)
			}

			server := &Server{cfg: cfg}
			alert := &alertmanager.Alert{Labels: tt.labels}

			tags, err := server.renderLabelsTemplate(alert)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(tags) != len(tt.expectedTags) {
				t.Errorf("Expected %d tags, got %d: %v", len(tt.expectedTags), len(tags), tags)
				return
			}

			tagSet := make(map[string]bool)
			for _, tag := range tags {
				tagSet[strings.TrimSpace(tag)] = true
			}

			for _, expected := range tt.expectedTags {
				if !tagSet[expected] {
					t.Errorf("Expected tag %q not found in %v", expected, tags)
				}
			}
		})
	}
}
