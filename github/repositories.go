package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
)

// Repository represents the structure of a repository from the GitHub API response
type Repository struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	Private       bool   `json:"private"`
	Url           string `json:"url"`
	DefaultBranch string `json:"default_branch"`
}

// FetchOrgRepositoriesInput represents the input parameters for the FetchOrgRepositories function
type FetchOrgRepositoriesInput struct {
	// Organization is the name of the organization for which to fetch repositories
	Organization string

	// Token is the GitHub personal access token to use for authentication
	Token string
}

// FetchOrgRepositories fetches all repositories within a given organization, handling pagination
func FetchOrgRepositories(input FetchOrgRepositoriesInput) ([]Repository, error) {
	if input.Organization == "" {
		return nil, fmt.Errorf("organization name is required")
	}

	baseURL := fmt.Sprintf("https://api.github.com/orgs/%s/repos", input.Organization)

	var allRepos []Repository
	url := baseURL

	for url != "" {
		// Create a new HTTP request
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %v", err)
		}

		// Add authentication header if a token is provided
		if input.Token == "" {
			return nil, fmt.Errorf("authentication token is required")
		}
		req.Header.Set("Authorization", "Bearer "+input.Token)

		q := req.URL.Query()
		q.Add("per_page", "100")
		req.URL.RawQuery = q.Encode()

		// Send the HTTP request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to send request: %v", err)
		}
		defer resp.Body.Close()

		// Check for HTTP errors
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("API request failed: %s (status: %d)", string(body), resp.StatusCode)
		}

		// Parse the JSON response
		var repos []Repository
		if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
			return nil, fmt.Errorf("failed to parse response: %v", err)
		}
		allRepos = append(allRepos, repos...)

		// Check for pagination
		url = ""
		if linkHeader := resp.Header.Get("Link"); linkHeader != "" {
			url = extractNextPageURL(linkHeader)
		}
	}

	sort.Slice(allRepos, func(i, j int) bool {
		return allRepos[i].Name < allRepos[j].Name
	})

	return allRepos, nil
}

// FetchUserRepositories fetches all repositories for the authenticated user, handling pagination.
func FetchUserRepositories(token string) ([]Repository, error) {
	// Initial API endpoint
	baseURL := "https://api.github.com/user/repos"

	var allRepos []Repository
	url := baseURL

	for url != "" {
		// Create a new HTTP request
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %v", err)
		}

		// Add authentication header
		if token == "" {
			return nil, fmt.Errorf("authentication token is required")
		}
		req.Header.Set("Authorization", "Bearer "+token)

		// add per_page=100 to the query string
		q := req.URL.Query()
		q.Add("per_page", "100")
		req.URL.RawQuery = q.Encode()

		// Send the HTTP request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to send request: %v", err)
		}
		defer resp.Body.Close()

		// Check for HTTP errors
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("API request failed: %s (status: %d)", string(body), resp.StatusCode)
		}

		// Parse the JSON response
		var repos []Repository
		if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
			return nil, fmt.Errorf("failed to parse response: %v", err)
		}
		allRepos = append(allRepos, repos...)

		// Check for pagination
		url = ""
		if linkHeader := resp.Header.Get("Link"); linkHeader != "" {
			url = extractNextPageURL(linkHeader)
		}
	}

	sort.Slice(allRepos, func(i, j int) bool {
		return allRepos[i].Name < allRepos[j].Name
	})

	return allRepos, nil
}
