package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	urlpkg "net/url"
	"strings"
	"text/template"

	"github.com/alexbakker/alertmanager-ntfy/internal/alertmanager"
	"github.com/alexbakker/alertmanager-ntfy/internal/config"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	keyRequestID = "request_id"
)

type Server struct {
	e      *gin.Engine
	cfg    *config.Config
	logger *zap.Logger
	http   *http.Client
}

func New(logger *zap.Logger, cfg *config.Config) *Server {
	gin.SetMode(gin.ReleaseMode)
	e := gin.New()
	e.Use(func(c *gin.Context) {
		// If there's no X-Request-Id in the headers, we generate one ourselves
		// so that we can correlate log lines to a single request
		var requestID string
		if requestID = c.Writer.Header().Get("X-Request-Id"); requestID == "" {
			requestID = generateRequestID()
		}

		c.Set(keyRequestID, requestID)
		c.Next()
	})
	e.Use(ginzap.GinzapWithConfig(logger, &ginzap.Config{
		Context: ginzap.Fn(func(c *gin.Context) []zapcore.Field {
			requestID, ok := c.Get("request_id")
			if !ok {
				panic("request_id is not set in gin context")
			}
			return []zapcore.Field{zap.String(keyRequestID, requestID.(string))}
		}),
	}))
	e.Use(ginzap.RecoveryWithZap(logger, true))

	if cfg.HTTP.Auth.Valid() {
		e.Use(gin.BasicAuth(gin.Accounts{
			cfg.HTTP.Auth.Username: cfg.HTTP.Auth.Password,
		}))
	} else {
		logger.Warn("Basic auth is disabled")
	}

	s := Server{
		e:      e,
		cfg:    cfg,
		logger: logger,
		http:   &http.Client{Timeout: cfg.Ntfy.Timeout},
	}
	s.e.POST("/hook", s.handleWebhook)
	return &s
}

func (s *Server) handleWebhook(c *gin.Context) {
	logger := s.logger
	if requestID, ok := c.Get(keyRequestID); ok {
		logger = logger.With(zap.String(keyRequestID, requestID.(string)))
	}
	logger.Info("Handling webhook")

	var payload alertmanager.Data
	if err := json.NewDecoder(c.Request.Body).Decode(&payload); err != nil {
		logger.Error("Failed to unmarshal webhook payload", zap.Error(err))
		c.Status(http.StatusBadRequest)
		return
	}

	if len(payload.Alerts) == 0 {
		logger.Warn("Received an empty list of alerts")
	} else {
		go s.forwardAlerts(logger, payload.Alerts)
	}

	c.Status(http.StatusAccepted)
}

func (s *Server) forwardAlerts(logger *zap.Logger, alerts []*alertmanager.Alert) {
	for _, alert := range alerts {
		logger := logger.With(zap.String("alert_fingerprint", alert.Fingerprint))
		if err := s.forwardAlert(logger, alert); err != nil {
			logger.Error("Failed to forward alert to ntfy", zap.Error(err))
		} else {
			logger.Info("Successfully forwarded alert to ntfy")
		}
	}
}

func (s *Server) forwardAlert(logger *zap.Logger, alert *alertmanager.Alert) error {
	var titleBuf bytes.Buffer
	if err := (*template.Template)(s.cfg.Ntfy.Notification.Templates.Title).Execute(&titleBuf, alert); err != nil {
		return fmt.Errorf("render title template: %w", err)
	}
	title := strings.TrimSpace(titleBuf.String())

	var descBuf bytes.Buffer
	if err := (*template.Template)(s.cfg.Ntfy.Notification.Templates.Description).Execute(&descBuf, alert); err != nil {
		return fmt.Errorf("render description template: %w", err)
	}
	description := strings.TrimSpace(descBuf.String())

	// If the description is empty, send the title as the description so that
	// the ntfy app doesn't fall back to setting "triggered" as the description.
	if description == "" {
		description = title
		title = ""
	}

	url, err := s.getUrl(alert)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url.String(), strings.NewReader(description))
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}

	if s.cfg.Ntfy.Auth.Valid() {
		req.SetBasicAuth(s.cfg.Ntfy.Auth.Username, s.cfg.Ntfy.Auth.Password)
	}

	var tags []string
	for _, tag := range s.cfg.Ntfy.Notification.Tags {
		if tag.Condition != nil {
			match, err := tag.Condition.Evaluable.EvalBool(context.Background(), alert.Map())
			if err != nil {
				// Expression evaluation errors should not prevent the notification from being sent
				logger.Warn(
					"Tag condition expression evaluation failed",
					zap.String("expression", tag.Condition.Text),
					zap.Error(err),
				)
				continue
			}

			if !match {
				continue
			}
		}

		tags = append(tags, tag.Tag)
	}
	tags = append(tags, convertLabelsToTags(alert.Labels)...)

	if title != "" {
		req.Header.Set("X-Title", title)
	}
	if len(tags) > 0 {
		req.Header.Set("X-Tags", strings.Join(tags, tagSeparator))
	}
	if s.cfg.Ntfy.Notification.Priority != nil {
		priority, err := evalStringExpr(s.cfg.Ntfy.Notification.Priority, alert)
		if err != nil {
			// Expression evaluation errors should not prevent the notification from being sent
			logger.Warn(
				"Priority expression evaluation failed",
				zap.String("expression", s.cfg.Ntfy.Notification.Priority.Expression.Text),
				zap.Error(err),
			)
		}

		if priority != "" {
			req.Header.Set("X-Priority", priority)
		}
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("http %d, %s", res.StatusCode, http.StatusText(res.StatusCode))
	}

	return nil
}

func (s *Server) Run(addr string) error {
	return s.e.Run(addr)
}

func (s *Server) getUrl(alert *alertmanager.Alert) (*urlpkg.URL, error) {
	url, err := urlpkg.Parse(s.cfg.Ntfy.BaseURL)
	if err != nil {
		return nil, err
	}

	topic, err := evalStringExpr(&s.cfg.Ntfy.Notification.Topic, alert)
	if err != nil {
		return nil, fmt.Errorf("topic expression eval: %w", err)
	}

	if topic == "" {
		return nil, errors.New("topic is empty")
	}

	url.Path, err = urlpkg.JoinPath(url.Path, topic)
	if err != nil {
		return nil, fmt.Errorf("url path join: %w", err)
	}

	return url, nil
}

func evalStringExpr(expr *config.StringExpression, alert *alertmanager.Alert) (string, error) {
	if expr.Expression != nil {
		return expr.Expression.Evaluable.EvalString(context.Background(), alert.Map())
	}

	return expr.Text, nil
}
