package config

import (
	"fmt"
	"net/url"
	urlpkg "net/url"
	"strings"
	"text/template"
	"time"

	"go.uber.org/zap"
)

type Template template.Template

type Templates struct {
	Title       *Template `yaml:"title"`
	Description *Template `yaml:"description"`
}

type BasicAuth struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Ntfy struct {
	BaseURL   string        `yaml:"baseurl"`
	Topic     string        `yaml:"topic"`
	Timeout   time.Duration `yaml:"timeout"`
	Auth      *BasicAuth    `yaml:"auth"`
	Priority  string        `yaml:"priority"`
	Templates *Templates    `yaml:"templates"`
}

type HTTP struct {
	Addr string     `yaml:"addr"`
	Auth *BasicAuth `yaml:"auth"`
}

type Config struct {
	HTTP *HTTP       `yaml:"http"`
	Ntfy *Ntfy       `yaml:"ntfy"`
	Log  *zap.Config `yaml:"log"`
}

func (c *Ntfy) URL() (*url.URL, error) {
	url, err := urlpkg.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	url.Path, err = urlpkg.JoinPath(url.Path, c.Topic)
	if err != nil {
		return nil, fmt.Errorf("url path join: %w", err)
	}

	return url, nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (t *Template) UnmarshalText(text []byte) error {
	s := strings.TrimSpace(string(text))

	tmpl, err := template.New("").Parse(s)
	if err != nil {
		return err
	}

	*t = Template(*tmpl)
	return nil
}

// Valid reports whether this basic authentication configuration is considered
// valid. Returns false if it is nil, or if the username or password is empty.
func (a *BasicAuth) Valid() bool {
	return a != nil && a.Username != "" && a.Password != ""
}
