# API Examples and Testing

This document provides comprehensive examples for testing all endpoints of the Contextual News Data Retrieval System.

## Starting the Server

```bash
# Start the server (default port 8080)
go run ./cmd/server

# Or with custom port
PORT=8085 go run ./cmd/server
```

## Quick Test Script

Use the provided test script to validate all endpoints:

```bash
./test_endpoints.sh
```

## Detailed Examples

### 1. Health Check

```bash
curl http://localhost:8080/health
```

**Response:**
```json
{"status":"ok"}
```

### 2. Category Endpoint

Get news articles from a specific category, ranked by publication date.

```bash
curl "http://localhost:8080/api/v1/news/category?category=technology&limit=5"
```

**Parameters:**
- `category`: Category name (e.g., technology, sports, business, entertainment)
- `limit`: Number of results (default: 5, max: 100)

### 3. Source Endpoint

Get news articles from a specific source, ranked by publication date.

```bash
curl "http://localhost:8080/api/v1/news/source?source=Reuters&limit=5"
```

**Parameters:**
- `source`: Source name (e.g., Reuters, BBC, New York Times)
- `limit`: Number of results (default: 5, max: 100)

### 4. Score Endpoint

Get articles with relevance score above a threshold, ranked by relevance score.

```bash
curl "http://localhost:8080/api/v1/news/score?min=0.7&limit=5"
```

**Parameters:**
- `min`: Minimum relevance score (0.0 to 1.0)
- `limit`: Number of results (default: 5, max: 100)

### 5. Search Endpoint

Search articles by keywords with combined ranking (40% relevance + 60% text match).

```bash
curl "http://localhost:8080/api/v1/news/search?query=cricket%20match&limit=5"
```

**Parameters:**
- `query`: Search keywords
- `limit`: Number of results (default: 5, max: 100)

### 6. Nearby Endpoint

Get news near a geographic location within a radius, ranked by distance.

```bash
# Mumbai, India
curl "http://localhost:8080/api/v1/news/nearby?lat=19.0760&lon=72.8777&radius=50&limit=5"

# San Francisco, USA
curl "http://localhost:8080/api/v1/news/nearby?lat=37.7749&lon=-122.4194&radius=100&limit=5"
```

**Parameters:**
- `lat`: Latitude
- `lon`: Longitude
- `radius`: Search radius in kilometers
- `limit`: Number of results (default: 5, max: 100)

### 7. Trending Endpoint

Get trending news based on simulated user activity, with location-based ranking.

```bash
curl "http://localhost:8080/api/v1/news/trending?lat=19.0760&lon=72.8777&limit=5"
```

**Parameters:**
- `lat`: User latitude
- `lon`: User longitude
- `limit`: Number of results (default: 5, max: 100)

**Features:**
- Considers user interaction events (views and clicks)
- Applies recency decay to favor recent activity
- Weights geographical proximity to user location
- Caches results by location clusters (TTL: 5 minutes by default)

### 8. Query Endpoint (LLM-Powered)

Use natural language to query news. The system extracts intent and dispatches to appropriate endpoint.

```bash
# Technology category intent
curl "http://localhost:8080/api/v1/news/query?query=Show%20me%20technology%20news&limit=5"

# Search intent
curl "http://localhost:8080/api/v1/news/query?query=Latest%20cricket%20match%20results&limit=5"

# Source intent
curl "http://localhost:8080/api/v1/news/query?query=News%20from%20Reuters&limit=5"

# Score/quality intent
curl "http://localhost:8080/api/v1/news/query?query=Show%20me%20the%20best%20quality%20news&limit=5"

# Nearby intent (with location parameters)
curl "http://localhost:8080/api/v1/news/query?query=News%20near%20me&lat=19.0760&lon=72.8777&limit=5"
```

**Parameters:**
- `query`: Natural language query
- `lat`: (Optional) User latitude for nearby queries
- `lon`: (Optional) User longitude for nearby queries
- `limit`: Number of results (default: 5, max: 100)

