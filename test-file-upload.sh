#!/bin/bash

# Quick test script for file upload functionality
# This script sets up and tests the file upload feature

set -e

echo "🚀 Bot-Platform File Upload Test Setup"
echo "========================================"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

echo ""
echo -e "${BLUE}Step 1: Checking if bot-platform is compiled...${NC}"
if [ ! -f "./bot" ]; then
    echo -e "${YELLOW}Bot not found, compiling...${NC}"
    make build
fi
echo -e "${GREEN}✓ Bot platform ready${NC}"

echo ""
echo -e "${BLUE}Step 2: Checking if test plugin is compiled...${NC}"
if [ ! -f "./examples/plugin-filetest/filetest-plugin" ]; then
    echo -e "${YELLOW}Test plugin not found, compiling...${NC}"
    cd examples/plugin-filetest
    go build -o filetest-plugin .
    cd ../..
fi
echo -e "${GREEN}✓ Test plugin ready${NC}"

echo ""
echo -e "${BLUE}Step 3: Setting up plugin directories...${NC}"
mkdir -p plugins-bin
mkdir -p plugins-config

echo ""
echo -e "${BLUE}Step 4: Copying test plugin...${NC}"
cp examples/plugin-filetest/filetest-plugin plugins-bin/
echo -e "${GREEN}✓ Plugin copied to plugins-bin/${NC}"

echo ""
echo -e "${BLUE}Step 5: Creating plugin configuration...${NC}"
cat > plugins-config/filetest.json << 'EOF'
{
  "name": "filetest",
  "enabled": true,
  "path": "./plugins-bin/filetest-plugin",
  "auto_start": true
}
EOF
echo -e "${GREEN}✓ Configuration created${NC}"

echo ""
echo -e "${BLUE}Step 6: Creating test file...${NC}"
TEST_FILE="/tmp/bot_test_file.txt"
cat > "$TEST_FILE" << 'EOF'
This is a test file for bot-platform file upload functionality.
Created by the test setup script.

Test timestamp: $(date)
EOF
echo -e "${GREEN}✓ Test file created at: $TEST_FILE${NC}"

echo ""
echo "========================================"
echo -e "${GREEN}✅ Setup Complete!${NC}"
echo ""
echo "📝 Next Steps:"
echo ""
echo "1. Start the bot platform:"
echo "   ./bot"
echo ""
echo "2. In QQ, send these commands to test:"
echo ""
echo "   ${YELLOW}Auto Test (Recommended):${NC}"
echo "   /testfile"
echo ""
echo "   ${YELLOW}Manual Test - Group:${NC}"
echo "   /uploadgroup $TEST_FILE"
echo "   /uploadgroup $TEST_FILE my_custom_name.txt"
echo "   /uploadgroup $TEST_FILE my_custom_name.txt /documents"
echo ""
echo "   ${YELLOW}Manual Test - Private:${NC}"
echo "   /uploadprivate $TEST_FILE"
echo "   /uploadprivate $TEST_FILE my_custom_name.txt"
echo ""
echo "3. Check the results in QQ to verify file upload works"
echo ""
echo "📚 For more details, see:"
echo "   - FILE_UPLOAD_REVIEW.md"
echo "   - examples/plugin-filetest/README.md"
echo ""
