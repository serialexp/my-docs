// ABOUTME: HTTP client for grep.app search API.
// ABOUTME: Handles URL construction, requests, and response parsing.

package grepapp

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

const baseURL = "https://grep.app/api/search"

type Response struct {
	Time   int    `json:"time"`
	Facets Facets `json:"facets"`
	Hits   Hits   `json:"hits"`
}

type Facets struct {
	Path FacetGroup `json:"path"`
	Repo FacetGroup `json:"repo"`
	Lang FacetGroup `json:"lang"`
}

type FacetGroup struct {
	Buckets []Bucket `json:"buckets"`
}

type Bucket struct {
	Val   string `json:"val"`
	Count int    `json:"count"`
}

type Hits struct {
	Total int   `json:"total"`
	Hits  []Hit `json:"hits"`
}

type Hit struct {
	Repo    string  `json:"repo"`
	Branch  string  `json:"branch"`
	Path    string  `json:"path"`
	Content Content `json:"content"`
}

type Content struct {
	Snippet string `json:"snippet"`
}

type Match struct {
	Line int
	Text string
}

func BuildURL(query, repo string) string {
	params := url.Values{}
	params.Set("q", query)
	if repo != "" {
		params.Set("f.repo", repo)
	}
	return baseURL + "?" + params.Encode()
}

func Search(query, repo string) (*Response, error) {
	searchURL := BuildURL(query, repo)

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "my-docs/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("grep.app returned status %d", resp.StatusCode)
	}

	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

var lineRegex = regexp.MustCompile(`data-line="(\d+)"`)
var tagRegex = regexp.MustCompile(`<[^>]+>`)

func ExtractText(snippet string) []Match {
	var matches []Match

	rows := strings.Split(snippet, "</tr>")
	for _, row := range rows {
		lineMatch := lineRegex.FindStringSubmatch(row)
		if lineMatch == nil {
			continue
		}
		lineNum, _ := strconv.Atoi(lineMatch[1])

		preStart := strings.Index(row, "<pre>")
		preEnd := strings.Index(row, "</pre>")
		if preStart == -1 || preEnd == -1 {
			continue
		}

		content := row[preStart+5 : preEnd]
		text := tagRegex.ReplaceAllString(content, "")
		text = html.UnescapeString(text)
		text = strings.TrimSpace(text)

		matches = append(matches, Match{Line: lineNum, Text: text})
	}

	return matches
}
