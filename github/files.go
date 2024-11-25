package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
)

// FetchRepoFilesInput encapsulates the parameters for fetching repository files.
type FetchRepoFilesInput struct {
	Owner string // Repository owner
	Ref   string // Branch, tag, or commit reference
	Repo  string // Repository name
	Path  string // Path within the repository (use "" for root)
	Token string // GitHub personal access token
	Type  string // Type of file to fetch ("file", "dir", or "symlink")
}

// File represents the structure of a file or directory item in a GitHub repository.
type File struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Type string `json:"type"` // "file", "dir", or "symlink"
	Url  string `json:"url"`
}

// FetchRepoFiles lists all files within a specified path in a GitHub repository, handling pagination.
func FetchRepoFiles(input FetchRepoFilesInput) ([]File, error) {
	// GitHub API endpoint for repository contents
	baseURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", input.Owner, input.Repo, input.Path)

	var allFiles []File
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

		// add the ref parameter if it is provided
		if input.Ref != "" {
			q := req.URL.Query()
			q.Add("ref", input.Ref)
			req.URL.RawQuery = q.Encode()
		}

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
		var files []File
		if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
			return nil, fmt.Errorf("failed to parse response: %v", err)
		}
		allFiles = append(allFiles, files...)

		// Check for pagination
		url = ""
		if linkHeader := resp.Header.Get("Link"); linkHeader != "" {
			url = extractNextPageURL(linkHeader)
		}
	}

	sort.Slice(allFiles, func(i, j int) bool {
		return allFiles[i].Name > allFiles[j].Name
	})

	if input.Type != "" {
		files := []File{}
		for _, file := range allFiles {
			if file.Type == input.Type {
				files = append(files, file)
			}
		}
		return files, nil
	}

	return allFiles, nil
}
