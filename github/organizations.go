package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
)

// Organization represents the structure of an organization from the GitHub API response
type Organization struct {
	// Login is the organization's username
	Login string `json:"login"`

	// ID is the organization's unique identifier
	Description string `json:"description"`

	// Url is the URL to the organization's profile
	Url string `json:"url"`
}

// FetchOrganizationsInput represents the input parameters for the FetchOrganizations function
type FetchOrganizationsInput struct {
	// Username is the GitHub username for which to fetch organizations
	Username string

	// Token is the GitHub personal access token to use for authentication
	Token string
}

// FetchOrganizations fetches all the organizations for a given GitHub user
func FetchOrganizations(input FetchOrganizationsInput) ([]Organization, error) {
	if input.Username == "" {
		return nil, fmt.Errorf("username is required")
	}

	baseURL := fmt.Sprintf("https://api.github.com/users/%s/orgs", input.Username)

	var allOrgs []Organization
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
		var orgs []Organization
		if err := json.NewDecoder(resp.Body).Decode(&orgs); err != nil {
			return nil, fmt.Errorf("failed to parse response: %v", err)
		}
		allOrgs = append(allOrgs, orgs...)

		// Check for pagination
		url = ""
		if linkHeader := resp.Header.Get("Link"); linkHeader != "" {
			// Parse the Link header to find the URL for the next page
			url = extractNextPageURL(linkHeader)
		}
	}

	sort.Slice(allOrgs, func(i, j int) bool {
		return allOrgs[i].Login < allOrgs[j].Login
	})

	return allOrgs, nil
}

// extractNextPageURL parses the Link header to extract the "next" page URL.
func extractNextPageURL(linkHeader string) string {
	// Split the header by commas, as each part represents a link
	links := strings.Split(linkHeader, ",")
	for _, link := range links {
		// Find the part that contains rel="next"
		parts := strings.Split(strings.TrimSpace(link), ";")
		if len(parts) == 2 && strings.TrimSpace(parts[1]) == `rel="next"` {
			// Extract the URL from the angle brackets
			url := strings.Trim(parts[0], "<>")
			return url
		}
	}
	return ""
}
