package config

import (
	"fmt"
	"net/url"
	urlpkg "net/url"
	"strings"
	"text/template"
	"time"
	"unicode"

	"github.com/PaesslerAG/gval"
	"go.uber.org/zap"
)

var (
	exprLang = gval.Full()
)

type Template template.Template

type Expression struct {
	Text      string
	Evaluable gval.Evaluable
}

type Templates struct {
	Title       *Template `yaml:"title"`
	Description *Template `yaml:"description"`
}

type Tag struct {
	Tag       string      `yaml:"tag"`
	Condition *Expression `yaml:"condition"`
}

type BasicAuth struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Priority struct {
	Text       string
	Expression *Expression
}

type Notification struct {
	Topic     string     `yaml:"topic"`
	Priority  *Priority  `yaml:"priority"`
	Tags      []*Tag     `yaml:"tags"`
	Templates *Templates `yaml:"templates"`
}

type Ntfy struct {
	BaseURL      string        `yaml:"baseurl"`
	Timeout      time.Duration `yaml:"timeout"`
	Auth         *BasicAuth    `yaml:"auth"`
	Notification Notification  `yaml:"notification"`
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

	url.Path, err = urlpkg.JoinPath(url.Path, c.Notification.Topic)
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

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (e *Expression) UnmarshalText(text []byte) error {
	s := strings.TrimSpace(string(text))
	evaluable, err := exprLang.NewEvaluable(s)
	if err != nil {
		return fmt.Errorf("bad expression: %w", err)
	}

	*e = Expression{
		Text:      s,
		Evaluable: evaluable,
	}
	return nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (p *Priority) UnmarshalText(text []byte) error {
	s := strings.TrimSpace(string(text))
	prio := Priority{Text: s}

	if !isAlphaNumeric(s) {
		var expr Expression
		if err := expr.UnmarshalText(text); err != nil {
			return err
		}

		prio.Expression = &expr
	}

	*p = prio
	return nil
}

// Valid reports whether this basic authentication configuration is considered
// valid. Returns false if it is nil, or if the username or password is empty.
func (a *BasicAuth) Valid() bool {
	return a != nil && a.Username != "" && a.Password != ""
}

func isAlphaNumeric(s string) bool {
	for _, c := range s {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
			return false
		}
	}

	return true
}
