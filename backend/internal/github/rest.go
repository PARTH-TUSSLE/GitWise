package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// GHUser represents a GitHub user account response.
type GHUser struct {
	Login           string    `json:"login"`
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	AvatarURL       string    `json:"avatar_url"`
	Bio             string    `json:"bio"`
	Company         string    `json:"company"`
	Location        string    `json:"location"`
	Blog            string    `json:"blog"`
	TwitterUsername string    `json:"twitter_username"`
	PublicRepos     int       `json:"public_repos"`
	PublicGists     int       `json:"public_gists"`
	Followers       int       `json:"followers"`
	Following       int       `json:"following"`
	CreatedAt       time.Time `json:"created_at"`
}

// GHRepo represents a GitHub repository response.
type GHRepo struct {
	Name            string `json:"name"`
	FullName        string `json:"full_name"`
	Description     string `json:"description"`
	StargazersCount int    `json:"stargazers_count"`
	ForksCount      int    `json:"forks_count"`
	Language        string `json:"language"`
	DefaultBranch   string `json:"default_branch"`
	HTMLURL         string `json:"html_url"`
	Fork            bool   `json:"fork"`
	Owner           struct {
		Login string `json:"login"`
	} `json:"owner"`
}

// GHEvent represents an activity stream event from GitHub.
type GHEvent struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Actor struct {
		Login string `json:"login"`
	} `json:"actor"`
	Repo struct {
		Name string `json:"name"`
	} `json:"repo"`
	Payload   json.RawMessage `json:"payload"`
	CreatedAt time.Time       `json:"created_at"`
}

// GHSearchIssuesResult represents a paginated search response from GitHub search issues API.
type GHSearchIssuesResult struct {
	TotalCount        int  `json:"total_count"`
	IncompleteResults bool `json:"incomplete_results"`
}

// GetUser fetches basic profile information for a GitHub user.
func (c *Client) GetUser(ctx context.Context, username string) (*GHUser, error) {
	endpoint := fmt.Sprintf("/users/%s", url.PathEscape(username))
	resp, err := c.Do(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := c.checkResponseStatus(resp, "user", username); err != nil {
		return nil, err
	}

	var user GHUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user json: %w", err)
	}

	return &user, nil
}

// GetUserRepos fetches repositories owned or accessible by the user, sorted by recent activity.
func (c *Client) GetUserRepos(ctx context.Context, username string, perPage int) ([]GHRepo, error) {
	if perPage <= 0 || perPage > 100 {
		perPage = 30
	}
	endpoint := fmt.Sprintf("/users/%s/repos?type=owner&sort=pushed&per_page=%d", url.PathEscape(username), perPage)
	resp, err := c.Do(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := c.checkResponseStatus(resp, "user_repos", username); err != nil {
		return nil, err
	}

	var repos []GHRepo
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, fmt.Errorf("failed to decode repos json: %w", err)
	}

	return repos, nil
}

// GetUserEvents fetches recent public activity events for the given user.
func (c *Client) GetUserEvents(ctx context.Context, username string, perPage int) ([]GHEvent, error) {
	if perPage <= 0 || perPage > 100 {
		perPage = 100
	}
	endpoint := fmt.Sprintf("/users/%s/events/public?per_page=%d", url.PathEscape(username), perPage)
	resp, err := c.Do(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := c.checkResponseStatus(resp, "user_events", username); err != nil {
		return nil, err
	}

	var events []GHEvent
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, fmt.Errorf("failed to decode events json: %w", err)
	}

	return events, nil
}

// SearchUserPRs searches public PRs authored by the user with the given state (e.g. "is:merged" or "is:open").
func (c *Client) SearchUserPRs(ctx context.Context, username, stateQuery string) (int, error) {
	query := fmt.Sprintf("type:pr author:%s %s", username, stateQuery)
	endpoint := fmt.Sprintf("/search/issues?q=%s&per_page=1", url.QueryEscape(query))

	resp, err := c.Do(ctx, "GET", endpoint, nil)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if err := c.checkResponseStatus(resp, "search_prs", username); err != nil {
		return 0, err
	}

	var result GHSearchIssuesResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode search prs json: %w", err)
	}

	return result.TotalCount, nil
}

// SearchUserIssues searches public issues authored by the user.
func (c *Client) SearchUserIssues(ctx context.Context, username string) (int, error) {
	query := fmt.Sprintf("type:issue author:%s", username)
	endpoint := fmt.Sprintf("/search/issues?q=%s&per_page=1", url.QueryEscape(query))

	resp, err := c.Do(ctx, "GET", endpoint, nil)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if err := c.checkResponseStatus(resp, "search_issues", username); err != nil {
		return 0, err
	}

	var result GHSearchIssuesResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode search issues json: %w", err)
	}

	return result.TotalCount, nil
}

