package data

import (
	"fmt"
	"sort"
	"time"

	"github.com/charmbracelet/log"
	gh "github.com/cli/go-gh/v2/pkg/api"
)

var restClient *gh.RESTClient

// Actor represents the user who triggered the workflow run
type Actor struct {
	Login     string `json:"login"`
	AvatarUrl string `json:"avatar_url"`
}

// WorkflowRunRepository represents repo info from the REST API response
type WorkflowRunRepository struct {
	Id       int64  `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Private  bool   `json:"private"`
}

// WorkflowRunData represents a single workflow run from GitHub Actions REST API
type WorkflowRunData struct {
	Id           int64                 `json:"id"`
	Name         string                `json:"name"`
	DisplayTitle string                `json:"display_title"`
	Status       string                `json:"status"`
	Conclusion   string                `json:"conclusion"`
	HeadBranch   string                `json:"head_branch"`
	HeadSha      string                `json:"head_sha"`
	Event        string                `json:"event"`
	RunNumber    int                   `json:"run_number"`
	RunAttempt   int                   `json:"run_attempt"`
	WorkflowId   int64                 `json:"workflow_id"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
	RunStartedAt time.Time             `json:"run_started_at"`
	HtmlUrl      string                `json:"html_url"`
	Actor        Actor                 `json:"actor"`
	Repository   WorkflowRunRepository `json:"repository"`
}

// WorkflowRunsResponse represents the paginated response from GitHub API
type WorkflowRunsResponse struct {
	Runs       []WorkflowRunData
	TotalCount int
}

// REST API response structure
type workflowRunsAPIResponse struct {
	TotalCount   int               `json:"total_count"`
	WorkflowRuns []WorkflowRunData `json:"workflow_runs"`
}

// FetchWorkflowRuns fetches workflow runs for specified repos
func FetchWorkflowRuns(repos []string, limit int) (WorkflowRunsResponse, error) {
	var err error
	if restClient == nil {
		restClient, err = gh.DefaultRESTClient()
		if err != nil {
			return WorkflowRunsResponse{}, fmt.Errorf("failed to create REST client: %w", err)
		}
	}

	allRuns := make([]WorkflowRunData, 0)

	// Calculate per-repo limit
	perPage := limit
	if len(repos) > 1 {
		perPage = (limit / len(repos)) + 1
	}
	if perPage > 100 {
		perPage = 100 // GitHub API max
	}
	if perPage < 5 {
		perPage = 5
	}

	for _, repo := range repos {
		var response workflowRunsAPIResponse

		endpoint := fmt.Sprintf("repos/%s/actions/runs?per_page=%d", repo, perPage)
		log.Debug("Fetching workflow runs", "repo", repo, "endpoint", endpoint)

		err := restClient.Get(endpoint, &response)
		if err != nil {
			log.Warn("Failed to fetch workflow runs for repo", "repo", repo, "err", err)
			continue
		}

		log.Debug("Fetched workflow runs", "repo", repo, "count", len(response.WorkflowRuns))
		allRuns = append(allRuns, response.WorkflowRuns...)
	}

	// Sort by updated_at descending
	sort.Slice(allRuns, func(i, j int) bool {
		return allRuns[i].UpdatedAt.After(allRuns[j].UpdatedAt)
	})

	// Limit total results
	if len(allRuns) > limit {
		allRuns = allRuns[:limit]
	}

	return WorkflowRunsResponse{
		Runs:       allRuns,
		TotalCount: len(allRuns),
	}, nil
}

// FetchLatestWorkflowRuns fetches the latest run per workflow for specified repos
// This deduplicates to show only the most recent run for each workflow file
func FetchLatestWorkflowRuns(repos []string, limit int) (WorkflowRunsResponse, error) {
	// Fetch more runs to ensure we have enough after deduplication
	response, err := FetchWorkflowRuns(repos, limit*3)
	if err != nil {
		return response, err
	}

	// Deduplicate: keep only latest run per workflow per repo
	seen := make(map[string]bool)
	latestRuns := make([]WorkflowRunData, 0)

	for _, run := range response.Runs {
		// Key by repo + workflow ID to deduplicate
		key := fmt.Sprintf("%s:%d", run.Repository.FullName, run.WorkflowId)
		if !seen[key] {
			seen[key] = true
			latestRuns = append(latestRuns, run)
		}
	}

	// Limit to requested count
	if len(latestRuns) > limit {
		latestRuns = latestRuns[:limit]
	}

	return WorkflowRunsResponse{
		Runs:       latestRuns,
		TotalCount: len(latestRuns),
	}, nil
}

// GetRepoNameWithOwner returns the repo name in owner/name format
func (w WorkflowRunData) GetRepoNameWithOwner() string {
	return w.Repository.FullName
}

// GetTitle returns a display title for the workflow run
func (w WorkflowRunData) GetTitle() string {
	return w.DisplayTitle
}

// GetNumber returns the run number (used for interface compatibility)
func (w WorkflowRunData) GetNumber() int {
	return w.RunNumber
}

// GetUrl returns the HTML URL for the workflow run
func (w WorkflowRunData) GetUrl() string {
	return w.HtmlUrl
}

// GetUpdatedAt returns the last update time
func (w WorkflowRunData) GetUpdatedAt() time.Time {
	return w.UpdatedAt
}

// Duration returns the duration of the workflow run
func (w WorkflowRunData) Duration() time.Duration {
	if w.Status != "completed" {
		// For running workflows, show elapsed time
		return time.Since(w.RunStartedAt)
	}
	// For completed workflows, show total duration
	return w.UpdatedAt.Sub(w.RunStartedAt)
}

// ShortSha returns the first 7 characters of the commit SHA
func (w WorkflowRunData) ShortSha() string {
	if len(w.HeadSha) >= 7 {
		return w.HeadSha[:7]
	}
	return w.HeadSha
}
