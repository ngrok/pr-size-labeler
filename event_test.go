package main

import (
	"testing"

	"github.com/google/go-github/v50/github"
	"github.com/stretchr/testify/assert"
)

type LabelEventExpected struct {
	repoName  string
	repoOwner string
	prNumber  int
	labels    []string
}

func newTestGitHubPullRequestEvent(number int, repoName, repoOwner string, labels []string) PullRequestEvent {
	labelObjs := make([]*github.Label, len(labels))
	for i, label := range labels {
		labelObjs[i] = &github.Label{Name: ptr(label)}
	}

	return PullRequestEvent{
		event: github.PullRequestEvent{
			PullRequest: &github.PullRequest{
				Number: ptr(number),
				Labels: labelObjs,
				Base: &github.PullRequestBranch{
					Repo: &github.Repository{
						Name:  ptr(repoName),
						Owner: &github.User{Login: ptr(repoOwner)},
					},
				},
			},
		},
	}
}

func TestPullRequestEvent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		event    PullRequestEvent
		expected LabelEventExpected
	}{
		{
			name:  "standard PR with multiple labels",
			event: newTestGitHubPullRequestEvent(123, "test-repo", "test-owner", []string{"size/S", "bug"}),
			expected: LabelEventExpected{
				repoName:  "test-repo",
				repoOwner: "test-owner",
				prNumber:  123,
				labels:    []string{"size/S", "bug"},
			},
		},
		{
			name:  "PR with no labels",
			event: newTestGitHubPullRequestEvent(456, "my-project", "ngrok", []string{}),
			expected: LabelEventExpected{
				repoName:  "my-project",
				repoOwner: "ngrok",
				prNumber:  456,
				labels:    []string{},
			},
		},
		{
			name:  "PR with single label",
			event: newTestGitHubPullRequestEvent(789, "api-server", "acme-corp", []string{"enhancement"}),
			expected: LabelEventExpected{
				repoName:  "api-server",
				repoOwner: "acme-corp",
				prNumber:  789,
				labels:    []string{"enhancement"},
			},
		},
		{
			name:  "PR with many labels",
			event: newTestGitHubPullRequestEvent(1, "frontend", "company", []string{"size/XL", "bug", "high-priority", "needs-review", "documentation"}),
			expected: LabelEventExpected{
				repoName:  "frontend",
				repoOwner: "company",
				prNumber:  1,
				labels:    []string{"size/XL", "bug", "high-priority", "needs-review", "documentation"},
			},
		},
		{
			name:  "PR with special characters in repo name",
			event: newTestGitHubPullRequestEvent(9999, "my-awesome-repo-v2", "user-123", []string{"hotfix"}),
			expected: LabelEventExpected{
				repoName:  "my-awesome-repo-v2",
				repoOwner: "user-123",
				prNumber:  9999,
				labels:    []string{"hotfix"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected.repoName, tt.event.RepoName())
			assert.Equal(t, tt.expected.repoOwner, tt.event.RepoOwner())
			assert.Equal(t, tt.expected.prNumber, tt.event.PRNumber())

			labels := tt.event.PRLabels()
			assert.Len(t, labels, len(tt.expected.labels))
			for i, expectedLabel := range tt.expected.labels {
				assert.Equal(t, expectedLabel, *labels[i].Name)
			}
		})
	}
}

func newTestGitHubPullRequestTargetEvent(number int, repoName, repoOwner string, labels []string) PullRequestTargetEvent {
	labelObjs := make([]*github.Label, len(labels))
	for i, label := range labels {
		labelObjs[i] = &github.Label{Name: ptr(label)}
	}

	return PullRequestTargetEvent{
		event: github.PullRequestTargetEvent{
			PullRequest: &github.PullRequest{
				Number: ptr(number),
				Labels: labelObjs,
				Base: &github.PullRequestBranch{
					Repo: &github.Repository{
						Name:  ptr(repoName),
						Owner: &github.User{Login: ptr(repoOwner)},
					},
				},
			},
		},
	}
}

func TestPullRequestTargetEvent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		event    PullRequestTargetEvent
		expected LabelEventExpected
	}{
		{
			name:  "standard PR target with multiple labels",
			event: newTestGitHubPullRequestTargetEvent(234, "backend", "org-name", []string{"size/M", "feature"}),
			expected: LabelEventExpected{
				repoName:  "backend",
				repoOwner: "org-name",
				prNumber:  234,
				labels:    []string{"size/M", "feature"},
			},
		},
		{
			name:  "PR target with no labels",
			event: newTestGitHubPullRequestTargetEvent(567, "infrastructure", "devops-team", []string{}),
			expected: LabelEventExpected{
				repoName:  "infrastructure",
				repoOwner: "devops-team",
				prNumber:  567,
				labels:    []string{},
			},
		},
		{
			name:  "PR target with single label",
			event: newTestGitHubPullRequestTargetEvent(890, "core-lib", "platform", []string{"refactor"}),
			expected: LabelEventExpected{
				repoName:  "core-lib",
				repoOwner: "platform",
				prNumber:  890,
				labels:    []string{"refactor"},
			},
		},
		{
			name:  "PR target with many labels",
			event: newTestGitHubPullRequestTargetEvent(42, "auth-service", "security-team", []string{"size/L", "security", "critical", "approved"}),
			expected: LabelEventExpected{
				repoName:  "auth-service",
				repoOwner: "security-team",
				prNumber:  42,
				labels:    []string{"size/L", "security", "critical", "approved"},
			},
		},
		{
			name:  "PR target with numeric PR number",
			event: newTestGitHubPullRequestTargetEvent(10000, "legacy-system", "old-org", []string{"wontfix", "duplicate"}),
			expected: LabelEventExpected{
				repoName:  "legacy-system",
				repoOwner: "old-org",
				prNumber:  10000,
				labels:    []string{"wontfix", "duplicate"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected.repoName, tt.event.RepoName())
			assert.Equal(t, tt.expected.repoOwner, tt.event.RepoOwner())
			assert.Equal(t, tt.expected.prNumber, tt.event.PRNumber())

			labels := tt.event.PRLabels()
			assert.Len(t, labels, len(tt.expected.labels))
			for i, expectedLabel := range tt.expected.labels {
				assert.Equal(t, expectedLabel, *labels[i].Name)
			}
		})
	}
}

func ptr[T any](v T) *T {
	return &v
}
