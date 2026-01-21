package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Intent types
const (
	IntentCategory = "category"
	IntentSource   = "source"
	IntentSearch   = "search"
	IntentNearby   = "nearby"
	IntentScore    = "score"
)

type Client struct {
	apiKey string
	model  string
	client *http.Client
}

type ExtractionResult struct {
	Intent   string   `json:"intent"`
	Entities []string `json:"entities"`
	Query    string   `json:"query"`
}

type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func NewClient(apiKey, model string) *Client {
	return &Client{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// ExtractIntentAndEntities extracts intent and entities from a natural language query
func (c *Client) ExtractIntentAndEntities(query string) (*ExtractionResult, error) {
	if c.apiKey == "" {
		// Fallback to heuristic extraction
		return c.fallbackExtraction(query)
	}

	prompt := fmt.Sprintf(`Analyze the following news query and extract:
1. Intent: one of [category, source, search, nearby, score]
2. Entities: list of relevant people, organizations, locations, or events
3. The main search query

Query: %s

Respond in JSON format:
{
  "intent": "<intent_type>",
  "entities": ["entity1", "entity2"],
  "query": "<extracted_query>"
}

Intent guidelines:
- "category" if asking about a specific news category (technology, sports, etc.)
- "source" if asking about a specific news source or publication
- "nearby" if asking about news near a location
- "score" if asking about high-quality or important news
- "search" for general keyword searches`, query)

	reqBody := OpenAIRequest{
		Model: c.model,
		Messages: []Message{
			{Role: "system", Content: "You are a news query analyzer. Always respond with valid JSON."},
			{Role: "user", Content: prompt},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return c.fallbackExtraction(query)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return c.fallbackExtraction(query)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return c.fallbackExtraction(query)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.fallbackExtraction(query)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.fallbackExtraction(query)
	}

	var openAIResp OpenAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return c.fallbackExtraction(query)
	}

	if len(openAIResp.Choices) == 0 {
		return c.fallbackExtraction(query)
	}

	content := openAIResp.Choices[0].Message.Content
	
	// Try to extract JSON from the response
	var result ExtractionResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		// Try to find JSON in markdown code blocks
		if start := strings.Index(content, "```json"); start != -1 {
			start += 7
			if end := strings.Index(content[start:], "```"); end != -1 {
				jsonStr := content[start : start+end]
				if err := json.Unmarshal([]byte(jsonStr), &result); err == nil {
					return &result, nil
				}
			}
		}
		return c.fallbackExtraction(query)
	}

	return &result, nil
}

// fallbackExtraction provides heuristic extraction when LLM is not available
func (c *Client) fallbackExtraction(query string) (*ExtractionResult, error) {
	lowerQuery := strings.ToLower(query)
	
	result := &ExtractionResult{
		Intent:   IntentSearch,
		Entities: extractEntities(query),
		Query:    query,
	}

	// Detect intent based on keywords
	if strings.Contains(lowerQuery, "near") || strings.Contains(lowerQuery, "nearby") || 
	   strings.Contains(lowerQuery, "around") || strings.Contains(lowerQuery, "location") {
		result.Intent = IntentNearby
	} else if strings.Contains(lowerQuery, "category:") || 
	          containsCategory(lowerQuery) {
		result.Intent = IntentCategory
	} else if strings.Contains(lowerQuery, "source:") || 
	          strings.Contains(lowerQuery, "from ") {
		result.Intent = IntentSource
	} else if strings.Contains(lowerQuery, "important") || 
	          strings.Contains(lowerQuery, "high quality") || 
	          strings.Contains(lowerQuery, "top news") {
		result.Intent = IntentScore
	}

	return result, nil
}

// extractEntities extracts potential entities from the query
func extractEntities(query string) []string {
	// Simple entity extraction: capitalize words, known entities
	words := strings.Fields(query)
	entities := []string{}
	
	for _, word := range words {
		// Skip common words - check if first letter is uppercase
		if len(word) > 3 && word[0] >= 'A' && word[0] <= 'Z' {
			entities = append(entities, word)
		}
	}
	
	return entities
}

// containsCategory checks if query contains a news category
func containsCategory(query string) bool {
	categories := []string{
		"technology", "tech", "sports", "business", "entertainment", 
		"science", "health", "politics", "world", "national",
	}
	
	for _, cat := range categories {
		if strings.Contains(query, cat) {
			return true
		}
	}
	return false
}

// GenerateSummary generates a summary for an article
func (c *Client) GenerateSummary(title, description string) (string, error) {
	if c.apiKey == "" {
		// Fallback to a simple summary
		return c.fallbackSummary(title, description), nil
	}

	prompt := fmt.Sprintf(`Summarize the following news article in 1-2 concise sentences:

Title: %s
Description: %s

Summary:`, title, description)

	reqBody := OpenAIRequest{
		Model: c.model,
		Messages: []Message{
			{Role: "system", Content: "You are a news summarizer. Provide concise 1-2 sentence summaries."},
			{Role: "user", Content: prompt},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return c.fallbackSummary(title, description), nil
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return c.fallbackSummary(title, description), nil
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return c.fallbackSummary(title, description), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.fallbackSummary(title, description), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.fallbackSummary(title, description), nil
	}

	var openAIResp OpenAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return c.fallbackSummary(title, description), nil
	}

	if len(openAIResp.Choices) == 0 {
		return c.fallbackSummary(title, description), nil
	}

	summary := strings.TrimSpace(openAIResp.Choices[0].Message.Content)
	return summary, nil
}

// fallbackSummary provides a simple summary when LLM is not available
func (c *Client) fallbackSummary(title, description string) string {
	// Truncate description to first 150 characters and add title context
	summary := description
	if len(summary) > 150 {
		summary = summary[:150] + "..."
	}
	return fmt.Sprintf("This article about '%s' reports that %s", title, strings.ToLower(summary))
}
