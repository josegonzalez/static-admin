package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type CreatePullRequestIfNecessaryInput struct {
	Owner      string
	Repo       string
	Branch     string
	BaseBranch string
	Title      string
	Body       string
	Token      string
}

type CreateBranchAndUpdateFileInput struct {
	Owner      string
	Repo       string
	Path       string
	Content    string
	Branch     string
	BaseBranch string
	CommitMsg  string
	Token      string
}

type CreateCommit struct {
	Message string   `json:"message"`
	Parents []string `json:"parents"`
	Tree    string   `json:"tree"`
}

type Commit struct {
	Tree CommitTree `json:"tree"`
}

type CommitTree struct {
	SHA string `json:"sha"`
}

type PullRequest struct {
	Number int64 `json:"number"`
}

type Ref struct {
	Ref    string    `json:"ref"`
	NodeID string    `json:"node_id"`
	URL    string    `json:"url"`
	Object RefObject `json:"object"`
}

type RefObject struct {
	Type string `json:"type"`
	SHA  string `json:"sha"`
	URL  string `json:"url"`
}

type Tree struct {
	BaseTree string       `json:"base_tree"`
	Tree     []TreeObject `json:"tree"`
}

type TreeObject struct {
	Path    string `json:"path"`
	Mode    string `json:"mode"`
	Type    string `json:"type"`
	Content string `json:"content"`
}

func getHeadRef(owner, repo, branch, token string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/ref/heads/%s", owner, repo, branch)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		println("Error creating head request:", err)
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		println("Error getting head:", err)
		return "", err
	}
	defer resp.Body.Close()

	var headData Ref
	if err := json.NewDecoder(resp.Body).Decode(&headData); err != nil {
		println("Error decoding head response:", err)
		return "", err
	}
	return headData.Object.SHA, nil
}

func getCommitTree(owner, repo, commitSHA, token string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/commits/%s", owner, repo, commitSHA)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		println("Error creating commit request:", err)
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		println("Error getting commit:", err)
		return "", err
	}
	defer resp.Body.Close()

	var commitData Commit
	if err := json.NewDecoder(resp.Body).Decode(&commitData); err != nil {
		println("Error decoding commit response:", err)
		return "", err
	}
	return commitData.Tree.SHA, nil
}

func createTree(owner, repo, baseTree, path, content, token string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/trees", owner, repo)
	treeData := Tree{
		BaseTree: baseTree,
		Tree: []TreeObject{
			{
				Path:    path,
				Mode:    "100644",
				Type:    "blob",
				Content: content,
			},
		},
	}

	body, err := json.Marshal(treeData)
	if err != nil {
		println("Error marshaling tree data:", err)
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		println("Error creating tree request:", err)
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		println("Error creating tree:", err)
		return "", err
	}
	defer resp.Body.Close()

	var newTreeData CommitTree
	if err := json.NewDecoder(resp.Body).Decode(&newTreeData); err != nil {
		println("Error decoding tree response:", err)
		return "", err
	}
	return newTreeData.SHA, nil
}