**Intent Detection:**
Without OpenAI API key, the system uses heuristic fallback to detect intent:
- **Category Intent**: Keywords like "technology", "sports", "business", "entertainment"
- **Source Intent**: Keywords like "times", "post", "bbc", "cnn", "reuters"
- **Score Intent**: Keywords like "best", "top", "important", "quality"
- **Nearby Intent**: Keywords like "near", "nearby", "around", "location"
- **Search Intent**: Default for all other queries

## Response Format

All endpoints return a consistent JSON structure:

```json
{
  "articles": [
    {
      "id": "uuid",
      "title": "Article Title",
      "description": "Article description...",
      "url": "https://example.com/article",
      "publication_date": "2025-03-24T11:08:11Z",
      "source_name": "Source Name",
      "category": ["category1", "category2"],
      "relevance_score": 0.85,
      "latitude": 19.0760,
      "longitude": 72.8777,
      "llm_summary": "AI-generated summary..."
    }
  ],
  "meta": {
    "count": 5,
    "limit": 5,
    "endpoint": "search",
    "query": "cricket match"
  }
}
```

## Testing with Different Scenarios

### Example 1: Finding Technology News from a Specific Source

```bash
# First, search by category
curl "http://localhost:8080/api/v1/news/category?category=technology&limit=10"

# Then filter by source
curl "http://localhost:8080/api/v1/news/source?source=Reuters&limit=10"
```

### Example 2: High-Quality Local News

```bash
# Get high relevance score articles
curl "http://localhost:8080/api/v1/news/score?min=0.8&limit=20"

# Then filter by location
curl "http://localhost:8080/api/v1/news/nearby?lat=19.0760&lon=72.8777&radius=100&limit=10"
```

### Example 3: Trending Topics in Your Area

```bash
# Get trending news (considers simulated user activity)
curl "http://localhost:8080/api/v1/news/trending?lat=37.7749&lon=-122.4194&limit=10"
```

### Example 4: Natural Language Queries

```bash
# Business news
curl "http://localhost:8080/api/v1/news/query?query=Latest%20business%20developments&limit=5"

# Sports updates
curl "http://localhost:8080/api/v1/news/query?query=Cricket%20tournament%20updates&limit=5"

# Location-based query
curl "http://localhost:8080/api/v1/news/query?query=News%20happening%20nearby&lat=19.0760&lon=72.8777&limit=5"
```

## Error Responses

The API returns appropriate HTTP status codes and error messages:

```json
{
  "error": "category parameter is required"
}
```

**Common Status Codes:**
- `200`: Success
- `400`: Bad Request (missing or invalid parameters)
- `500`: Internal Server Error (database or processing error)

## Performance Notes

1. **Search Endpoint**: May be slower with large datasets as it loads all articles for ranking
2. **Trending Endpoint**: Cached by location clusters (default 5-minute TTL)
3. **LLM Summaries**: Cached in database after first generation
4. **Event Simulation**: Runs every 5 seconds in the background

## Advanced Usage

### Custom Configuration

Create a `.env` file:

```env
DATABASE_URL=custom_news.db
OPENAI_API_KEY=sk-your-api-key-here
LLM_MODEL=gpt-4
TRENDING_CACHE_TTL=600
LOCATION_CLUSTER_DEGREES=1.0
PORT=3000
```

### With OpenAI API Key

When you provide an OpenAI API key, the system uses actual LLM for:
- More accurate intent detection
- Better entity extraction
- Higher quality article summaries

Example with better intent extraction:
```bash
export OPENAI_API_KEY=sk-your-key-here
go run ./cmd/server

curl "http://localhost:8080/api/v1/news/query?query=I%20want%20to%20read%20about%20the%20latest%20developments%20in%20artificial%20intelligence%20from%20reputable%20tech%20publications&limit=5"
```

## Monitoring and Debugging

Check server logs for:
- API request logs
- Database operations
- Event simulation activity
- Cache hits/misses

Enable verbose logging by setting Gin to debug mode (default).

## Data Import

Before using the API, import your news data:

```bash
go run import_data.go "news_data (1).json"
```

This will:
- Parse the JSON file
- Validate article structure
- Insert/update articles in the database
- Display progress every 1000 articles
