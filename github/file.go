package github

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"unicode/utf8"
)

type GitHubAPIResponse struct {
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
}

type GitHubFileRequest struct {
	RepoOwner string // Repository owner (e.g., "gin-gonic")
	RepoName  string // Repository name (e.g., "gin")
	FilePath  string // Path to the file in the repository (e.g., "README.md")
	Branch    string // Branch or commit reference (e.g., "main")
	Token     string // GitHub personal access token (optional)
}

// FetchFileFromGitHub fetches a file's content from the GitHub API.
func FetchFileFromGitHub(req GitHubFileRequest) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s", req.RepoOwner, req.RepoName, req.FilePath, req.Branch)

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	// Add authentication header if a token is provided
	if req.Token != "" {
		httpReq.Header.Set("Authorization", fmt.Sprintf("token %s", req.Token))
	}

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch file: %s", resp.Status)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var apiResponse GitHubAPIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return "", err
	}

	if apiResponse.Encoding != "base64" {
		return "", errors.New("unsupported encoding, expected base64")
	}

	decodedContent, err := base64.StdEncoding.DecodeString(apiResponse.Content)
	if err != nil {
		return "", err
	}

	// Validate the decoded content as text
	if !isTextFile(decodedContent) {
		return "", fmt.Errorf("file is not a valid text file")
	}

	return string(decodedContent), nil
}

// isTextFile checks if the given content is a text file by ensuring it contains valid UTF-8.
func isTextFile(content []byte) bool {
	if len(content) == 0 {
		return false // Empty content is not valid
	}
	return utf8.Valid(content)
}
