#!/bin/bash
set -e

TITLE="${1:?Usage: $0 <title> [venue] [tickets]}"
VENUE="${2:-Colosseum}"
TICKETS="${3:-10000}"

curl -s -X POST http://localhost:8080/api/shows \
  -H "Content-Type: application/json" \
  -d "{
    \"dead_nation_id\": \"0fe9f3bf-160f-49be-9509-862e91ee8c33\",
    \"number_of_tickets\": $TICKETS,
    \"start_time\": \"2024-02-04T19:00:00Z\",
    \"title\": \"$TITLE\",
    \"venue\": \"$VENUE\"
  }" | jq .
