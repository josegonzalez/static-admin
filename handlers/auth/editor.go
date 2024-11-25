package auth

import (
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"

	"static-admin/config"
	"static-admin/github"
	"static-admin/markdown"
	"static-admin/middleware"
	"static-admin/session"

	"github.com/gin-gonic/gin"
)

// GitHubMarkdownRequest represents the form data for fetching a markdown file from GitHub
type GitHubMarkdownRequest struct {
	// Organization is the owner of the GitHub repository
	Organization string `form:"organization" binding:"required"`

	// Repository is the name of the GitHub repository
	Repository string `form:"repository" binding:"required"`

	// Path is the path to the markdown file in the repository
	Path string `form:"path" binding:"required"`

	// Branch is the branch of the repository
	Branch string `form:"branch" binding:"required"`
}

// NewEditorHandler creates a new handler for the editor page
func NewEditorHandler(config config.Config) (EditorHandler, error) {
	tmpl, err := template.ParseFS(config.StaticFiles, "assets/edit.html")
	if err != nil {
		return EditorHandler{}, err
	}

	return EditorHandler{
		StaticFiles: config.StaticFiles,
		GithubToken: config.GithubToken,
		Template:    tmpl,
	}, err
}

// EditorHandler is a handler for the editor page
type EditorHandler struct {
	// StaticFiles contains the embedded static files
	StaticFiles embed.FS

	// GithubToken is the GitHub personal access token
	GithubToken string

	// Template is the template for the editor page
	Template *template.Template
}

// Register registers the handler with the given router
func (h EditorHandler) AuthRegister(auth *gin.RouterGroup) {
	auth.GET("/editor", h.handler)
}

// handler handles the request for the editor page
func (h EditorHandler) handler(c *gin.Context) {
	var req GitHubMarkdownRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		context := session.PageContext(c, err)
		c.HTML(http.StatusBadRequest, "edit.html", context)
		return
	}

	user, ok := c.Get("githubUser")
	if !ok {
		// todo: better error handling
		c.String(http.StatusInternalServerError, "Failed to get GitHub user")
		return
	}

	githubUser := user.(middleware.GithubUser)
	// Create the request struct for GitHub API
	gitHubRequest := github.GitHubFileRequest{
		RepoOwner: req.Organization,
		RepoName:  req.Repository,
		FilePath:  req.Path,
		Branch:    req.Branch,
		Token:     githubUser.AccessToken,
	}

	fmt.Printf("Fetching file from GitHub: %v\n", gitHubRequest)
	fmt.Printf("RepoOwner: %v\n", gitHubRequest.RepoOwner)
	fmt.Printf("RepoName: %v\n", gitHubRequest.RepoName)
	fmt.Printf("FilePath: %v\n", gitHubRequest.FilePath)
	fmt.Printf("Branch: %v\n", gitHubRequest.Branch)

	source, err := github.FetchFileFromGitHub(gitHubRequest)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to read markdown file: %s", err.Error())
		return
	}

	frontmatter, content, err := markdown.ExtractFrontMatter([]byte(source))
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to extract frontmatter: %s", err.Error())
		return
	}

	blocks, err := markdown.ParseMarkdownToBlocks(content)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to parse markdown: %s", err.Error())
		return
	}

	c.HTML(http.StatusOK, "edit.html", gin.H{
		"Frontmatter": frontmatter,
		"Blocks":      blocks,
	})
}

// toJSONString converts the given data to a JSON string
func toJSONString(data interface{}) string {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "{}"
	}
	return string(jsonData)
}
