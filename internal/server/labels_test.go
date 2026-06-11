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
		status       string
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
			templateStr:  "{{range $key, $value := .Labels}}{{$key}}: {{$value}}, {{end}}",
			labels:       map[string]string{"severity": "critical"},
			expectedTags: []string{"severity: critical"},
		},
		{
			name:         "filter labels",
			templateStr:  "{{range $key, $value := .Labels}}{{if ne $key \"internal\"}}{{$key}}={{$value}}, {{end}}{{end}}",
			labels:       map[string]string{"severity": "critical", "internal": "debug"},
			expectedTags: []string{"severity=critical"},
		},
		{
			name:         "capitalize function",
			templateStr:  "{{range $key, $value := .Labels}}{{ capitalize $value }}, {{end}}",
			labels:       map[string]string{"env": "production"},
			expectedTags: []string{"Production"},
		},
		{
			name:        "firing alert with emoji label uses emoji label value",
			templateStr: "{{- if eq .Status \"firing\" -}}{{- with index .Labels \"emoji\" -}}{{ . }}{{- else -}}rotating_light{{- end -}}{{- else -}}white_check_mark{{- end -}}",
			labels:      map[string]string{"emoji": "blue_car", "severity": "info"},
			status:      "firing",
			expectedTags: []string{"blue_car"},
		},
		{
			name:        "firing alert without emoji label falls back to rotating_light",
			templateStr: "{{- if eq .Status \"firing\" -}}{{- with index .Labels \"emoji\" -}}{{ . }}{{- else -}}rotating_light{{- end -}}{{- else -}}white_check_mark{{- end -}}",
			labels:      map[string]string{"severity": "warning"},
			status:      "firing",
			expectedTags: []string{"rotating_light"},
		},
		{
			name:        "resolved alert emits white_check_mark",
			templateStr: "{{- if eq .Status \"firing\" -}}{{- with index .Labels \"emoji\" -}}{{ . }}{{- else -}}rotating_light{{- end -}}{{- else -}}white_check_mark{{- end -}}",
			labels:      map[string]string{"severity": "info"},
			status:      "resolved",
			expectedTags: []string{"white_check_mark"},
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
			alert := &alertmanager.Alert{Labels: tt.labels, Status: tt.status}
			ctx := &templateContext{Alert: alert}

			tags, err := server.renderLabelsTemplate(ctx)
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
