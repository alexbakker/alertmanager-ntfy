package alertmanager

import (
	"time"

	"github.com/fatih/structs"
)

// Source: https://github.com/prometheus/alertmanager/blob/ba8da18fb2b769ace00d270d677980c4d57310e7/template/template.go

type Data struct {
	Receiver string   `json:"receiver"`
	Status   string   `json:"status"`
	Alerts   []*Alert `json:"alerts"`

	GroupLabels       map[string]string `json:"groupLabels"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`

	ExternalURL string `json:"externalURL"`
}

type Alert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}

func (a *Alert) Map() map[string]interface{} {
	s := structs.New(a)
	s.TagName = "json"
	return s.Map()
}
