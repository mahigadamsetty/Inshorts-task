# Contextual News Data Retrieval System

A Go backend for fetching and organizing news articles with LLM-powered entity extraction, intent classification, and location-based trending analysis.

## Features

- **Multiple API Endpoints**: Category, source, score, search, nearby, trending, and LLM-powered query endpoints
- **LLM Integration**: OpenAI-powered entity extraction and summarization with graceful fallback
- **Location-Based Features**: Haversine distance calculation and trending news by location
- **Trending System**: Simulated user events with temporal decay and geographical relevance
- **Caching**: Location-clustered trending feed caching with configurable TTL
- **SQLite Storage**: GORM-based database with automatic migrations

## Prerequisites

- Go 1.21 or higher
- OpenAI API key (optional - system works with fallback if not provided)

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

3. Create environment configuration (optional):
```bash
cp .env.example .env
# Edit .env to add your OpenAI API key and other settings
```

## Configuration

Environment variables (all optional with sensible defaults):

- `DATABASE_URL`: SQLite database file path (default: `news.db`)
- `OPENAI_API_KEY`: OpenAI API key for LLM features (optional)
- `LLM_MODEL`: OpenAI model to use (default: `gpt-4o-mini`)
- `TRENDING_CACHE_TTL`: Cache TTL in seconds (default: `300`)
- `LOCATION_CLUSTER_DEGREES`: Location clustering granularity (default: `0.5`)
- `PORT`: Server port (default: `8080`)

## Usage

### 1. Import News Data

Import the news dataset into the database:

```bash
go run import_data.go "news_data (1).json"
```

This will:
- Parse and import 2000 news articles
- Simulate 1000 user interaction events for trending analysis
- Create database indexes for efficient querying

### 2. Start the Server

```bash
go run ./cmd/server
```

The server will start on `http://localhost:8080`

## API Endpoints

Base URL: `/api/v1/news`

### 1. Get by Category
```bash
GET /api/v1/news/category?category=Technology&limit=5
```

**Parameters:**
- `category` (required): News category (e.g., Technology, Sports, Business)
- `limit` (optional): Number of articles to return (default: 5)

**Ranking:** Publication date (newest first)

### 2. Get by Source
```bash
GET /api/v1/news/source?source=New%20York%20Times&limit=5
```

**Parameters:**
- `source` (required): News source name
- `limit` (optional): Number of articles (default: 5)

**Ranking:** Publication date (newest first)

### 3. Get by Relevance Score
```bash
GET /api/v1/news/score?min=0.7&limit=5
```

**Parameters:**
- `min` (optional): Minimum relevance score (default: 0.0)
- `limit` (optional): Number of articles (default: 5)

**Ranking:** Relevance score (highest first)

### 4. Search
```bash
GET /api/v1/news/search?query=Elon%20Musk%20Twitter&limit=5
```

**Parameters:**
- `query` (required): Search keywords
- `limit` (optional): Number of articles (default: 5)

**Ranking:** Combined score (40% relevance_score + 60% text match)

### 5. Nearby News
```bash
GET /api/v1/news/nearby?lat=37.4220&lon=-122.0840&radius=10&limit=5
```

**Parameters:**
- `lat` (required): Latitude
- `lon` (required): Longitude
- `radius` (optional): Search radius in km (default: 10)
- `limit` (optional): Number of articles (default: 5)

**Ranking:** Distance (nearest first using Haversine formula)

### 6. Trending News
```bash
GET /api/v1/news/trending?lat=37.4220&lon=-122.0840&limit=5
```

**Parameters:**
- `lat` (required): Latitude
- `lon` (required): Longitude
- `limit` (optional): Number of articles (default: 5)

**Ranking:** Trending score based on:
- User interaction volume (clicks weighted more than views)
- Recency of interactions (exponential decay)
- Geographical proximity to query location

**Caching:** Results cached by location cluster with configurable TTL

### 7. LLM-Powered Query
```bash
GET /api/v1/news/query?query=Latest%20developments%20in%20the%20Elon%20Musk%20Twitter%20acquisition%20near%20Palo%20Alto&lat=37.4419&lon=-122.1430&limit=5
```

