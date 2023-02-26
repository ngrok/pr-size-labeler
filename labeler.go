package main

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/google/go-github/v50/github"
	"github.com/sethvargo/go-githubactions"
	"k8s.io/test-infra/prow/gitattributes"
)

type GitHubPRSizeLabeler struct {
	action *githubactions.Action
	client *github.Client
	event  github.PullRequestEvent

	labels []Label `yaml:"labels"`
}

func newGitHubPRSizeLabeler(client *github.Client, action *githubactions.Action, labels []Label) (*GitHubPRSizeLabeler, error) {
	event, err := getPREvent(action)
	if err != nil {
		return nil, err
	}

	return &GitHubPRSizeLabeler{
		client: client,
		action: action,
		event:  event,
		labels: labels,
	}, nil
}

func (l *GitHubPRSizeLabeler) repoOwner() string {
	return *l.event.PullRequest.Base.Repo.Owner.Login
}

func (l *GitHubPRSizeLabeler) repoName() string {
	return *l.event.PullRequest.Base.Repo.Name
}

func (l *GitHubPRSizeLabeler) prNumber() int {
	return *l.event.PullRequest.Number
}

func (l *GitHubPRSizeLabeler) currentPRLabels() []*github.Label {
	return l.event.PullRequest.Labels
}

func (l *GitHubPRSizeLabeler) prHasLabel(label string) bool {
	for _, l := range l.currentPRLabels() {
		if *l.Name == label {
			return true
		}
	}
	return false
}

func (l *GitHubPRSizeLabeler) loadGitAttributesFile() func() ([]byte, error) {
	return func() ([]byte, error) {
		content, err := fs.ReadFile(".gitattributes")
		if err != nil {
			l.action.Debugf("Successfully loaded .gitattributes file")
		}
		return content, err
	}
}

func (l *GitHubPRSizeLabeler) hasGitattributesFile() bool {
	_, err := fs.Stat(".gitattributes")
	if os.IsNotExist(err) {
		return false
	}
	return true
}

// CreateSizeLabels creates or updates the configured size labels for the
// repository.
func (l *GitHubPRSizeLabeler) CreateSizeLabels(ctx context.Context) error {
	l.action.Group("Creating or Updating configured size labels for repository")
	defer l.action.EndGroup()

	remoteLabels, err := l.getAllLabels(ctx)
	if err != nil {
		return err
	}

	for _, label := range l.labels {
		remoteLabel, ok := remoteLabels[label.Name]
		if !ok {
			l.action.Infof("Creating label %s", label.Name)
			// label doesn't exist, create it
			_, _, err := l.client.Issues.CreateLabel(ctx, l.repoOwner(), l.repoName(), &github.Label{
				Name:        &label.Name,
				Description: &label.Description,
				Color:       &label.Color,
			})
			if err != nil {
				return err
			}
			continue
		}

		// label exists, check if it needs to be updated
		if !label.Matches(*remoteLabel) {
			l.action.Infof("Label %s exists but is out of date, updating", label.Name)
			_, _, err := l.client.Issues.EditLabel(ctx, l.repoOwner(), l.repoName(), label.Name, &github.Label{
				Name:        &label.Name,
				Description: &label.Description,
				Color:       &label.Color,
			})
			if err != nil {
				return err
			}
			continue
		}

		l.action.Infof("Label %s already exists and is up to date", label.Name)
	}
	return nil
}

