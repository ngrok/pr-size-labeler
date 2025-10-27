package main

import "github.com/google/go-github/v50/github"

// LabelEvent is an interface that represents a GitHub event that can be used for labeling.
type LabelEvent interface {
	// The name of the repository
	RepoName() string
	// The owner of the repository
	RepoOwner() string
	// The number of the pull request
	PRNumber() int
	// Current labels on the pull request
	PRLabels() []*github.Label
}

// PullRequestEvent represents a GitHub Pull Request event.
// It implements the LabelEvent interface.
type PullRequestEvent struct {
	event github.PullRequestEvent
}

// PRLabels returns the labels on the pull request.
func (e PullRequestEvent) PRLabels() []*github.Label {
	return e.event.PullRequest.Labels
}

// PRNumber returns the number of the pull request.
func (e PullRequestEvent) PRNumber() int {
	return *e.event.PullRequest.Number
}

// RepoName returns the name of the repository.
func (e PullRequestEvent) RepoName() string {
	return *e.event.PullRequest.Base.Repo.Name
}

// RepoOwner returns the owner of the repository.
func (e PullRequestEvent) RepoOwner() string {
	return *e.event.PullRequest.Base.Repo.Owner.Login
}

// PullRequestTargetEvent represents a GitHub Pull Request Target event.
// It implements the LabelEvent interface.
type PullRequestTargetEvent struct {
	event github.PullRequestTargetEvent
}

// PRLabels returns the labels on the pull request.
func (e PullRequestTargetEvent) PRLabels() []*github.Label {
	return e.event.PullRequest.Labels
}

// PRNumber returns the number of the pull request.
func (e PullRequestTargetEvent) PRNumber() int {
	return *e.event.PullRequest.Number
}

// RepoName returns the name of the repository.
func (e PullRequestTargetEvent) RepoName() string {
	return *e.event.PullRequest.Base.Repo.Name
}

// RepoOwner returns the owner of the repository.
func (e PullRequestTargetEvent) RepoOwner() string {
	return *e.event.PullRequest.Base.Repo.Owner.Login
}