// GHIssue represents a GitHub repository issue response.
type GHIssue struct {
	ID        int64     `json:"id"`
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	State     string    `json:"state"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	User      struct {
		Login string `json:"login"`
	} `json:"user"`
	Labels []struct {
		Name string `json:"name"`
	} `json:"labels"`
	PullRequest *struct {
		URL string `json:"url"`
	} `json:"pull_request,omitempty"`
}

// GetIssue fetches a specific issue by owner, repo name, and issue number.
func (c *Client) GetIssue(ctx context.Context, owner, repo string, number int) (*GHIssue, error) {
	endpoint := fmt.Sprintf("/repos/%s/%s/issues/%d", url.PathEscape(owner), url.PathEscape(repo), number)
	resp, err := c.Do(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := c.checkResponseStatus(resp, "issue", fmt.Sprintf("%s/%s#%d", owner, repo, number)); err != nil {
		return nil, err
	}

	var issue GHIssue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, fmt.Errorf("failed to decode issue json: %w", err)
	}

	return &issue, nil
}

// GetRepoIssues fetches recent issues for a repository (excluding pull requests).
func (c *Client) GetRepoIssues(ctx context.Context, owner, repo string, perPage int) ([]GHIssue, error) {
	if perPage <= 0 || perPage > 100 {
		perPage = 30
	}
	endpoint := fmt.Sprintf("/repos/%s/%s/issues?state=open&sort=updated&per_page=%d",
		url.PathEscape(owner), url.PathEscape(repo), perPage,
	)

	resp, err := c.Do(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := c.checkResponseStatus(resp, "repo_issues", fmt.Sprintf("%s/%s", owner, repo)); err != nil {
		return nil, err
	}

	var allItems []GHIssue
	if err := json.NewDecoder(resp.Body).Decode(&allItems); err != nil {
		return nil, fmt.Errorf("failed to decode repo issues json: %w", err)
	}

	// Filter out pull requests
	issues := make([]GHIssue, 0, len(allItems))
	for _, item := range allItems {
		if item.PullRequest == nil {
			issues = append(issues, item)
		}
	}

	return issues, nil
}

// GHPullRequest represents a GitHub pull request response.
type GHPullRequest struct {
	ID        int64     `json:"id"`
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	State     string    `json:"state"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	User      struct {
		Login string `json:"login"`
	} `json:"user"`
	Head struct {
		Ref string `json:"ref"`
		SHA string `json:"sha"`
	} `json:"head"`
	Base struct {
		Ref string `json:"ref"`
		SHA string `json:"sha"`
	} `json:"base"`
	Additions    int `json:"additions"`
	Deletions    int `json:"deletions"`
	ChangedFiles int `json:"changed_files"`
	Commits      int `json:"commits"`
}

// GetPullRequest fetches details for a specific pull request.
func (c *Client) GetPullRequest(ctx context.Context, owner, repo string, number int) (*GHPullRequest, error) {
	endpoint := fmt.Sprintf("/repos/%s/%s/pulls/%d", url.PathEscape(owner), url.PathEscape(repo), number)
	resp, err := c.Do(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := c.checkResponseStatus(resp, "pull_request", fmt.Sprintf("%s/%s#%d", owner, repo, number)); err != nil {
		return nil, err
	}

	var pr GHPullRequest
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return nil, fmt.Errorf("failed to decode pull request json: %w", err)
	}

	return &pr, nil
}

// GetRepoPullRequests fetches recent open pull requests for a repository.
func (c *Client) GetRepoPullRequests(ctx context.Context, owner, repo string, perPage int) ([]GHPullRequest, error) {
	if perPage <= 0 || perPage > 100 {
		perPage = 30
	}
	endpoint := fmt.Sprintf("/repos/%s/%s/pulls?state=open&sort=updated&per_page=%d",
		url.PathEscape(owner), url.PathEscape(repo), perPage,
	)

	resp, err := c.Do(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := c.checkResponseStatus(resp, "repo_pulls", fmt.Sprintf("%s/%s", owner, repo)); err != nil {
		return nil, err
	}

	var prs []GHPullRequest
	if err := json.NewDecoder(resp.Body).Decode(&prs); err != nil {
		return nil, fmt.Errorf("failed to decode repo pull requests json: %w", err)
	}

	return prs, nil
}

// GetPullRequestDiff fetches the raw unified diff for a pull request.
func (c *Client) GetPullRequestDiff(ctx context.Context, owner, repo string, number int) (string, error) {
	endpoint := fmt.Sprintf("%s/repos/%s/%s/pulls/%d", c.baseURL, url.PathEscape(owner), url.PathEscape(repo), number)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github.v3.diff")
	req.Header.Set("User-Agent", "GitWise-Backend/2.1")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch pr diff: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch pr diff: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read pr diff: %w", err)
	}

	return string(body), nil
}
