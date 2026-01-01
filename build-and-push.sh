#!/bin/bash

# Build and Push Container Image to DockerHub
# This script builds the bot-platform with file upload functionality and pushes to DockerHub
# Supports both Docker and Podman

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
DOCKER_USERNAME="daikonsushi"
IMAGE_NAME="bot-platform"
FULL_IMAGE_NAME="${DOCKER_USERNAME}/${IMAGE_NAME}"

# Detect container runtime (Podman or Docker)
if command -v podman &> /dev/null; then
    CONTAINER_CMD="podman"
    echo -e "${GREEN}✓ Detected Podman${NC}"
elif command -v docker &> /dev/null; then
    CONTAINER_CMD="docker"
    echo -e "${GREEN}✓ Detected Docker${NC}"
else
    echo -e "${RED}❌ Error: Neither Podman nor Docker found${NC}"
    exit 1
fi

# Get version from git tag or use timestamp
VERSION=$(git describe --tags --always 2>/dev/null || echo "dev-$(date +%Y%m%d-%H%M%S)")

echo -e "${BLUE}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  Bot-Platform Container Build & Push Script           ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${GREEN}Container Runtime:${NC} ${CONTAINER_CMD}"
echo -e "${GREEN}Image:${NC} ${FULL_IMAGE_NAME}"
echo -e "${GREEN}Version:${NC} ${VERSION}"
echo ""

# Check if container runtime is working
if ! ${CONTAINER_CMD} info > /dev/null 2>&1; then
    echo -e "${RED}❌ Error: ${CONTAINER_CMD} is not running or not accessible${NC}"
    exit 1
fi

# Check if logged in to DockerHub
echo -e "${BLUE}Step 1: Checking DockerHub login status...${NC}"
if ! ${CONTAINER_CMD} login --get-login docker.io > /dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  Not logged in to DockerHub${NC}"
    echo -e "${YELLOW}Please login to DockerHub:${NC}"
    ${CONTAINER_CMD} login docker.io
else
    echo -e "${GREEN}✓ Already logged in to DockerHub${NC}"
fi

# Ask user for build options
echo ""
echo -e "${BLUE}Step 2: Select build options${NC}"
echo ""
echo "Choose build type:"
echo "  1) Quick build (current platform only) - Fast"
if [ "$CONTAINER_CMD" = "podman" ]; then
    echo "  2) Multi-platform build (amd64 + arm64) - Requires podman with buildx support"
else
    echo "  2) Multi-platform build (amd64 + arm64) - Slower but recommended"
fi
echo ""
read -p "Enter choice [1-2] (default: 1): " BUILD_CHOICE
BUILD_CHOICE=${BUILD_CHOICE:-1}

# Ask for version tag
echo ""
read -p "Enter version tag (default: ${VERSION}): " USER_VERSION
VERSION=${USER_VERSION:-$VERSION}

# Confirm before building
echo ""
echo -e "${YELLOW}═══════════════════════════════════════════════════════${NC}"
echo -e "${YELLOW}Build Configuration:${NC}"
echo -e "  Runtime: ${CONTAINER_CMD}"
echo -e "  Image: ${FULL_IMAGE_NAME}"
echo -e "  Tags: latest, ${VERSION}"
if [ "$BUILD_CHOICE" = "2" ]; then
    echo -e "  Platforms: linux/amd64, linux/arm64"
else
    echo -e "  Platform: Current platform only"
fi
echo -e "${YELLOW}═══════════════════════════════════════════════════════${NC}"
echo ""
read -p "Continue? [Y/n]: " CONFIRM
CONFIRM=${CONFIRM:-Y}

if [[ ! $CONFIRM =~ ^[Yy]$ ]]; then
    echo -e "${RED}Build cancelled${NC}"
    exit 0
fi

# Build and push
echo ""
echo -e "${BLUE}Step 3: Building and pushing container image...${NC}"
echo ""