// AddSizeLabel adds the appropriate size label to the PR based on the number of lines changed.
// If the PR has a label that is no longer applicable, it will be removed.
// If there is a .gitattributes file in the repository, linguist generated files will be ignored in
// calculating the number of lines changed.
func (l *GitHubPRSizeLabeler) AddSizeLabel(ctx context.Context) error {
	l.action.Group("Adding/Updating size label for PR")
	defer l.action.EndGroup()

	filesChanged, err := l.getPRFilesChanged(ctx)
	if err != nil {
		return err
	}

	var ga *gitattributes.Group

	if !l.hasGitattributesFile() {
		l.action.Infof("No .gitattributes file found, skipping linguist generated file checks")
	} else {
		ga, err = gitattributes.NewGroup(l.loadGitAttributesFile())
		if err != nil {
			return err
		}
		l.action.Infof("Ignoring linguist generated files based on .gitattributes file")
	}

	var linesChanged int
	for _, change := range filesChanged {
		if ga != nil && ga.IsLinguistGenerated(*change.Filename) {
			l.action.Debugf("Skipping linguist generated file %s", *change.Filename)
			continue
		}
		linesChanged += *change.Additions + *change.Deletions
	}

	l.action.Infof("Calculated PR %d has %d lines changed", l.prNumber(), linesChanged)

	sizeLabels := l.labels
	// Sort the size labels from largest to smallest
	sort.Slice(sizeLabels, func(i, j int) bool {
		return sizeLabels[i].MinLines > sizeLabels[j].MinLines
	})

	var newLabel string
	// Find the first label in decreasing order that has a MinLines value less than the number of lines changed
	// Also remove any labels that are no longer applicable
	for _, label := range sizeLabels {
		if newLabel == "" && linesChanged >= label.MinLines {
			newLabel = label.Name
			continue
		}

		if l.prHasLabel(label.Name) {
			err := l.removeLabel(ctx, label.Name)
			if err != nil {
				l.action.Warningf("Failed to remove label %s: %v", label.Name, err)
			}
		}
	}

	if l.prHasLabel(newLabel) {
		l.action.Infof("PR already has label %s, skipping", newLabel)
		return nil
	}

	return l.addLabel(ctx, newLabel)
}

// getAllLabels returns a map of all labels in the repository key'd by the label name.
func (l *GitHubPRSizeLabeler) getAllLabels(ctx context.Context) (map[string]*github.Label, error) {
	l.action.Infof("Getting all labels for repository")
	labels := map[string]*github.Label{}
	opts := &github.ListOptions{PerPage: 100}

	for {
		page, resp, err := l.client.Issues.ListLabels(ctx, l.repoOwner(), l.repoName(), opts)
		if err != nil {
			return labels, err
		}

		for _, label := range page {
			labels[*label.Name] = label
		}

		if resp.NextPage == 0 {
			break
		}

		opts.Page = resp.NextPage
	}

	l.action.Infof("Found %d labels", len(labels))
	return labels, nil
}

func (l *GitHubPRSizeLabeler) getPRFilesChanged(ctx context.Context) ([]github.CommitFile, error) {
	filesChanged := []github.CommitFile{}

	l.action.Infof("Getting files changed in pr #%d", l.prNumber())

	opts := &github.ListOptions{PerPage: 100}

	for {
		page, resp, err := l.client.PullRequests.ListFiles(ctx, l.repoOwner(), l.repoName(), l.prNumber(), opts)
		if err != nil {
			return filesChanged, err
		}

		for _, c := range page {
			filesChanged = append(filesChanged, *c)
		}

		if resp.NextPage == 0 {
			break
		}

		opts.Page = resp.NextPage
	}

	l.action.Infof("Found %d files changed in pr", len(filesChanged))
	return filesChanged, nil
}

func (l *GitHubPRSizeLabeler) addLabel(ctx context.Context, label string) error {
	l.action.Infof("Adding label %s to pr", label)
	_, _, err := l.client.Issues.AddLabelsToIssue(ctx, l.repoOwner(), l.repoName(), l.prNumber(), []string{label})
	return err
}

func (l *GitHubPRSizeLabeler) removeLabel(ctx context.Context, label string) error {
	l.action.Infof("Removing label %s from pr", label)
	_, err := l.client.Issues.RemoveLabelForIssue(ctx, l.repoOwner(), l.repoName(), l.prNumber(), label)
	return err
}

func getPREvent(action *githubactions.Action) (github.PullRequestEvent, error) {
	ghContext, err := action.Context()
	if err != nil {
		return github.PullRequestEvent{}, err
	}

	payloadRaw, err := fs.ReadFile(ghContext.EventPath)
	if err != nil {
		return github.PullRequestEvent{}, err
	}

	event, err := github.ParseWebHook(ghContext.EventName, payloadRaw)
	if err != nil {
		return github.PullRequestEvent{}, err
	}

	switch event := event.(type) {
	case *github.PullRequestEvent:
		action.Debugf("event: %v", event)
		return *event, nil
	default:
		return github.PullRequestEvent{}, fmt.Errorf("Event is not a pull request event")
	}
}
