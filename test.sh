#!/bin/bash

# Integration Test Script for URL Shortener
# This script tests the complete flow of the application

set -e

BASE_URL="${BASE_URL:-http://localhost:8080}"
FAILED=0
PASSED=0

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "================================"
echo "URL Shortener Integration Tests"
echo "================================"
echo "Base URL: $BASE_URL"
echo ""

# Helper function to print test results
print_result() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓ PASS${NC}: $2"
        ((PASSED++))
    else
        echo -e "${RED}✗ FAIL${NC}: $2"
        ((FAILED++))
    fi
}

# Test 1: Health Check
echo "Test 1: Health Check"
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/health")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

if [ "$HTTP_CODE" -eq 200 ] && echo "$BODY" | grep -q "healthy"; then
    print_result 0 "Health check endpoint"
else
    print_result 1 "Health check endpoint (HTTP $HTTP_CODE)"
fi
echo ""

# Test 2: Create Short URL
echo "Test 2: Create Short URL"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/v1/urls" \
    -H "Content-Type: application/json" \
    -d '{"long_url":"https://www.google.com"}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

if [ "$HTTP_CODE" -eq 201 ]; then
    SHORT_URL=$(echo "$BODY" | grep -o '"short_url":"[^"]*"' | cut -d'"' -f4)
    SHORT_CODE=$(echo "$SHORT_URL" | rev | cut -d'/' -f1 | rev)
    print_result 0 "Create short URL (code: $SHORT_CODE)"
else
    print_result 1 "Create short URL (HTTP $HTTP_CODE)"
    SHORT_CODE=""
fi
echo ""

# Test 3: Redirect
if [ -n "$SHORT_CODE" ]; then
    echo "Test 3: Redirect to Original URL"
    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -L "$BASE_URL/$SHORT_CODE")
    
    if [ "$HTTP_CODE" -eq 200 ]; then
        print_result 0 "Redirect to original URL"
    else
        print_result 1 "Redirect to original URL (HTTP $HTTP_CODE)"
    fi
    echo ""
fi

# Test 4: Analytics
if [ -n "$SHORT_CODE" ]; then
    echo "Test 4: Get Analytics"
    RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/v1/analytics/$SHORT_CODE")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n-1)
    
    if [ "$HTTP_CODE" -eq 200 ] && echo "$BODY" | grep -q "click_count"; then
        CLICK_COUNT=$(echo "$BODY" | grep -o '"click_count":[0-9]*' | cut -d':' -f2)
        print_result 0 "Get analytics (clicks: $CLICK_COUNT)"
    else
        print_result 1 "Get analytics (HTTP $HTTP_CODE)"
    fi
    echo ""
fi

# Test 5: Custom Alias
echo "Test 5: Create URL with Custom Alias"
CUSTOM_ALIAS="test-$(date +%s)"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/v1/urls" \
    -H "Content-Type: application/json" \
    -d "{\"long_url\":\"https://www.github.com\",\"custom_alias\":\"$CUSTOM_ALIAS\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

if [ "$HTTP_CODE" -eq 201 ] && echo "$BODY" | grep -q "$CUSTOM_ALIAS"; then
    print_result 0 "Create URL with custom alias"
else
    print_result 1 "Create URL with custom alias (HTTP $HTTP_CODE)"
fi
echo ""

# Test 6: TTL
echo "Test 6: Create URL with TTL"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/v1/urls" \
    -H "Content-Type: application/json" \
    -d '{"long_url":"https://www.example.com","ttl_days":7}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

if [ "$HTTP_CODE" -eq 201 ] && echo "$BODY" | grep -q "expires_at"; then
    print_result 0 "Create URL with TTL"
else
    print_result 1 "Create URL with TTL (HTTP $HTTP_CODE)"
fi
echo ""

# Test 7: Invalid URL
echo "Test 7: Reject Invalid URL"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/v1/urls" \
    -H "Content-Type: application/json" \
    -d '{"long_url":"not-a-valid-url"}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)

if [ "$HTTP_CODE" -eq 400 ]; then
    print_result 0 "Reject invalid URL"
else
    print_result 1 "Reject invalid URL (expected 400, got $HTTP_CODE)"
fi
echo ""

# Test 8: Missing Long URL
echo "Test 8: Reject Missing Long URL"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/v1/urls" \
    -H "Content-Type: application/json" \
    -d '{"custom_alias":"test"}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)

if [ "$HTTP_CODE" -eq 400 ]; then
    print_result 0 "Reject missing long URL"
else
    print_result 1 "Reject missing long URL (expected 400, got $HTTP_CODE)"
fi
echo ""

# Test 9: Non-existent Short Code
echo "Test 9: Handle Non-existent Short Code"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/nonexistent123")

# Handler redirects to /not-found with 303 (See Other)
if [ "$HTTP_CODE" -eq 303 ]; then
    print_result 0 "Handle non-existent short code"
else
    print_result 1 "Handle non-existent short code (expected 303, got $HTTP_CODE)"
fi
echo ""

# Summary
echo "================================"
echo "Test Summary"
echo "================================"
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
fi
