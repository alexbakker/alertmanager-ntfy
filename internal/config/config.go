package config

import (
	"fmt"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/PaesslerAG/gval"
	"go.uber.org/zap"
)

var (
	exprLang = gval.Full()

	// Source: https://github.com/binwiederhier/ntfy/blob/30301c8a7ff9e54ae505daf73a7f1571e7fefae3/user/types.go#L245
	allowedTopicRegex = regexp.MustCompile(`^[-_A-Za-z0-9]{1,64}$`)
)

type Template template.Template

type Expression struct {
	Text      string
	Evaluable gval.Evaluable
}

type Templates struct {
	Title       *Template            `yaml:"title"`
	Description *Template            `yaml:"description"`
	Actions     []Action             `yaml:"actions,omitempty"`
	Headers     map[string]*Template `yaml:headers`
}

type Tag struct {
	Tag       string      `yaml:"tag"`
	Condition *Expression `yaml:"condition"`
}

type BasicAuth struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type StringExpression struct {
	Text       string
	Expression *Expression
}

type Notification struct {
	Topic               StringExpression  `yaml:"topic"`
	Priority            *StringExpression `yaml:"priority"`
	Tags                []*Tag            `yaml:"tags"`
	Templates           *Templates        `yaml:"templates"`
	ConvertLabelsToTags bool              `yaml:"convertLabelsToTags"`
}

type NtfyAuth struct {
	BasicAuth *BasicAuth `yaml:"basic"`
	Token     *string    `yaml:"token"`
}

type Ntfy struct {
	BaseURL      string        `yaml:"baseurl"`
	Timeout      time.Duration `yaml:"timeout"`
	Auth         *NtfyAuth     `yaml:"auth"`
	Notification Notification  `yaml:"notification"`
	Async        bool          `yaml:"async"`
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

type Action struct {
	Action    string      `yaml:"action"`
	Label     string      `yaml:"label"`
	URL       string      `yaml:"url"`
	Condition *Expression `yaml:"condition,omitempty"`
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
func (e *StringExpression) UnmarshalText(text []byte) error {
	s := strings.TrimSpace(string(text))
	se := StringExpression{Text: s}

	if isExpression(s) {
		var expr Expression
		if err := expr.UnmarshalText(text); err != nil {
			return err
		}

		se.Expression = &expr
	}

	*e = se
	return nil
}

// Valid reports whether this basic authentication configuration is considered
// valid. Returns false if it is nil, or if the username or password is empty.
func (a *BasicAuth) Valid() bool {
	return a != nil && a.Username != "" && a.Password != ""
}

// isExpression reports whether the given string is likely to be an expression by
// checking whether it'd be a valid topic.
func isExpression(s string) bool {
	return !allowedTopicRegex.Match([]byte(s))
}
