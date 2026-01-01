# Quick Deployment Guide for Linux Server

## 🚀 Deploy on Your x86 Linux Server

### Prerequisites
- Docker or Podman installed
- docker-compose installed (if using Docker Compose)

---

## Method 1: Using Docker Compose (Recommended)

### Step 1: Pull the latest image

```bash
cd /path/to/your/napcat/bot-platform
docker-compose pull bot-platform
```

### Step 2: Restart the service

```bash
docker-compose up -d bot-platform
```

### Step 3: Check logs

```bash
docker-compose logs -f bot-platform
```

---

## Method 2: Using Docker Directly

### Pull and run:

```bash
docker pull docker.io/daikonsushi/bot-platform:latest

docker run -d \
  --name bot-platform \
  --network napcat_bot-network \
  -p 8080:8080 \
  -p 50051:50051 \
  -e NAPCAT_HTTP_URL=http://napcat:3000 \
  -e NAPCAT_WS_URL=ws://napcat:3001 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  -v $(pwd)/plugins-bin:/app/plugins-bin \
  -v $(pwd)/plugins-config:/app/plugins-config \
  docker.io/daikonsushi/bot-platform:latest
```

---

## Method 3: Update Existing Container

### Stop and remove old container:

```bash
docker stop bot-platform
docker rm bot-platform
```

### Pull new image:

```bash
docker pull docker.io/daikonsushi/bot-platform:latest
```

### Start with docker-compose:

```bash
docker-compose up -d bot-platform
```

---

## Verification

### Check if container is running:

```bash
docker ps | grep bot-platform
```

### Check logs:

```bash
docker logs -f bot-platform
```

### Test the bot:

Send `/help` to your QQ bot to verify it's working.

---

## Troubleshooting

### Container won't start

```bash
# Check logs
docker logs bot-platform

# Check if config file exists
ls -la config.yaml

# Check if ports are available
netstat -tulpn | grep -E '8080|50051'
```

### Image pull fails

```bash
# Try with full registry path
docker pull docker.io/daikonsushi/bot-platform:latest

# Or use a specific date tag
docker pull docker.io/daikonsushi/bot-platform:20260101
```

### Permission issues

```bash
# Fix volume permissions
sudo chown -R 1000:1000 plugins-bin plugins-config
```

---

## Rollback to Previous Version

If something goes wrong, you can rollback:

```bash
# Use date-tagged version
docker pull docker.io/daikonsushi/bot-platform:20260101
docker-compose up -d bot-platform
```

---

## Health Check

```bash
# Check if admin API is responding
curl http://localhost:8080/health

# Check if gRPC port is listening
netstat -tulpn | grep 50051
```

---

## Complete Restart (Both Services)

```bash
# Restart everything
docker-compose down
docker-compose pull
docker-compose up -d

# Check status
docker-compose ps
```

---

## Image Information

- **Registry**: docker.io
- **Repository**: daikonsushi/bot-platform
- **Tags**: 
  - `latest` - Most recent build
  - `20260101` - Date-tagged version (YYYYMMDD)
- **Architecture**: linux/amd64 (x86_64)
- **Size**: ~23 MB

---

## What's New in This Version

✅ Help plugin now supports configuration customization
✅ Added pluginctl for managing plugins via bot commands
✅ Improved error handling and logging
✅ Updated dependencies

---

## Support

If you encounter issues:
1. Check logs: `docker logs bot-platform`
2. Verify config: `cat config.yaml`
3. Check network: `docker network inspect napcat_bot-network`
4. Restart services: `docker-compose restart`

---

## Quick Commands Reference

```bash
# Pull latest
docker-compose pull bot-platform

# Start
docker-compose up -d bot-platform

# Stop
docker-compose stop bot-platform

# Restart
docker-compose restart bot-platform

# Logs
docker-compose logs -f bot-platform

# Shell access
docker exec -it bot-platform sh

# Remove
docker-compose down bot-platform
```
