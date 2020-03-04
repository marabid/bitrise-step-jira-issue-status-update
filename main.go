package main

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-tools/go-steputils/stepconf"
)

type configFactory interface {
	create() (*Config, error)
}

type bitriseConfigFactory struct{}

func (factory bitriseConfigFactory) create() (*Config, error) {
	var config Config
	if err := stepconf.Parse(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

// Config ...
type Config struct {
	BaseURL    string `env:"base_url,required"`
	UserName   string `env:"user_name,required"`
	APIToken   string `env:"api_token,required"`
	IssueKeys  string `env:"issue_keys,required"`
	ToStatusID string `env:"to_status_id,required"`
}

func main() {
	runStep(bitriseConfigFactory{}, httpJiraServiceFactory{})
}

func runStep(configFactory configFactory, serviceFactory jiraServiceFactory) {
	cfg, err := configFactory.create()
	if err != nil {
		failf("Issue with input: %s", err)
	}

	stepconf.Print(cfg)
	fmt.Println()

	service := serviceFactory.create(cfg.UserName, cfg.APIToken, cfg.BaseURL)
	var waitGroup sync.WaitGroup

	for _, issueKey := range distinct(strings.Split(cfg.IssueKeys, `|`)) {
		waitGroup.Add(1)
		go updateIssue(service, issueKey, cfg.ToStatusID, &waitGroup)
	}

	waitGroup.Wait()
	log.Infof("Processed all issues")
}

func updateIssue(service jiraService, issueKey string, statusID string, waitGroup *sync.WaitGroup) {
	log.Infof("Checking issue %s status", issueKey)

	transitions, err := service.getAvailableTransitions(issueKey)
	if err != nil {
		log.Warnf("Failed to get available transitions for issue %s, error: %s", issueKey, err)
		waitGroup.Done()
		return
	}

	transition, err := transitions.findTransition(statusID)
	if err != nil {
		log.Warnf("Failed to update status to %s for issue %s, error: %s", statusID, issueKey, err)
		waitGroup.Done()
		return
	}

	err = service.makeTransition(issueKey, *transition)
	if err != nil {
		log.Warnf("Failed to update status to %s for issue %s, error: %s", statusID, issueKey, err)
		waitGroup.Done()
		return
	}

	log.Infof("Successfully updated issue %s to status %s", issueKey, statusID)
	waitGroup.Done()
}

func distinct(input []string) []string {
	result := make([]string, 0, len(input))
	presenceFlags := make(map[string]bool)

	for _, value := range input {
		if _, exists := presenceFlags[value]; !exists {
			presenceFlags[value] = true
			result = append(result, value)
		}
	}
	return result
}

func failf(message string, arguments ...interface{}) {
	log.Errorf(message, arguments...)
	os.Exit(1)
}