if [ "$BUILD_CHOICE" = "2" ]; then
    # Multi-platform build
    echo -e "${YELLOW}Building for multiple platforms (this may take a while)...${NC}"
    
    if [ "$CONTAINER_CMD" = "podman" ]; then
        # Podman multi-platform build
        echo -e "${YELLOW}Using podman manifest for multi-platform build...${NC}"
        
        # Remove existing manifest/image if exists
        echo -e "${YELLOW}Cleaning up existing manifest/image...${NC}"
        ${CONTAINER_CMD} manifest rm ${FULL_IMAGE_NAME}:latest 2>/dev/null || true
        ${CONTAINER_CMD} rmi localhost/${FULL_IMAGE_NAME}:latest 2>/dev/null || true
        ${CONTAINER_CMD} rmi ${FULL_IMAGE_NAME}:latest 2>/dev/null || true
        
        # Create manifest
        echo -e "${YELLOW}Creating new manifest...${NC}"
        ${CONTAINER_CMD} manifest create ${FULL_IMAGE_NAME}:latest
        
        # Build for amd64
        echo -e "${YELLOW}Building for linux/amd64...${NC}"
        ${CONTAINER_CMD} build \
            --platform linux/amd64 \
            --manifest ${FULL_IMAGE_NAME}:latest \
            .
        
        # Build for arm64
        echo -e "${YELLOW}Building for linux/arm64...${NC}"
        ${CONTAINER_CMD} build \
            --platform linux/arm64 \
            --manifest ${FULL_IMAGE_NAME}:latest \
            .
        
        # Push manifest
        echo -e "${YELLOW}Pushing manifest...${NC}"
        ${CONTAINER_CMD} manifest push ${FULL_IMAGE_NAME}:latest docker://${FULL_IMAGE_NAME}:latest
        
        # Tag and push version
        ${CONTAINER_CMD} tag ${FULL_IMAGE_NAME}:latest ${FULL_IMAGE_NAME}:${VERSION}
        ${CONTAINER_CMD} push ${FULL_IMAGE_NAME}:${VERSION}
        
    else
        # Docker buildx multi-platform build
        # Create buildx builder if not exists
        if ! ${CONTAINER_CMD} buildx ls | grep -q "multiplatform"; then
            echo -e "${YELLOW}Creating buildx builder...${NC}"
            ${CONTAINER_CMD} buildx create --name multiplatform --use
            ${CONTAINER_CMD} buildx inspect --bootstrap
        else
            ${CONTAINER_CMD} buildx use multiplatform
        fi
        
        # Build and push
        ${CONTAINER_CMD} buildx build \
            --platform linux/amd64,linux/arm64 \
            --tag ${FULL_IMAGE_NAME}:latest \
            --tag ${FULL_IMAGE_NAME}:${VERSION} \
            --push \
            .
    fi
    
    echo -e "${GREEN}✓ Multi-platform build completed${NC}"
else
    # Single platform build
    echo -e "${YELLOW}Building for current platform...${NC}"
    
    # Build
    ${CONTAINER_CMD} build \
        --tag ${FULL_IMAGE_NAME}:latest \
        --tag ${FULL_IMAGE_NAME}:${VERSION} \
        .
    
    # Push
    echo -e "${YELLOW}Pushing to DockerHub...${NC}"
    ${CONTAINER_CMD} push ${FULL_IMAGE_NAME}:latest
    ${CONTAINER_CMD} push ${FULL_IMAGE_NAME}:${VERSION}
    
    echo -e "${GREEN}✓ Single platform build completed${NC}"
fi

# Summary
echo ""
echo -e "${GREEN}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║  ✅ Build and Push Completed Successfully!             ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${BLUE}📦 Images pushed:${NC}"
echo -e "  • ${FULL_IMAGE_NAME}:latest"
echo -e "  • ${FULL_IMAGE_NAME}:${VERSION}"
echo ""
echo -e "${BLUE}🚀 To update your remote server:${NC}"
echo ""
echo -e "${YELLOW}1. SSH to your remote server${NC}"
echo ""
echo -e "${YELLOW}2. Pull the latest image:${NC}"
if [ "$CONTAINER_CMD" = "podman" ]; then
    echo "   podman-compose pull bot-platform"
    echo "   # or: podman pull ${FULL_IMAGE_NAME}:latest"
else
    echo "   docker-compose pull bot-platform"
fi
echo ""
echo -e "${YELLOW}3. Restart the service:${NC}"
if [ "$CONTAINER_CMD" = "podman" ]; then
    echo "   podman-compose up -d bot-platform"
    echo "   # or: podman-compose restart bot-platform"
else
    echo "   docker-compose up -d bot-platform"
fi
echo ""
echo -e "${YELLOW}4. Check logs:${NC}"
if [ "$CONTAINER_CMD" = "podman" ]; then
    echo "   podman-compose logs -f bot-platform"
    echo "   # or: podman logs -f <container-name>"
else
    echo "   docker-compose logs -f bot-platform"
fi
echo ""
echo -e "${BLUE}📝 Or use the one-liner:${NC}"
if [ "$CONTAINER_CMD" = "podman" ]; then
    echo "   podman-compose pull bot-platform && podman-compose up -d bot-platform && podman-compose logs -f bot-platform"
else
    echo "   docker-compose pull bot-platform && docker-compose up -d bot-platform && docker-compose logs -f bot-platform"
fi
echo ""
echo -e "${GREEN}Done! 🎉${NC}"
echo ""
