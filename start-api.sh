#!/bin/bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=""
export DB_NAME=xupu
export JWT_SECRET=test-secret
export FANQIE_COOKIE=""
export PORT=8080

./bin/xupu-api
