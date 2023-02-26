package main

import (
	"context"

	"github.com/google/go-github/v50/github"
	"github.com/sethvargo/go-githubactions"
)

func main() {
	action := githubactions.New()

	repoToken := action.GetInput("repo-token")
	if repoToken == "" {
		action.Fatalf("missing required input: repo-token")
	}

	configPath := action.GetInput("config-path")
	if configPath == "" {
		action.Fatalf("missing required input: config-path")
	}

	config, err := loadConfig(configPath)
	if err != nil {
		action.Fatalf("%v", err)
	}

	ctx := context.Background()
	client := github.NewTokenClient(ctx, repoToken)

	labeler, err := newGitHubPRSizeLabeler(client, action, config.Labels)
	if err != nil {
		action.Fatalf("%v", err)
	}

	if err := labeler.CreateSizeLabels(ctx); err != nil {
		action.Fatalf("%v", err)
	}

	if err := labeler.AddSizeLabel(ctx); err != nil {
		action.Fatalf("%v", err)
	}
}
