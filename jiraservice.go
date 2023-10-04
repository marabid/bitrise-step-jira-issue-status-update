package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/urlutil"
)

type jiraServiceFactory interface {
	create(user string, token string, baseURL string) jiraService
}

type jiraService interface {
	getAvailableTransitions(issueKey string) (*Transitions, error)
	makeTransition(issueKey string, transition Transition) error
}

type httpJiraServiceFactory struct{}

func (factory httpJiraServiceFactory) create(user string, token string, baseURL string) jiraService {
	basicAuth := []byte(user + `:` + token)
	return httpJiraService{
		client: http.Client{
			Timeout: time.Second * 10,
		},
		baseURL:    baseURL,
		authHeader: "Basic " + base64.StdEncoding.EncodeToString(basicAuth),
	}
}

type httpJiraService struct {
	client     http.Client
	baseURL    string
	authHeader string
}

func (service httpJiraService) getAvailableTransitions(issueKey string) (*Transitions, error) {
	httpURL, err := urlutil.Join(service.baseURL, "rest/api/3/issue", strings.TrimSpace(issueKey), "transitions")
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(http.MethodGet, httpURL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", service.authHeader)
	request.Header.Set("Content-Type", "application/json")

	response, err := service.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer silentClose(response.Body)

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var transitions Transitions
	err = json.Unmarshal(body, &transitions)
	if err != nil {
		return nil, err
	}

	return &transitions, nil
}

func (service httpJiraService) makeTransition(issueKey string, transition Transition) error {
	httpURL, err := urlutil.Join(service.baseURL, "rest/api/3/issue", strings.TrimSpace(issueKey), "transitions")
	if err != nil {
		return err
	}

	requestBody, err := json.Marshal(TransitionRequestBody{
		Transition: transition,
	})
	if err != nil {
		return err
	}

	request, err := http.NewRequest(http.MethodPost, httpURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", service.authHeader)
	request.Header.Set("Content-Type", "application/json")

	response, err := service.client.Do(request)
	if err != nil {
		return err
	}
	defer silentClose(response.Body)

	return nil
}

// To ...
type To struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (to To) String() string {
	return fmt.Sprintf("To(ID=%s, Name=%s)", to.ID, to.Name)
}

// Transition ...
type Transition struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	To   To     `json:"to"`
}

func (transition Transition) String() string {
	return fmt.Sprintf("Transition(ID=%s, Name=%s, To=%s)", transition.ID, transition.Name, transition.To)
}

// Transitions ...
type Transitions struct {
	Transitions []Transition `json:"transitions"`
}

func (transitions Transitions) String() string {
	var slice []string
	for _, transition := range transitions.Transitions {
		slice = append(slice, transition.String())
	}
	return fmt.Sprintf("[%s]", strings.Join(slice, ", "))
}

func (transitions Transitions) findTransition(toID string) (*Transition, error) {
	var matchingTransitions []Transition
	for _, transition := range transitions.Transitions {
		if transition.To.ID == toID {
			matchingTransitions = append(matchingTransitions, transition)
		}
	}

	transitionsCount := len(matchingTransitions)
	if transitionsCount <= 0 {
		return nil, fmt.Errorf("No matching transitions found. Available transitions: %s", transitions)
	} else if transitionsCount > 1 {
		var slice []string
		for _, transition := range transitions.Transitions {
			slice = append(slice, transition.String())
		}
		return nil, fmt.Errorf("More than 1 matching transition found. Matching transitions: %s", strings.Join(slice, ", "))
	} else {
		return &matchingTransitions[0], nil
	}
}

// TransitionRequestBody ...
type TransitionRequestBody struct {
	Transition Transition `json:"transition"`
}

func silentClose(closer io.Closer) {
	err := closer.Close()
	if err != nil {
		log.Warnf("Failed to close, error: %s", err)
	}
}
