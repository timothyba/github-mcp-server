package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/google/go-github/v76/github"
	"github.com/mark3labs/mcp-go/mcp"
)

var (
	// Cache for compiled regex patterns to avoid recompilation
	regexCache      = make(map[string]*regexp.Regexp)
	regexCacheMutex sync.RWMutex
)

// getCompiledRegex returns a cached compiled regex or compiles and caches a new one
func getCompiledRegex(pattern string) (*regexp.Regexp, error) {
	regexCacheMutex.RLock()
	re, exists := regexCache[pattern]
	regexCacheMutex.RUnlock()
	
	if exists {
		return re, nil
	}
	
	regexCacheMutex.Lock()
	defer regexCacheMutex.Unlock()
	
	// Double-check after acquiring write lock
	if re, exists := regexCache[pattern]; exists {
		return re, nil
	}
	
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	
	regexCache[pattern] = re
	return re, nil
}

func hasFilter(query, filterType string) bool {
	// Quick check: if the filter prefix doesn't exist at all, return false
	prefix := filterType + ":"
	if !strings.Contains(query, prefix) {
		return false
	}
	
	// Match filter at start of string, after whitespace, or after non-word characters like '('
	pattern := fmt.Sprintf(`(^|\s|\W)%s:\S+`, regexp.QuoteMeta(filterType))
	re, err := getCompiledRegex(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(query)
}

func hasSpecificFilter(query, filterType, filterValue string) bool {
	// Quick check: if the filter prefix doesn't exist at all, return false
	prefix := filterType + ":" + filterValue
	if !strings.Contains(query, prefix) {
		return false
	}
	
	// Match specific filter:value at start, after whitespace, or after non-word characters
	// End with word boundary, whitespace, or non-word characters like ')'
	pattern := fmt.Sprintf(`(^|\s|\W)%s:%s($|\s|\W)`, regexp.QuoteMeta(filterType), regexp.QuoteMeta(filterValue))
	re, err := getCompiledRegex(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(query)
}

func hasRepoFilter(query string) bool {
	return hasFilter(query, "repo")
}

func hasTypeFilter(query string) bool {
	return hasFilter(query, "type")
}

func searchHandler(
	ctx context.Context,
	getClient GetClientFn,
	request mcp.CallToolRequest,
	searchType string,
	errorPrefix string,
) (*mcp.CallToolResult, error) {
	query, err := RequiredParam[string](request, "query")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if !hasSpecificFilter(query, "is", searchType) {
		query = fmt.Sprintf("is:%s %s", searchType, query)
	}

	owner, err := OptionalParam[string](request, "owner")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	repo, err := OptionalParam[string](request, "repo")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if owner != "" && repo != "" && !hasRepoFilter(query) {
		query = fmt.Sprintf("repo:%s/%s %s", owner, repo, query)
	}

	sort, err := OptionalParam[string](request, "sort")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	order, err := OptionalParam[string](request, "order")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	pagination, err := OptionalPaginationParams(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	opts := &github.SearchOptions{
		// Default to "created" if no sort is provided, as it's a common use case.
		Sort:  sort,
		Order: order,
		ListOptions: github.ListOptions{
			Page:    pagination.Page,
			PerPage: pagination.PerPage,
		},
	}

	client, err := getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get GitHub client: %w", errorPrefix, err)
	}
	result, resp, err := client.Search.Issues(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errorPrefix, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to read response body: %w", errorPrefix, err)
		}
		return mcp.NewToolResultError(fmt.Sprintf("%s: %s", errorPrefix, string(body))), nil
	}

	r, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal response: %w", errorPrefix, err)
	}

	return mcp.NewToolResultText(string(r)), nil
}
