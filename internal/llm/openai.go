package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type IntentType string

const (
	IntentCategory IntentType = "category"
	IntentSource   IntentType = "source"
	IntentSearch   IntentType = "search"
	IntentNearby   IntentType = "nearby"
	IntentScore    IntentType = "score"
)

type ExtractionResult struct {
	Intent   IntentType         `json:"intent"`
	Entities map[string]string  `json:"entities"`
}

type Client struct {
	client   *openai.Client
	model    string
	hasKey   bool
}

func NewClient(apiKey, model string) *Client {
	var client *openai.Client
	hasKey := apiKey != ""
	
	if hasKey {
		client = openai.NewClient(apiKey)
	}
	
	return &Client{
		client: client,
		model:  model,
		hasKey: hasKey,
	}
}

// ExtractEntitiesAndIntent extracts entities and determines intent from query
func (c *Client) ExtractEntitiesAndIntent(ctx context.Context, query string) (*ExtractionResult, error) {
	if !c.hasKey {
		return c.fallbackExtraction(query), nil
	}
	
	prompt := fmt.Sprintf(`Analyze this news query and extract:
1. Intent (one of: category, source, search, nearby, score)
2. Relevant entities (people, organizations, locations, events, categories, sources)

Query: "%s"

Respond with JSON only in this format:
{
  "intent": "category|source|search|nearby|score",
  "entities": {
    "category": "value if present",
    "source": "value if present",
    "location": "value if present",
    "keywords": "key search terms if present"
  }
}

Guidelines:
- "category" intent: user wants news from a specific category (technology, sports, business, etc.)
- "source" intent: user wants news from a specific source/publication
- "search" intent: user wants to search for specific keywords, people, events
- "nearby" intent: user mentions a specific location or "near me" or "nearby"
- "score" intent: user asks for high-quality or highly relevant news`, query)

	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0.3,
	})
	
	if err != nil {
		return c.fallbackExtraction(query), nil
	}
	
	if len(resp.Choices) == 0 {
		return c.fallbackExtraction(query), nil
	}
	
	content := resp.Choices[0].Message.Content
	
	// Try to parse JSON response
	var result ExtractionResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		// Try to extract JSON from markdown code blocks
		if start := strings.Index(content, "{"); start != -1 {
			if end := strings.LastIndex(content, "}"); end != -1 {
				jsonStr := content[start : end+1]
				if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
					return c.fallbackExtraction(query), nil
				}
			}
		} else {
			return c.fallbackExtraction(query), nil
		}
	}
	
	return &result, nil
}

// Summarize generates a summary for an article
func (c *Client) Summarize(ctx context.Context, title, description string) (string, error) {
	if !c.hasKey {
		return c.fallbackSummary(title, description), nil
	}
	
	prompt := fmt.Sprintf(`Summarize this news article in 1-2 concise sentences:

Title: %s
Description: %s

Provide only the summary, no additional text.`, title, description)

	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0.5,
		MaxTokens:   100,
	})
	
	if err != nil {
		return c.fallbackSummary(title, description), nil
	}
	
	if len(resp.Choices) == 0 {
		return c.fallbackSummary(title, description), nil
	}
	
	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

// fallbackExtraction provides heuristic-based extraction when no API key
func (c *Client) fallbackExtraction(query string) *ExtractionResult {
	queryLower := strings.ToLower(query)
	entities := make(map[string]string)
	
	// Check for location indicators
	if strings.Contains(queryLower, "near") || strings.Contains(queryLower, "nearby") ||
		strings.Contains(queryLower, "around") || strings.Contains(queryLower, "in ") {
		return &ExtractionResult{
			Intent:   IntentNearby,
			Entities: entities,
		}
	}
	
	// Check for category keywords
	categories := []string{"technology", "tech", "sports", "business", "entertainment",
		"politics", "science", "health", "world", "national"}
	for _, cat := range categories {
		if strings.Contains(queryLower, cat) {
			entities["category"] = cat
			return &ExtractionResult{
				Intent:   IntentCategory,
				Entities: entities,
			}
		}
	}
	
	// Check for source indicators
	sources := []string{"times", "post", "news", "bbc", "cnn", "reuters", "associated press"}
	for _, src := range sources {
		if strings.Contains(queryLower, src) {
			entities["source"] = src
			return &ExtractionResult{
				Intent:   IntentSource,
				Entities: entities,
			}
		}
	}
	
	// Check for quality/score indicators
	if strings.Contains(queryLower, "best") || strings.Contains(queryLower, "top") ||
		strings.Contains(queryLower, "important") || strings.Contains(queryLower, "quality") {
		return &ExtractionResult{
			Intent:   IntentScore,
			Entities: entities,
		}
	}
	
	// Default to search
	entities["keywords"] = query
	return &ExtractionResult{
		Intent:   IntentSearch,
		Entities: entities,
	}
}

// fallbackSummary provides simple summary when no API key
func (c *Client) fallbackSummary(title, description string) string {
	// If description is short enough, use it
	if len(description) <= 150 {
		return description
	}
	
	// Otherwise, create a simple truncated summary
	summary := description
	if len(summary) > 150 {
		summary = summary[:147] + "..."
	}
	
	return summary
}
