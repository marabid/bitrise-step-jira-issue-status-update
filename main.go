package main

import (
	"fmt"
	"os"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-tools/go-steputils/stepconf"
)

// Config ...
type Config struct {
	BaseURL   string `env:"base_url,required"`
	UserName  string `env:"user_name,required"`
	APIToken  string `env:"api_token,required"`
	IssueKeys string `env:"issue_keys,required"`
	ToStatus  string `env:"to_status,required"`
}

func main() {
	var cfg Config
	if err := stepconf.Parse(&cfg); err != nil {
		failf("Issue with input: %s", err)
	}

	stepconf.Print(cfg)
	fmt.Println()
}

func failf(message string, arguments ...interface{}) {
	log.Errorf(message, arguments...)
	os.Exit(1)
}