func createCommit(owner, repo, message, parentSHA, treeSHA, token string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/commits", owner, repo)
	commitData := CreateCommit{
		Message: message,
		Parents: []string{parentSHA},
		Tree:    treeSHA,
	}

	body, err := json.Marshal(commitData)
	if err != nil {
		println("Error marshaling commit data:", err)
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		println("Error creating commit request:", err)
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		println("Error creating commit:", err)
		return "", err
	}
	defer resp.Body.Close()

	var commitResp struct {
		SHA string `json:"sha"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&commitResp); err != nil {
		println("Error decoding commit response:", err)
		return "", err
	}
	return commitResp.SHA, nil
}

func updateBranchRef(owner, repo, branch, sha, token string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/refs/heads/%s", owner, repo, branch)
	refData := struct {
		SHA string `json:"sha"`
	}{
		SHA: sha,
	}

	body, err := json.Marshal(refData)
	if err != nil {
		println("Error marshaling ref data:", err)
		return err
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(body))
	if err != nil {
		println("Error creating ref request:", err)
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		println("Error updating branch ref:", err)
		return err
	}
	defer resp.Body.Close()

	return nil
}

func createRef(owner, repo, ref, sha, token string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/refs", owner, repo)
	refData := struct {
		Ref string `json:"ref"`
		SHA string `json:"sha"`
	}{
		Ref: "refs/heads/" + ref,
		SHA: sha,
	}

	body, err := json.Marshal(refData)
	if err != nil {
		println("Error marshaling ref data:", err)
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		println("Error creating ref request:", err)
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		println("Error creating ref:", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		println("Error response from GitHub:", string(body))
		return fmt.Errorf("failed to create ref: %s", string(body))
	}

	return nil
}

func CreateBranchAndUpdateFile(input CreateBranchAndUpdateFileInput) error {
	updateBranch := true
	lastCommitSHA, err := getHeadRef(input.Owner, input.Repo, input.Branch, input.Token)
	if err != nil || lastCommitSHA == "" {
		println("Using base branch as fallback")
		updateBranch = false
		lastCommitSHA, err = getHeadRef(input.Owner, input.Repo, input.BaseBranch, input.Token)
		if err != nil {
			return err
		}
	} else {
		println("Using branch as fallback")
	}

	// Get tree SHA from commit
	lastTreeSHA, err := getCommitTree(input.Owner, input.Repo, lastCommitSHA, input.Token)
	if err != nil {
		return err
	}

	// Create new tree with file
	newTreeSHA, err := createTree(input.Owner, input.Repo, lastTreeSHA, input.Path, input.Content, input.Token)
	if err != nil {
		return err
	}

	// Create new commit
	newCommitSHA, err := createCommit(input.Owner, input.Repo, input.CommitMsg, lastCommitSHA, newTreeSHA, input.Token)
	if err != nil {
		return err
	}

	if updateBranch {
		return updateBranchRef(input.Owner, input.Repo, input.Branch, newCommitSHA, input.Token)
	}

	// Create new branch ref
	return createRef(input.Owner, input.Repo, input.Branch, newCommitSHA, input.Token)
}

func checkIfPullRequestExists(owner, repo, branch, baseBranch, token string) (int64, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls", owner, repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		println("Error creating PR check request:", err)
		return 0, err
	}

	q := req.URL.Query()
	q.Add("head", fmt.Sprintf("%s:%s", owner, branch))
	q.Add("base", baseBranch)
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		println("Error checking for PR:", err)
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		println("Error decoding PR response:", err)
		return 0, err
	}

	var prs []PullRequest
	if err := json.NewDecoder(resp.Body).Decode(&prs); err != nil {
		println("Error decoding PR response:", err)
		return 0, err
	}

	for _, pr := range prs {
		return pr.Number, nil
	}

	return 0, nil
}

func CreatePullRequestIfNecessary(input CreatePullRequestIfNecessaryInput) (int64, error) {
	prNumber, err := checkIfPullRequestExists(input.Owner, input.Repo, input.Branch, input.BaseBranch, input.Token)
	if err != nil {
		return 0, err
	}

	if prNumber != 0 {
		return prNumber, nil
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls", input.Owner, input.Repo)

	// Create pull request data
	prData := struct {
		Title string `json:"title"`
		Body  string `json:"body"`
		Head  string `json:"head"`
		Base  string `json:"base"`
	}{
		Title: input.Title,
		Body:  input.Body,
		Head:  input.Branch,
		Base:  input.BaseBranch,
	}

	// Marshal the request body
	body, err := json.Marshal(prData)
	if err != nil {
		println("Error marshaling PR data:", err)
		return 0, err
	}

	// Create request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		println("Error creating PR request:", err)
		return 0, err
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+input.Token)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		println("Error sending PR request:", err)
		return 0, err
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		println("Error reading PR response:", err)
		return 0, err
	}

	// Check status code
	if resp.StatusCode >= 300 {
		println("Error creating PR. Status:", resp.StatusCode, "Body:", string(respBody))
		return 0, fmt.Errorf("failed to create PR: %s", string(respBody))
	}

	// Parse response
	var prResp struct {
		Number int64 `json:"number"`
	}
	if err := json.Unmarshal(respBody, &prResp); err != nil {
		println("Error parsing PR response:", err)
		return 0, err
	}

	return prResp.Number, nil
}
