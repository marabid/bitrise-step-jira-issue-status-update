package main

import (
	"fmt"
)

type fakeConfigFactory struct {
	config Config
}

func (factory fakeConfigFactory) create() (*Config, error) {
	return &factory.config, nil
}

type fakeJiraServiceFactory struct {
	service fakeJiraService
}

func (factory fakeJiraServiceFactory) create(user string, token string, baseURL string) jiraService {
	return factory.service
}

type fakeJiraService struct {
	availableTransitions  map[string]Transitions
	disallowedTransitions []transitionIntent
}

func (service fakeJiraService) getAvailableTransitions(issueKey string) (*Transitions, error) {
	if transitions, exists := service.availableTransitions[issueKey]; exists {
		return &transitions, nil
	}
	return nil, fmt.Errorf("No transition available")
}

func (service fakeJiraService) makeTransition(issueKey string, transition Transition) error {
	intent := transitionIntent{
		issueKey:   issueKey,
		transition: transition,
	}
	if exists := exists(intent, service.disallowedTransitions); exists {
		return fmt.Errorf("Transition is not allowed")
	}

	return nil
}

func (service fakeJiraService) addTransitions(issueKey string, transitions Transitions) {
	service.availableTransitions[issueKey] = transitions
}

func exists(item transitionIntent, elements []transitionIntent) bool {
	for _, element := range elements {
		if item == element {
			return true
		}
	}
	return false
}

type transitionIntent struct {
	issueKey   string
	transition Transition
}
