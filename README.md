# Contextual News Data Retrieval System

A Go-based backend system for fetching, organizing, and enriching news articles with LLM-generated insights.

## Features

- **Multiple Retrieval Endpoints**: Category, source, relevance score, search, nearby, and trending news
- **LLM Integration**: Entity extraction, intent detection, and article summarization using OpenAI
- **Intelligent Ranking**: Different ranking strategies for different endpoints
- **Location-based Services**: Find news by proximity using Haversine distance calculation
- **Trending News**: Simulated user activity events with location-based trending calculations
- **Caching**: Efficient location-clustered caching for trending feeds
- **Graceful Fallback**: Works without OpenAI API key using heuristic fallbacks

## Prerequisites

- Go 1.21 or higher
- SQLite (included with GORM)
- OpenAI API key (optional, for enhanced LLM features)

## Installation

1. Clone the repository:
```bash
git clone https://github.com/mahigadamsetty/Inshorts-task.git
cd Inshorts-task
```

2. Install dependencies:
```bash
go mod download
```

3. Configure environment variables (optional):
```bash
cp .env.example .env
# Edit .env and add your OPENAI_API_KEY if you have one
```

## Setup

1. Import the news data into the database:
```bash
go run import_data.go "news_data (1).json"
```

This will create a SQLite database (`news.db`) and import all articles.

2. Start the server:
```bash
go run ./cmd/server
```

The server will start on port 8080 by default.

## API Endpoints

All endpoints are prefixed with `/api/v1/news` and return JSON responses.

### 1. Category-based Retrieval
Get news articles from a specific category, ranked by publication date.

```bash
GET /api/v1/news/category?category=Technology&limit=5
```

### 2. Source-based Retrieval
Get news articles from a specific source, ranked by publication date.

```bash
GET /api/v1/news/source?source=New%20York%20Times&limit=5
```

### 3. Relevance Score Retrieval
Get articles with relevance score above a threshold, ranked by score.

```bash
GET /api/v1/news/score?min=0.7&limit=5
```

### 4. Search
Search articles by keywords, ranked by combined relevance score (40%) and text match (60%).

```bash
GET /api/v1/news/search?query=Elon%20Musk%20Twitter&limit=5
```

### 5. Nearby News
Get news near a location within a radius, ranked by distance.

```bash
GET /api/v1/news/nearby?lat=37.4220&lon=-122.0840&radius=10&limit=5
```

Parameters:
- `lat`: Latitude
- `lon`: Longitude
- `radius`: Radius in kilometers
- `limit`: Number of results (default: 5)

### 6. Trending News
Get trending news based on simulated user activity, with location-based ranking and caching.

```bash
GET /api/v1/news/trending?lat=37.4220&lon=-122.0840&limit=5
```

### 7. Natural Language Query
Use natural language to query news. The LLM extracts intent and entities, then dispatches to the appropriate endpoint.

```bash
GET /api/v1/news/query?query=Latest%20developments%20in%20the%20Elon%20Musk%20Twitter%20acquisition%20near%20Palo%20Alto&lat=37.4419&lon=-122.1430&limit=5
```

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
      "publication_date": "2025-03-24T11:08:11",
      "source_name": "Source Name",
      "category": ["General"],
      "relevance_score": 0.85,
      "latitude": 37.4220,
      "longitude": -122.0840,
      "llm_summary": "AI-generated summary of the article..."
    }
  ],
  "meta": {
    "count": 5,
    "limit": 5,
    "endpoint": "search",
    "query": "Elon Musk Twitter"
  }
}
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | SQLite database file path | `news.db` |
| `OPENAI_API_KEY` | OpenAI API key for LLM features | (empty) |
| `LLM_MODEL` | OpenAI model to use | `gpt-4o-mini` |
| `TRENDING_CACHE_TTL` | Trending cache TTL in seconds | `300` |
| `LOCATION_CLUSTER_DEGREES` | Degrees for location clustering | `0.5` |
| `PORT` | Server port | `8080` |

## Architecture

```
.
├── cmd/
│   └── server/
│       └── main.go              # Server entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration loading
│   ├── db/
│   │   └── db.go                # Database setup and migrations
│   ├── models/
│   │   ├── article.go           # Article model
│   │   └── event.go             # Event model
│   ├── llm/
│   │   └── openai.go            # LLM integration
│   ├── utils/
│   │   ├── geo.go               # Geospatial utilities
│   │   └── text.go              # Text matching utilities
│   ├── services/
│   │   ├── ranking.go           # Ranking algorithms
│   │   └── trending.go          # Trending calculation and caching
│   ├── handlers/
│   │   └── news.go              # HTTP handlers
│   └── router/
│       └── router.go            # Route setup
├── import_data.go               # Data import CLI tool
├── news_data (1).json           # News dataset
├── go.mod                       # Go module definition
└── README.md                    # This file
```

## How It Works

### LLM Integration
- **With API Key**: Uses OpenAI to extract entities, determine intent, and generate summaries
- **Without API Key**: Falls back to keyword-based heuristics for intent detection and simple truncation for summaries

### Ranking Strategies
- **Category/Source**: Publication date (newest first)
- **Score**: Relevance score (highest first)
- **Search**: Combined score (40% relevance + 60% text match)
- **Nearby**: Distance (closest first)
- **Trending**: Interaction score with recency decay and geo-relevance

### Trending Algorithm
1. Simulates user events (views and clicks) with location data
2. Calculates score based on:
   - Event type weight (clicks > views)
   - Recency decay (exponential)
   - Geographic proximity to user location
3. Caches results by location clusters for efficiency

## Examples

### Find technology news
```bash
curl "http://localhost:8080/api/v1/news/category?category=technology&limit=5"
```

### Search for specific topics
```bash
curl "http://localhost:8080/api/v1/news/search?query=climate%20change&limit=10"
```

### Get nearby news in San Francisco
```bash
curl "http://localhost:8080/api/v1/news/nearby?lat=37.7749&lon=-122.4194&radius=50&limit=5"
```

### Get trending news
```bash
curl "http://localhost:8080/api/v1/news/trending?lat=37.7749&lon=-122.4194&limit=5"
```

### Natural language query
```bash
curl "http://localhost:8080/api/v1/news/query?query=What%20are%20the%20latest%20tech%20news&limit=5"
```

## Testing

The system includes:
- Consistent error handling with appropriate HTTP status codes
- CORS enabled for all origins (demo purposes)
- Health check endpoint at `/health`
- Automatic event simulation for trending features

## Development

To modify or extend the system:

1. Models are defined in `internal/models/`
2. Add new endpoints in `internal/handlers/news.go`
3. Register routes in `internal/router/router.go`
4. Ranking algorithms are in `internal/services/ranking.go`
5. LLM prompts can be tuned in `internal/llm/openai.go`

## License

This project is part of a technical assessment task.
