package mocks

import (
	"context"

	"github.com/google/go-github/v50/github"
)

type IssuesClient struct {
	Labels         map[string]*github.Label
	CreatedLabels  []*github.Label
	EditedLabels   []*github.Label
	AddedLabels    []string
	RemovedLabels  []string
	CreateLabelErr error
	EditLabelErr   error
	AddLabelsErr   error
	RemoveLabelErr error
}

func NewIssuesClient() *IssuesClient {
	return &IssuesClient{
		Labels:        make(map[string]*github.Label),
		CreatedLabels: []*github.Label{},
		EditedLabels:  []*github.Label{},
		AddedLabels:   []string{},
		RemovedLabels: []string{},
	}
}

func (m *IssuesClient) ListLabels(ctx context.Context, owner, repo string, opts *github.ListOptions) ([]*github.Label, *github.Response, error) {
	labels := []*github.Label{}
	for _, label := range m.Labels {
		labels = append(labels, label)
	}
	return labels, &github.Response{NextPage: 0}, nil
}

func (m *IssuesClient) CreateLabel(ctx context.Context, owner, repo string, label *github.Label) (*github.Label, *github.Response, error) {
	if m.CreateLabelErr != nil {
		return nil, nil, m.CreateLabelErr
	}
	m.CreatedLabels = append(m.CreatedLabels, label)
	m.Labels[*label.Name] = label
	return label, &github.Response{}, nil
}

func (m *IssuesClient) EditLabel(ctx context.Context, owner, repo, name string, label *github.Label) (*github.Label, *github.Response, error) {
	if m.EditLabelErr != nil {
		return nil, nil, m.EditLabelErr
	}
	m.EditedLabels = append(m.EditedLabels, label)
	m.Labels[name] = label
	return label, &github.Response{}, nil
}

func (m *IssuesClient) AddLabelsToIssue(ctx context.Context, owner, repo string, number int, labels []string) ([]*github.Label, *github.Response, error) {
	if m.AddLabelsErr != nil {
		return nil, nil, m.AddLabelsErr
	}
	m.AddedLabels = append(m.AddedLabels, labels...)
	result := []*github.Label{}
	for _, label := range labels {
		labelCopy := label
		result = append(result, &github.Label{Name: &labelCopy})
	}
	return result, &github.Response{}, nil
}

func (m *IssuesClient) RemoveLabelForIssue(ctx context.Context, owner, repo string, number int, label string) (*github.Response, error) {
	if m.RemoveLabelErr != nil {
		return nil, m.RemoveLabelErr
	}
	m.RemovedLabels = append(m.RemovedLabels, label)
	return &github.Response{}, nil
}

type PullRequestsClient struct {
	FilesChanged []*github.CommitFile
}

func NewPullRequestsClient() *PullRequestsClient {
	return &PullRequestsClient{
		FilesChanged: []*github.CommitFile{},
	}
}

func (m *PullRequestsClient) ListFiles(ctx context.Context, owner, repo string, number int, opts *github.ListOptions) ([]*github.CommitFile, *github.Response, error) {
	return m.FilesChanged, &github.Response{NextPage: 0}, nil
}
