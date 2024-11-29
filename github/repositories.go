package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"
	"time"
)

// Cache configuration
const cacheDuration = 5 * time.Minute

// Cache entry with expiration
type cacheEntry struct {
	data       interface{}
	expiration time.Time
}

var (
	cache     = make(map[string]cacheEntry)
	cacheLock sync.RWMutex
)

// getCached retrieves data from cache if available and not expired
func getCached(key string) (interface{}, bool) {
	cacheLock.RLock()
	defer cacheLock.RUnlock()

	entry, exists := cache[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.expiration) {
		delete(cache, key)
		return nil, false
	}

	return entry.data, true
}

// setCache stores data in cache with expiration
func setCache(key string, data interface{}) {
	cacheLock.Lock()
	defer cacheLock.Unlock()

	cache[key] = cacheEntry{
		data:       data,
		expiration: time.Now().Add(cacheDuration),
	}
}

// Repository represents the structure of a repository from the GitHub API response
type Repository struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	Private       bool   `json:"private"`
	Url           string `json:"url"`
	HtmlURL       string `json:"html_url"`
	DefaultBranch string `json:"default_branch"`
}

// FetchOrgRepositoriesInput represents the input parameters for the FetchOrgRepositories function
type FetchOrgRepositoriesInput struct {
	// Organization is the name of the organization for which to fetch repositories
	Organization string

	// Token is the GitHub personal access token to use for authentication
	Token string

	// UserID is the ID of the currently logged in user
	UserID uint
}

// Helper function to generate cache keys
func generateCacheKey(userID uint, url string) string {
	return fmt.Sprintf("user:%d:%s", userID, url)
}

// FetchOrgRepositories fetches all repositories within a given organization, handling pagination
func FetchOrgRepositories(input FetchOrgRepositoriesInput) ([]Repository, error) {
	if input.Organization == "" {
		return nil, fmt.Errorf("organization name is required")
	}

	baseURL := fmt.Sprintf("https://api.github.com/orgs/%s/repos", input.Organization)
	cacheKey := generateCacheKey(input.UserID, baseURL)

	// Check cache first
	if cached, ok := getCached(cacheKey); ok {
		return cached.([]Repository), nil
	}

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

	// Cache the result before returning
	setCache(cacheKey, allRepos)

	return allRepos, nil
}

type FetchUserRepositoriesInput struct {
	Token    string
	Username string
	UserID   uint
}

// FetchUserRepositories fetches all repositories for the authenticated user, handling pagination.
func FetchUserRepositories(input FetchUserRepositoriesInput) ([]Repository, error) {
	baseURL := "https://api.github.com/users/" + input.Username + "/repos"
	cacheKey := generateCacheKey(input.UserID, baseURL)

	// Check cache first
	if cached, ok := getCached(cacheKey); ok {
		return cached.([]Repository), nil
	}

	var allRepos []Repository
	url := baseURL

	for url != "" {
		// Create a new HTTP request
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %v", err)
		}

		// Add authentication header
		if input.Token == "" {
			return nil, fmt.Errorf("authentication token is required")
		}
		req.Header.Set("Authorization", "Bearer "+input.Token)

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

	// Cache the result before returning
	setCache(cacheKey, allRepos)

	return allRepos, nil
}

// StartCacheCleaner starts a goroutine to periodically clean expired cache entries
func StartCacheCleaner(quit chan struct{}) {
	ticker := time.NewTicker(time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				cacheLock.Lock()
				now := time.Now()
				for key, entry := range cache {
					if now.After(entry.expiration) {
						delete(cache, key)
					}
				}
				cacheLock.Unlock()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}
