#!/bin/bash

# Gex Shell Release Script
# Automates the release process

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we're in the right directory
if [ ! -f "go.mod" ] || [ ! -f "main.go" ]; then
    log_error "This script must be run from the gex project root directory"
    exit 1
fi

# Get current version from git tags
CURRENT_VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
log_info "Current version: $CURRENT_VERSION"

# Ask for new version
echo "Enter new version (format: v1.2.3):"
read -r NEW_VERSION

# Validate version format
if [[ ! $NEW_VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    log_error "Invalid version format. Use v1.2.3 format"
    exit 1
fi

# Check if version already exists
if git tag | grep -q "^$NEW_VERSION$"; then
    log_error "Version $NEW_VERSION already exists"
    exit 1
fi

log_info "Preparing release $NEW_VERSION..."

# Update version in files
log_info "Updating version in files..."
sed -i "s/VERSION = \".*\"/VERSION = \"${NEW_VERSION#v}\"/" main.go
sed -i "s/GEX_VERSION=\".*\"/GEX_VERSION=\"${NEW_VERSION#v}\"/" install.sh

# Run tests
log_info "Running tests..."
go test ./... || {
    log_error "Tests failed"
    exit 1
}

# Run linter
log_info "Running linter..."
golangci-lint run || {
    log_warning "Linter warnings found, but continuing..."
}

# Build and test
log_info "Building and testing..."
make clean
make build || {
    log_error "Build failed"
    exit 1
}

# Test the binary
log_info "Testing binary..."
echo "help" | ./build/gex > /dev/null || {
    log_error "Binary test failed"
    exit 1
}

# Create changelog entry
log_info "Creating changelog entry..."
CHANGELOG_FILE="CHANGELOG.md"

if [ ! -f "$CHANGELOG_FILE" ]; then
    cat > "$CHANGELOG_FILE" << EOF
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

EOF
fi

# Get commits since last tag
COMMITS=$(git log --pretty=format:"- %s" "${CURRENT_VERSION}..HEAD" 2>/dev/null || git log --pretty=format:"- %s" -10)

# Add new version to changelog
temp_file=$(mktemp)
cat > "$temp_file" << EOF
# Changelog

All notable changes to this project will be documented in this file.

## [${NEW_VERSION}] - $(date +%Y-%m-%d)

### Added
- High-performance Linux shell implementation
- Built-in commands (cd, pwd, echo, help, history, alias, etc.)
- Advanced readline with history and completion
- Command pipes and redirection
- Environment variable support
- Object pooling and caching for performance

### Changed
$COMMITS

### Fixed
- Various performance improvements
- Memory leak fixes
- Better error handling

EOF

# Append rest of changelog if it exists
if [ -f "$CHANGELOG_FILE" ]; then
    tail -n +4 "$CHANGELOG_FILE" >> "$temp_file"
fi

mv "$temp_file" "$CHANGELOG_FILE"

# Commit changes
log_info "Committing changes..."
git add .
git commit -m "Release $NEW_VERSION

- Update version to $NEW_VERSION
- Update changelog
- Prepare for release" || {
    log_warning "No changes to commit"
}

# Create and push tag
log_info "Creating and pushing tag..."
git tag -a "$NEW_VERSION" -m "Release $NEW_VERSION

$(echo "$COMMITS" | head -10)

Full changelog: https://github.com/dmitryzhvinklis/gex/blob/main/CHANGELOG.md"

# Ask for confirmation before pushing
echo ""
log_warning "Ready to push tag $NEW_VERSION to origin."
echo "This will trigger the GitHub Actions release workflow."
echo "Continue? (y/N)"
read -r CONFIRM

if [[ $CONFIRM =~ ^[Yy]$ ]]; then
    git push origin main
    git push origin "$NEW_VERSION"
    
    log_success "Tag $NEW_VERSION pushed successfully!"
    log_info "GitHub Actions will now build and create the release."
    log_info "Check the Actions tab: https://github.com/dmitryzhvinklis/gex/actions"
    log_info "Release will be available at: https://github.com/dmitryzhvinklis/gex/releases"
else
    log_info "Release cancelled. Tag created locally but not pushed."
    log_info "To push later: git push origin main && git push origin $NEW_VERSION"
    log_info "To delete local tag: git tag -d $NEW_VERSION"
fi

echo ""
log_success "Release process completed!"
