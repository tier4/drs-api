#!/bin/bash

# Build script for DRS Docker images

set -e

# Change to project root
cd "$(dirname "$0")/.."

echo "Building DRS Docker images..."

# Build API Gateway
echo "Building tier4/drs-api-gateway:latest..."
docker build -f docker/api-gateway/Dockerfile -t tier4/drs-api-gateway:latest .

# Build Dashboard
echo "Building tier4/drs-dashboard:latest..."
docker build -f docker/dashboard/Dockerfile -t tier4/drs-dashboard:latest .

echo "All images built successfully!"
echo ""
echo "Images created:"
docker images | grep "tier4/drs-" | head -4