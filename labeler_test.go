package main

import (
	"context"
	"testing"

	"github.com/google/go-github/v50/github"
	"github.com/ngrok/pr-size-labeler/mocks"
	"github.com/sethvargo/go-githubactions"
	"github.com/stretchr/testify/assert"
)

func newTestLabeler(event LabelEvent, labels []Label, issuesClient *mocks.IssuesClient, prClient *mocks.PullRequestsClient) *GitHubPRSizeLabeler {
	return &GitHubPRSizeLabeler{
		event:        event,
		labels:       labels,
		issues:       issuesClient,
		pullRequests: prClient,
		action:       githubactions.New(),
	}
}

func TestPrHasLabel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		labeler       *GitHubPRSizeLabeler
		labelToCheck  string
		expectedFound bool
	}{
		{
			name: "PR has the label",
			labeler: &GitHubPRSizeLabeler{
				event: newTestGitHubPullRequestEvent(1, "repo", "owner", []string{"size/S", "bug"}),
			},
			labelToCheck:  "size/S",
			expectedFound: true,
		},
		{
			name: "PR does not have the label",
			labeler: &GitHubPRSizeLabeler{
				event: newTestGitHubPullRequestEvent(1, "repo", "owner", []string{"size/S", "bug"}),
			},
			labelToCheck:  "size/XL",
			expectedFound: false,
		},
		{
			name: "PR has no labels",
			labeler: &GitHubPRSizeLabeler{
				event: newTestGitHubPullRequestEvent(1, "repo", "owner", []string{}),
			},
			labelToCheck:  "size/M",
			expectedFound: false,
		},
		{
			name: "PR has multiple labels, checking one",
			labeler: &GitHubPRSizeLabeler{
				event: newTestGitHubPullRequestEvent(1, "repo", "owner", []string{"size/L", "feature", "critical", "needs-review"}),
			},
			labelToCheck:  "critical",
			expectedFound: true,
		},
		{
			name: "label check is case-sensitive",
			labeler: &GitHubPRSizeLabeler{
				event: newTestGitHubPullRequestEvent(1, "repo", "owner", []string{"size/S"}),
			},
			labelToCheck:  "size/s",
			expectedFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.labeler.prHasLabel(tt.labelToCheck)
			assert.Equal(t, tt.expectedFound, result)
		})
	}
}

func TestCreateSizeLabels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		existingLabels map[string]*github.Label
		configLabels   []Label
		expectedCreate int
		expectedEdit   int
	}{
		{
			name:           "create all labels when none exist",
			existingLabels: map[string]*github.Label{},
			configLabels: []Label{
				{Name: "size/XS", Description: "Extra Small", Color: "00ff00", MinLines: 0},
				{Name: "size/S", Description: "Small", Color: "00ff00", MinLines: 10},
				{Name: "size/M", Description: "Medium", Color: "ffff00", MinLines: 100},
			},
			expectedCreate: 3,
			expectedEdit:   0,
		},
		{
			name: "no changes when labels are up to date",
			existingLabels: map[string]*github.Label{
				"size/XS": {Name: ptr("size/XS"), Description: ptr("Extra Small"), Color: ptr("00ff00")},
				"size/S":  {Name: ptr("size/S"), Description: ptr("Small"), Color: ptr("00ff00")},
			},
			configLabels: []Label{
				{Name: "size/XS", Description: "Extra Small", Color: "00ff00", MinLines: 0},
				{Name: "size/S", Description: "Small", Color: "00ff00", MinLines: 10},
			},
			expectedCreate: 0,
			expectedEdit:   0,
		},
		{
			name: "update labels when description changes",
			existingLabels: map[string]*github.Label{
				"size/S": {Name: ptr("size/S"), Description: ptr("Old Description"), Color: ptr("00ff00")},
			},
			configLabels: []Label{
				{Name: "size/S", Description: "New Description", Color: "00ff00", MinLines: 10},
			},
			expectedCreate: 0,
			expectedEdit:   1,
		},
		{
			name: "update labels when color changes",
			existingLabels: map[string]*github.Label{
				"size/M": {Name: ptr("size/M"), Description: ptr("Medium"), Color: ptr("ff0000")},
			},
			configLabels: []Label{
				{Name: "size/M", Description: "Medium", Color: "00ff00", MinLines: 100},
			},
			expectedCreate: 0,
			expectedEdit:   1,
		},
		{
			name: "mix of create and update",
			existingLabels: map[string]*github.Label{
				"size/S": {Name: ptr("size/S"), Description: ptr("Old"), Color: ptr("00ff00")},
			},
			configLabels: []Label{
				{Name: "size/S", Description: "New", Color: "00ff00", MinLines: 10},
				{Name: "size/M", Description: "Medium", Color: "ffff00", MinLines: 100},
				{Name: "size/L", Description: "Large", Color: "ff0000", MinLines: 500},
			},
			expectedCreate: 2,
			expectedEdit:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockIssues := mocks.NewIssuesClient()
			mockIssues.Labels = tt.existingLabels

			labeler := newTestLabeler(
				newTestGitHubPullRequestEvent(1, "repo", "owner", []string{}),
				tt.configLabels,
				mockIssues,
				mocks.NewPullRequestsClient(),
			)

			err := labeler.CreateSizeLabels(context.Background())
			assert.NoError(t, err)
			assert.Len(t, mockIssues.CreatedLabels, tt.expectedCreate)
			assert.Len(t, mockIssues.EditedLabels, tt.expectedEdit)
		})
	}
}

