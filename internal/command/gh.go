// Description: GitHub CLI commands.
//
// In the future, we might expand support for the yaml, like so:
//
// ```yml
// search:
//   github:
//     method: api
//     endpoint: search/repositories
//     query: |
//       q=org:myorg+language:go
// ```
//
// Or even support for other providers in the future:
//
// ```yml
// search:
//   gitlab:
//     method: projects
//     query: |
//       --member-username johndoe --topic kubernetes
// ```

package command

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/shlex"
)

const (
	apiRequestDelay = 100 * time.Millisecond
)

// Search the GitHub API for code.
func (e *Executor) GHAPISearchCode(query string) ([]string, error) {
	uniqueRepos := make(map[string]struct{})
	page := 1
	perPage := 30
	hasMoreResults := true

	for hasMoreResults {
		// TODO: add options to executor?
		// if e.options.Debug {
		// 	e.log.Info(fmt.Sprintf("Fetching page %d (items per page: %d)...\n", page, perPage))
		// }

		// Use GitHub API with correct query parameter format
		// https://docs.github.com/en/rest/search/search?apiVersion=2022-11-28#search-code
		result, err := e.Execute(
			"gh",
			[]string{
				"api",
				"-H", "Accept: application/vnd.github+json",
				"-H", "X-GitHub-Api-Version: 2022-11-28",
				"-X",
				"GET",
				"search/code",
				"--raw-field",
				fmt.Sprintf("q=%s", query),
				"--field",
				fmt.Sprintf("per_page=%d", perPage),
				"--field",
				fmt.Sprintf("page=%d", page),
			},
		)
		if err != nil {
			// Continue processing even if there's an error, as we might have partial results
			e.log.Warn(fmt.Sprintf("Warning: API search error (page %d): %v\n", page, err))
			hasMoreResults = false
			continue
		}

		// Parse the JSON response
		//
		// {
		//   "total_count": 42,
		//   "incomplete_results": false,
		//   "items": [
		//		 {
		//		   "repository": {
		//			 "full_name": "fredrikaverpil/multipr"
		//       ...
		//		   }
		//		 },
		//		 ...
		//   ]
		// }
		var response struct {
			TotalCount        int  `json:"total_count"`
			IncompleteResults bool `json:"incomplete_results"`
			Items             []struct {
				Repository struct {
					FullName string `json:"full_name"`
				} `json:"repository"`
			} `json:"items"`
		}

		if unmarshalErr := json.Unmarshal([]byte(result.Stdout), &response); unmarshalErr != nil {
			return nil, fmt.Errorf("failed to parse API response: %w", unmarshalErr)
		}

		// Add repositories to the unique set
		for _, item := range response.Items {
			uniqueRepos[item.Repository.FullName] = struct{}{}
		}

		// TODO: add options to executor?
		// if m.options.Debug {
		// 	e.log.Info(fmt.Sprintf("Found %d items on page %d (total count: %d)",
		// 		len(response.Items), page, response.TotalCount))
		// }

		// Check if we should continue pagination
		if len(response.Items) < perPage || response.IncompleteResults {
			hasMoreResults = false
		} else {
			page++
		}

		// GitHub has rate limits, so add a small delay to avoid hitting them
		time.Sleep(apiRequestDelay)
	}

	// Convert the unique set to a slice
	fullNames := make([]string, 0, len(uniqueRepos))
	for fullName := range uniqueRepos {
		fullNames = append(fullNames, fullName)
	}

	return fullNames, nil
}

func (e *Executor) GHSearchCode(query string, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 1000 // Default to max allowed by GitHub CLI
	}

	// Parse the query string as shell arguments
	queryArgs, err := shlex.Split(query)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}

	// Build arguments for gh search code command
	args := []string{
		"search",
		"code",
		"--json", "repository",
		"--limit", strconv.Itoa(limit),
	}
	args = append(args, queryArgs...)

	e.log.Debug(fmt.Sprintf("Executing: gh %s", strings.Join(args, " ")))

	// Execute the search command
	result, err := e.Execute("gh", args)
	if err != nil {
		return nil, fmt.Errorf("failed to search code: %w", err)
	}

	// Parse the JSON response
	var response []struct {
		Repository struct {
			FullName string `json:"nameWithOwner"`
		} `json:"repository"`
	}

	if jsonErr := json.Unmarshal([]byte(result.Stdout), &response); jsonErr != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", jsonErr)
	}

	// Extract unique repository names
	uniqueRepos := make(map[string]struct{})
	for _, item := range response {
		uniqueRepos[item.Repository.FullName] = struct{}{}
	}

	// Convert to slice
	fullNames := make([]string, 0, len(uniqueRepos))
	for fullName := range uniqueRepos {
		fullNames = append(fullNames, fullName)
	}

	return fullNames, nil
}

// GHSearchRepos searches for repositories using the GitHub CLI.
func (e *Executor) GHSearchRepos(query string, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 1000 // Default to max allowed by GitHub CLI
	}

	// Parse the query string as shell arguments
	queryArgs, err := shlex.Split(query)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}

	// Build arguments for gh search repos command
	args := []string{
		"search",
		"repos",
		"--json", "fullName",
		"--limit", strconv.Itoa(limit),
	}
	args = append(args, queryArgs...)

	e.log.Debug(fmt.Sprintf("Executing: gh %s", strings.Join(args, " ")))

	// Execute the search command
	result, err := e.Execute("gh", args)
	if err != nil {
		return nil, fmt.Errorf("failed to search repositories: %w", err)
	}

	// Parse the JSON response
	var response []struct {
		FullName string `json:"fullName"`
	}

	if jsonErr := json.Unmarshal([]byte(result.Stdout), &response); jsonErr != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", jsonErr)
	}

	// Extract repository names
	fullNames := make([]string, 0, len(response))
	for _, item := range response {
		fullNames = append(fullNames, item.FullName)
	}

	return fullNames, nil
}

// Clone repo using gh, expects repo to be in the format "owner/repo".
func (e *Executor) GHClone(repo, path string) error {
	_, err := e.Execute("gh", []string{"repo", "clone", repo, path})
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}
	return nil
}
