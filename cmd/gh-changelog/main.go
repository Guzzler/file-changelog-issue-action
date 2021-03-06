package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/google/go-github/v29/github"

	"go.uber.org/zap"

	"golang.org/x/oauth2"
)

var (
	logger *zap.Logger
)

func init() {
	logger, _ = zap.NewDevelopment()
}

func main() {
	ctx := context.Background()

	if eventType := os.Getenv("GITHUB_EVENT_NAME"); eventType != "pull_request" {
		logger.Fatal("Could support event type other than \"pull_request\"", zap.String("eventType", eventType))
	}

	eventPath := os.Getenv("GITHUB_EVENT_PATH")
	if eventPath == "" {
		logger.Fatal("Environment variable GITHUB_EVENT_PATH is mandatory")
	}

	payload, err := ioutil.ReadFile(eventPath)
	if err != nil {
		logger.Fatal("Could not read event file", zap.Error(err), zap.String("eventPath", eventPath))
	}

	var event github.PullRequestEvent
	err = json.Unmarshal(payload, &event)
	if err != nil {
		logger.Fatal("Could not unmarshal an event payload", zap.Error(err))
	}

	owner := event.GetRepo().GetOwner().GetLogin()
	repo := event.GetRepo().GetName()
	number := event.GetPullRequest().GetNumber()

	logger.Info("Successfully read an event payload",
		zap.String("owner", owner),
		zap.String("repo", repo),
		zap.Int("number", number),
	)

	var tc *http.Client
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc = oauth2.NewClient(ctx, ts)
	}
	client := github.NewClient(tc)

	successfullyAdded, err := createChangeLogIssue(ctx, client, owner, repo, number)
	if err != nil {
		logger.Fatal("Could not get createIssue", zap.Error(err))
	}
	logger.Info(fmt.Sprintf("The files were added with status %t to the change log", successfullyAdded))
}