func TestAddSizeLabel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		currentLabels  []string
		filesChanged   []*github.CommitFile
		configLabels   []Label
		expectedAdd    string
		expectedRemove []string
	}{
		{
			name:          "add XS label for small PR",
			currentLabels: []string{},
			filesChanged: []*github.CommitFile{
				{Filename: ptr("file1.go"), Additions: ptr(3), Deletions: ptr(2)},
			},
			configLabels: []Label{
				{Name: "size/XS", MinLines: 0},
				{Name: "size/S", MinLines: 10},
				{Name: "size/M", MinLines: 100},
			},
			expectedAdd:    "size/XS",
			expectedRemove: []string{},
		},
		{
			name:          "add S label for small PR",
			currentLabels: []string{},
			filesChanged: []*github.CommitFile{
				{Filename: ptr("file1.go"), Additions: ptr(8), Deletions: ptr(5)},
			},
			configLabels: []Label{
				{Name: "size/XS", MinLines: 0},
				{Name: "size/S", MinLines: 10},
				{Name: "size/M", MinLines: 100},
			},
			expectedAdd:    "size/S",
			expectedRemove: []string{},
		},
		{
			name:          "add M label for medium PR",
			currentLabels: []string{},
			filesChanged: []*github.CommitFile{
				{Filename: ptr("file1.go"), Additions: ptr(75), Deletions: ptr(30)},
			},
			configLabels: []Label{
				{Name: "size/XS", MinLines: 0},
				{Name: "size/S", MinLines: 10},
				{Name: "size/M", MinLines: 100},
			},
			expectedAdd:    "size/M",
			expectedRemove: []string{},
		},
		{
			name:          "replace old label with new label",
			currentLabels: []string{"size/S"},
			filesChanged: []*github.CommitFile{
				{Filename: ptr("file1.go"), Additions: ptr(75), Deletions: ptr(30)},
			},
			configLabels: []Label{
				{Name: "size/XS", MinLines: 0},
				{Name: "size/S", MinLines: 10},
				{Name: "size/M", MinLines: 100},
			},
			expectedAdd:    "size/M",
			expectedRemove: []string{"size/S"},
		},
		{
			name:          "don't add label if already present",
			currentLabels: []string{"size/M"},
			filesChanged: []*github.CommitFile{
				{Filename: ptr("file1.go"), Additions: ptr(75), Deletions: ptr(30)},
			},
			configLabels: []Label{
				{Name: "size/XS", MinLines: 0},
				{Name: "size/S", MinLines: 10},
				{Name: "size/M", MinLines: 100},
			},
			expectedAdd:    "",
			expectedRemove: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockIssues := mocks.NewIssuesClient()
			mockPR := mocks.NewPullRequestsClient()
			mockPR.FilesChanged = tt.filesChanged

			labeler := newTestLabeler(
				newTestGitHubPullRequestEvent(1, "repo", "owner", tt.currentLabels),
				tt.configLabels,
				mockIssues,
				mockPR,
			)

			err := labeler.AddSizeLabel(t.Context())
			assert.NoError(t, err)

			if tt.expectedAdd != "" {
				assert.Contains(t, mockIssues.AddedLabels, tt.expectedAdd)
			} else {
				assert.Empty(t, mockIssues.AddedLabels)
			}

			assert.ElementsMatch(t, tt.expectedRemove, mockIssues.RemovedLabels)
		})
	}
}
