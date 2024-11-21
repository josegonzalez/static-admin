package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// FileInfo represents a file or folder in a GitHub repository.
type FileInfo struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Type string `json:"type"` // "file" or "dir"
}

// ListGitHubFilesAndFolders lists the files and folders in a GitHub repository path.
func ListGitHubFilesAndFolders(accessToken, owner, repo, path, branch string) ([]FileInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s", owner, repo, path, branch)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch files and folders: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	var files []FileInfo
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, fmt.Errorf("failed to decode GitHub response: %w", err)
	}

	return files, nil
}

// GetGitHubFileContent retrieves the content of a file from a GitHub repository.
func GetGitHubFileContent(accessToken, owner, repo, path, branch string) (string, bool, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s", owner, repo, path, branch)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", false, fmt.Errorf("failed to fetch file content: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", false, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	var fileData struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
		Type     string `json:"type"` // "file" or "dir"
	}

	if err := json.NewDecoder(resp.Body).Decode(&fileData); err != nil {
		return "", false, fmt.Errorf("failed to decode GitHub response: %w", err)
	}

	if fileData.Type != "file" {
		return "", false, fmt.Errorf("path is not a file")
	}

	if fileData.Encoding != "base64" {
		return "", false, fmt.Errorf("unsupported encoding: %s", fileData.Encoding)
	}

	decodedContent, err := base64.StdEncoding.DecodeString(fileData.Content)
	if err != nil {
		return "", false, fmt.Errorf("failed to decode file content: %w", err)
	}

	// Detect if the file is binary
	isBinary := isBinaryContent(decodedContent)
	return string(decodedContent), isBinary, nil
}

// isBinaryContent checks if the content is likely binary
func isBinaryContent(content []byte) bool {
	for _, b := range content {
		if b == 0 {
			return true
		}
	}
	return false
}

// CreateOrUpdateGitHubFile creates or updates a file on GitHub.
func CreateOrUpdateGitHubFile(accessToken, owner, repo, path, content, branch, commitMsg string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, path)
	payload := fmt.Sprintf(`{
		"message": "%s",
		"content": "%s",
		"branch": "%s"
	}`, commitMsg, content, branch)

	req, _ := http.NewRequest("PUT", url, strings.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode >= 300 {
		return fmt.Errorf("failed to update GitHub file: %v", err)
	}
	return nil
}
