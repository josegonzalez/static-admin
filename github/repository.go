package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type FetchRepositoryInput struct {
	RepositoryURL string
	Token         string
}

func FetchRepository(input FetchRepositoryInput) (Repository, error) {
	if input.RepositoryURL == "" {
		return Repository{}, fmt.Errorf("repository URL is required")
	}

	if input.Token == "" {
		return Repository{}, fmt.Errorf("authentication token is required")
	}

	// Parse owner and repo from URL
	owner, name, err := parseGitHubURL(input.RepositoryURL)
	if err != nil {
		return Repository{}, fmt.Errorf("invalid repository URL: %v", err)
	}

	// Construct GitHub API URL
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, name)

	// Create HTTP request
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return Repository{}, fmt.Errorf("failed to create request: %v", err)
	}

	// Add authentication header
	req.Header.Set("Authorization", "Bearer "+input.Token)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return Repository{}, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return Repository{}, fmt.Errorf("API request failed: %s (status: %d)", string(body), resp.StatusCode)
	}

	// Read and parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Repository{}, fmt.Errorf("failed to read response body: %v", err)
	}

	var repo Repository
	if err := json.Unmarshal(body, &repo); err != nil {
		return Repository{}, fmt.Errorf("failed to parse response: %v", err)
	}

	return repo, nil
}

// parseGitHubURL extracts owner and repository name from a GitHub repository URL
func parseGitHubURL(url string) (owner, repo string, err error) {
	// Handle SSH URLs (git@github.com:owner/repo.git)
	if len(url) > 15 && url[:15] == "git@github.com:" {
		path := url[15:] // Remove "git@github.com:"
		return parseOwnerAndRepo(path)
	}

	// Handle HTTPS URLs (https://github.com/owner/repo)
	if len(url) > 19 && url[:19] == "https://github.com/" {
		path := url[19:] // Remove "https://github.com/"
		return parseOwnerAndRepo(path)
	}

	return "", "", fmt.Errorf("invalid GitHub repository URL format")
}

// parseOwnerAndRepo splits the repository path into owner and repository name
func parseOwnerAndRepo(path string) (owner, repo string, err error) {
	// Remove .git suffix if present
	if len(path) > 4 && path[len(path)-4:] == ".git" {
		path = path[:len(path)-4]
	}

	// Split path into owner and repo
	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repository path format")
	}

	owner = parts[0]
	repo = parts[1]

	// Validate owner and repo are not empty
	if owner == "" || repo == "" {
		return "", "", fmt.Errorf("owner and repository name cannot be empty")
	}

	return owner, repo, nil
}
