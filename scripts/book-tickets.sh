#!/bin/bash
set -e

SHOW_ID="${1:?Usage: $0 <show_id> <tickets> <email>}"
TICKETS="${2:?Missing number of tickets}"
EMAIL="${3:?Missing customer email}"

curl -s -X POST http://localhost:8080/api/book-tickets \
  -H "Content-Type: application/json" \
  -d "{
    \"customer_email\": \"$EMAIL\",
    \"number_of_tickets\": $TICKETS,
    \"show_id\": \"$SHOW_ID\"
  }" | jq .
