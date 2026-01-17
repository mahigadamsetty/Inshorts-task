#!/bin/bash

# Test script for News API endpoints
# Usage: ./test_endpoints.sh

PORT=8085
BASE_URL="http://localhost:${PORT}/api/v1/news"

echo "Testing News API Endpoints..."
echo "============================="
echo ""

# Test health endpoint
echo "1. Testing /health endpoint..."
curl -s "http://localhost:${PORT}/health" | python3 -m json.tool
echo ""
echo ""

# Test category endpoint
echo "2. Testing /category endpoint..."
curl -s "${BASE_URL}/category?category=technology&limit=2" | python3 -m json.tool | head -40
echo ""
echo ""

# Test source endpoint
echo "3. Testing /source endpoint..."
curl -s "${BASE_URL}/source?source=Reuters&limit=2" | python3 -m json.tool | head -40
echo ""
echo ""

# Test score endpoint
echo "4. Testing /score endpoint..."
curl -s "${BASE_URL}/score?min=0.9&limit=2" | python3 -m json.tool | head -40
echo ""
echo ""

# Test search endpoint
echo "5. Testing /search endpoint..."
curl -s "${BASE_URL}/search?query=cricket&limit=2" | python3 -m json.tool | head -40
echo ""
echo ""

# Test nearby endpoint
echo "6. Testing /nearby endpoint..."
curl -s "${BASE_URL}/nearby?lat=19.0760&lon=72.8777&radius=50&limit=2" | python3 -m json.tool | head -40
echo ""
echo ""

# Test trending endpoint
echo "7. Testing /trending endpoint..."
curl -s "${BASE_URL}/trending?lat=19.0760&lon=72.8777&limit=2" | python3 -m json.tool | head -40
echo ""
echo ""

# Test query endpoint with technology intent
echo "8. Testing /query endpoint (technology intent)..."
curl -s "${BASE_URL}/query?query=Show%20me%20technology%20news&limit=2" | python3 -m json.tool | head -40
echo ""
echo ""

# Test query endpoint with search intent
echo "9. Testing /query endpoint (search intent)..."
curl -s "${BASE_URL}/query?query=cricket%20match%20results&limit=2" | python3 -m json.tool | head -40
echo ""
echo ""

echo "============================="
echo "All endpoint tests completed!"
