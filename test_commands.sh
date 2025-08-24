#!/bin/bash

# Test script for Gex Shell built-in commands

echo "🚀 Testing Gex Shell Built-in Commands"
echo "======================================"

echo -e "ls -la\necho 'Testing file operations...'\nmkdir test_gex\ntouch test_gex/file1.txt test_gex/file2.txt\nls test_gex\necho 'Hello World' | cat\nps | head -10\ndf -h | head -5\nfree -h\nuptime\nuname -a\nrm -rf test_gex\nexit" | ./gex
