package main

import (
	"fmt"
	"os"
	"strings"

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

	service := createService(cfg.UserName, cfg.APIToken, cfg.BaseURL)

	for _, issueKey := range strings.Split(cfg.IssueKeys, `|`) {
		log.Infof("Checking issue %s status", issueKey)

		transitions, err := service.getAvailableTransitions(issueKey)
		if err != nil {
			log.Warnf("Failed to get status for issue %s, error: %s", issueKey, err)
			continue
		}

		transition, err := transitions.findTransition(cfg.ToStatus)
		if err != nil {
			log.Warnf("Failed to update status for issue %s, error: %s", issueKey, err)
			continue
		}

		err = service.makeTransition(issueKey, *transition)
		if err != nil {
			log.Warnf("Failed to update status for issue %s, error: %s", issueKey, err)
			continue
		}

		log.Infof("Successfully updated issue %s to status %s", issueKey, cfg.ToStatus)

		fmt.Println()
	}
}

func failf(message string, arguments ...interface{}) {
	log.Errorf(message, arguments...)
	os.Exit(1)
}
