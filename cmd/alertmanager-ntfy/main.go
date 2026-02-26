package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/alexbakker/alertmanager-ntfy/internal/config"
	"github.com/alexbakker/alertmanager-ntfy/internal/server"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/providers/structs"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	k = koanf.NewWithConf(koanf.Conf{
		Delim:       ".",
		StrictMerge: false,
	})

	defaultConfig = config.Config{
		HTTP: &config.HTTP{
			Addr: ":8000",
		},
		Ntfy: &config.Ntfy{
			BaseURL: "https://ntfy.sh",
			Timeout: 10 * time.Second,
			Notification: config.Notification{
				Topic:               config.StringExpression{Text: getDefaultTopic()},
				Priority:            &config.StringExpression{Text: "default"},
				ConvertLabelsToTags: true,
			},
		},
		Log: getDefaultLogConfig(zapcore.InfoLevel),
	}
)

func main() {
	f := pflag.NewFlagSet("config", pflag.ContinueOnError)
	f.Usage = func() {
		fmt.Println("alertmanager-ntfy is a forwarder for Prometheus Alertmanager notifications")
		fmt.Println()
		fmt.Println("Flags:")
		fmt.Println(f.FlagUsages())
		os.Exit(0)
	}

	configFiles := f.StringSlice("configs", []string{"config.yml"}, "the yaml configuration files to load and merge")
	f.String("log-level", defaultConfig.Log.Level.String(), "the log level to use")
	f.String("http-addr", defaultConfig.HTTP.Addr, "the address to have the HTTP server listen on")
	f.String("ntfy-baseurl", defaultConfig.Ntfy.BaseURL, "the ntfy url to forward alerts to")
	f.String("ntfy-topic", defaultConfig.Ntfy.Notification.Topic.Text, "the ntfy topic")
	f.String("ntfy-priority", defaultConfig.Ntfy.Notification.Priority.Text, "the ntfy priority")
	f.Duration("ntfy-timeout", defaultConfig.Ntfy.Timeout, "the ntfy request timeout")
	if err := f.Parse(os.Args[1:]); err != nil {
		exitWithError(err.Error())
	}

	if configFiles == nil || len(*configFiles) == 0 {
		exitWithError("Empty --config flag")
	}

	if err := k.Load(structs.Provider(&defaultConfig, "yaml"), nil); err != nil {
		exitWithError(fmt.Sprintf("Failed to load default config: %v", err))
	}

	for _, c := range *configFiles {
		if err := k.Load(file.Provider(c), yaml.Parser()); err != nil {
			exitWithError(fmt.Sprintf("Failed to load config file: %v", err))
		}
	}

	if err := k.Load(posflag.ProviderWithValue(f, ".", k, convertFlag), nil); err != nil {
		exitWithError(fmt.Sprintf("Failed to load config from flags: %v", err))
	}

	var cfg config.Config
	if err := k.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{Tag: "yaml"}); err != nil {
		exitWithError(fmt.Sprintf("Failed to parse config file: %v", err))
	}

	logger, err := newLogger(cfg.Log)
	if err != nil {
		exitWithError(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	logger.Info("Starting HTTP server", zap.String("addr", cfg.HTTP.Addr), zap.Error(err))
	if err := server.New(logger, &cfg).Run(cfg.HTTP.Addr); err != nil {
		logger.Fatal("Failed to start HTTP server", zap.Error(err))
	}
}

func exitWithError(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

func convertFlag(key string, value string) (string, interface{}) {
	return strings.ReplaceAll(key, "-", "."), value
}
