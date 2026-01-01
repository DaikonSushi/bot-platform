# Docker Build and Deployment Guide

This guide explains how to build and deploy the bot-platform Docker image from a Mac M3 (ARM64) to an x86 Linux server.

## Prerequisites

### On Mac (Build Machine)
- ✅ Podman installed
- ✅ Docker Hub account
- ✅ Logged in to Docker Hub: `podman login docker.io`

### On Linux Server (Target Machine)
- ✅ Docker or Podman installed
- ✅ docker-compose installed (if using Docker Compose)

## Quick Start

### Step 1: Login to Docker Hub (if not already)

```bash
podman login docker.io
# Enter your Docker Hub username and password
```

### Step 2: Build and Push

Simply run the build script:

```bash
./build-and-push.sh
```

This script will:
1. ✅ Check if podman is installed
2. ✅ Verify Docker Hub login
3. ✅ Build the image for linux/amd64 (x86_64)
4. ✅ Tag with `latest` and date (e.g., `20260101`)
5. ✅ Push to Docker Hub

### Step 3: Deploy on Linux Server

On your x86 Linux server, pull and run:

```bash
# Using docker-compose (recommended)
docker-compose pull bot-platform
docker-compose up -d bot-platform

# Or using docker directly
docker pull docker.io/daikonsushi/bot-platform:latest
docker run -d \
  --name bot-platform \
  -p 8080:8080 \
  -p 50051:50051 \
  -v ./config.yaml:/app/config.yaml \
  -v ./plugins-bin:/app/plugins-bin \
  -v ./plugins-config:/app/plugins-config \
  docker.io/daikonsushi/bot-platform:latest
```

## Manual Build (Advanced)

If you prefer to build manually:

```bash
# Build for x86_64 Linux
podman build \
  --platform linux/amd64 \
  --tag docker.io/daikonsushi/bot-platform:latest \
  -f Dockerfile \
  .

# Push to Docker Hub
podman push docker.io/daikonsushi/bot-platform:latest
```

## Multi-Architecture Build (Optional)

If you want to support both ARM64 and AMD64:

```bash
# Create a manifest
podman manifest create docker.io/daikonsushi/bot-platform:latest

# Build for AMD64
podman build \
  --platform linux/amd64 \
  --manifest docker.io/daikonsushi/bot-platform:latest \
  -f Dockerfile \
  .

# Build for ARM64
podman build \
  --platform linux/arm64 \
  --manifest docker.io/daikonsushi/bot-platform:latest \
  -f Dockerfile \
  .

# Push the manifest
podman manifest push docker.io/daikonsushi/bot-platform:latest
```

## Troubleshooting

### Issue: "Error: short-name resolution enforced"

**Solution**: Use the full image name with registry:
```bash
docker.io/daikonsushi/bot-platform:latest
```

### Issue: Build fails with "exec format error"

**Cause**: Trying to run ARM64 binary on x86_64 or vice versa.

**Solution**: Make sure you're building with `--platform linux/amd64` for x86 servers.

### Issue: "unauthorized: authentication required"

**Solution**: Login to Docker Hub first:
```bash
podman login docker.io
```

### Issue: Slow build on Mac M3

**Cause**: Cross-compilation from ARM64 to AMD64 uses QEMU emulation.

**Solution**: This is normal. The Dockerfile is optimized with multi-stage builds to minimize emulation overhead. The builder stage runs natively on ARM64, only the final binary is cross-compiled.

### Issue: "no space left on device"

**Solution**: Clean up old images:
```bash
podman system prune -a
```

## Image Tags

- `latest` - Always points to the most recent build
- `YYYYMMDD` - Date-tagged versions (e.g., `20260101`)

## Verification

After pushing, verify the image on Docker Hub:

```bash
# Check image info
podman manifest inspect docker.io/daikonsushi/bot-platform:latest

# Or visit Docker Hub
# https://hub.docker.com/r/daikonsushi/bot-platform
```

## CI/CD Integration (Future)

For automated builds, consider using:
- GitHub Actions
- GitLab CI
- Jenkins

Example GitHub Actions workflow:

```yaml
name: Build and Push Docker Image

on:
  push:
    branches: [ main ]
    tags: [ 'v*' ]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up QEMU
        uses: docker/setup-qemu-action@v2
      
      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v2
      
      - name: Login to Docker Hub
        uses: docker/login-action@v2
        with:
          username: ${{ secrets.DOCKERHUB_USERNAME }}
          password: ${{ secrets.DOCKERHUB_TOKEN }}
      
      - name: Build and push
        uses: docker/build-push-action@v4
        with:
          context: .
          platforms: linux/amd64,linux/arm64
          push: true
          tags: |
            daikonsushi/bot-platform:latest
            daikonsushi/bot-platform:${{ github.sha }}
```

## Security Notes

1. **Never commit credentials** - Use environment variables or secrets
2. **Use specific tags** - Avoid using `latest` in production
3. **Scan images** - Use `podman scan` or Docker Scout
4. **Minimize image size** - Current image uses Alpine Linux (small footprint)

## Performance Tips

1. **Use BuildKit** - Already enabled in Dockerfile
2. **Layer caching** - Dependencies are cached separately from source code
3. **Multi-stage builds** - Reduces final image size
4. **Go proxy** - Set `GOPROXY=https://goproxy.cn,direct` for faster builds in China

## Support

If you encounter issues:
1. Check the [Troubleshooting](#troubleshooting) section
2. Review Docker/Podman logs: `podman logs bot-platform`
3. Open an issue on GitHub

## Related Files

- [`Dockerfile`](../Dockerfile) - Multi-stage build configuration
- [`docker-compose.yaml`](../docker-compose.yaml) - Service orchestration
- [`build-and-push.sh`](../build-and-push.sh) - Build automation script