**Parameters:**
- `query` (required): Natural language query
- `lat` (optional): Latitude for location-based queries
- `lon` (optional): Longitude for location-based queries
- `limit` (optional): Number of articles (default: 5)

**Features:**
- LLM extracts entities and determines intent
- Automatically routes to appropriate endpoint
- Supports intents: category, source, search, nearby, score

## Response Format

All endpoints return a consistent JSON structure:

```json
{
  "articles": [
    {
      "id": "uuid",
      "title": "Article Title",
      "description": "Article description...",
      "url": "https://...",
      "publication_date": "2025-03-24T11:08:11",
      "source_name": "Source Name",
      "category": ["Category"],
      "relevance_score": 0.85,
      "latitude": 37.7749,
      "longitude": -122.4194,
      "llm_summary": "This article discusses..."
    }
  ],
  "meta": {
    "count": 5,
    "limit": 5,
    "endpoint": "search",
    "query": "search terms"
  }
}
```

## Example Requests

```bash
# Get technology news
curl "http://localhost:8080/api/v1/news/category?category=Technology&limit=3"

# High-quality articles
curl "http://localhost:8080/api/v1/news/score?min=0.8&limit=5"

# News near San Francisco
curl "http://localhost:8080/api/v1/news/nearby?lat=37.7749&lon=-122.4194&radius=50&limit=5"

# Trending in New York
curl "http://localhost:8080/api/v1/news/trending?lat=40.7128&lon=-74.0060&limit=5"

# Natural language query
curl "http://localhost:8080/api/v1/news/query?query=Show%20me%20top%20technology%20news&limit=5"

# Health check
curl "http://localhost:8080/health"
```

## Architecture

### Project Structure
```
.
├── cmd/
│   └── server/
│       └── main.go          # Server entry point
├── internal/
│   ├── config/
│   │   └── config.go        # Configuration management
│   ├── db/
│   │   └── db.go            # Database initialization
│   ├── models/
│   │   ├── article.go       # Article model
│   │   └── event.go         # Event model
│   ├── llm/
│   │   └── openai.go        # LLM integration
│   ├── utils/
│   │   └── geo.go           # Geospatial utilities
│   ├── services/
│   │   ├── ranking.go       # Ranking algorithms
│   │   └── trending.go      # Trending & caching
│   ├── handlers/
│   │   └── news.go          # HTTP handlers
│   └── router/
│       └── router.go        # Route configuration
├── import_data.go           # Data import CLI
├── go.mod
└── README.md
```

### Key Components

1. **Database Layer**: SQLite with GORM ORM for efficient querying and migrations
2. **LLM Service**: OpenAI integration with fallback to heuristic extraction
3. **Ranking Engine**: Multiple algorithms for different endpoint requirements
4. **Trending System**: Event simulation, scoring, and location-based caching
5. **HTTP Layer**: Gin framework with CORS support

## Development

Build the project:
```bash
go build -o server ./cmd/server
./server
```

Run with custom configuration:
```bash
DATABASE_URL=mydb.db PORT=3000 go run ./cmd/server
```

## Trending System Details

The trending system simulates user behavior and computes trending scores based on:

1. **Event Types**: 
   - Views (weight: 1.0)
   - Clicks (weight: 2.0)

2. **Temporal Decay**: 
   - Exponential decay with 12-hour half-life
   - Recent interactions weighted more heavily

3. **Geographical Relevance**:
   - Inverse distance weighting
   - Events closer to query location score higher

4. **Caching Strategy**:
   - Location clustering via configurable degree rounding
   - TTL-based cache invalidation
   - Automatic cleanup of expired entries

## Error Handling

The API returns standard HTTP status codes:

- `200`: Success
- `400`: Bad request (missing/invalid parameters)
- `500`: Internal server error

Error responses:
```json
{
  "error": "Error message description"
}
```

## License

MIT

## Contributing

Pull requests welcome! Please ensure code follows Go best practices and includes appropriate error handling.
