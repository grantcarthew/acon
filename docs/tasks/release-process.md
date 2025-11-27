# acon - Release Process

> **Purpose**: Repeatable process for releasing new versions of acon
> **Audience**: AI agents and maintainers performing releases
> **Last Updated**: 2025-11-27

This document provides step-by-step instructions for releasing acon. Execute each step in order.

---

## Prerequisites

Verify before starting:

- Write access to `grantcarthew/acon` repository
- Write access to `grantcarthew/homebrew-tap` repository
- Go 1.25+ installed
- Git configured with proper credentials
- GitHub CLI (`gh`) installed and authenticated
- All planned features/fixes merged to main branch

---

## Release Process

**Steps**:

1. Run pre-release validation
2. Determine version number
3. Commit any pending changes
4. Create and push git tag
5. Create GitHub Release
6. Update Homebrew tap
7. Verify installation
8. Clean up

**Estimated Time**: 15-20 minutes

---

## Step 1: Pre-Release Validation

Run validation checks:

```bash
# Ensure on main branch with latest changes
git checkout main
git pull origin main

# Verify build works
go build -o acon
./acon --version
rm acon

# Verify clean working directory
git status
```

**Expected results**:

- Build completes without errors
- `git status` shows clean working tree
- Documentation is current

**If any validation fails, stop and fix issues before proceeding.**

---

## Step 2: Determine Version Number

Set the version number using [Semantic Versioning](https://semver.org/):

- **MAJOR**: Breaking API changes (1.0.0 → 2.0.0)
- **MINOR**: New features, backward compatible (1.0.0 → 1.1.0)
- **PATCH**: Bug fixes only (1.0.0 → 1.0.1)

```bash
# Check current version
git tag -l | tail -1

# Set new version (example: v1.0.0)
export VERSION="1.0.0"
echo "Releasing version: v${VERSION}"
```

---

## Step 3: Commit Any Pending Changes

Ensure all changes are committed:

```bash
# Check for uncommitted changes
git status

# If there are changes, commit them
git add .
git commit -m "chore: prepare for v${VERSION} release"
git push origin main
```

---

## Step 4: Create and Push Git Tag

Create an annotated git tag:

```bash
# Get previous version and review changes
PREV_VERSION=$(git tag -l | tail -1)
git log ${PREV_VERSION}..HEAD --oneline

# Create one-line summary from the changes above
# Examples: "Initial release", "Add markdown conversion", "Fix API authentication"
SUMMARY="Your one-line summary here"

# Create and push annotated tag
git tag -a "v${VERSION}" -m "Release v${VERSION} - ${SUMMARY}"
git push origin "v${VERSION}"

# Verify tag exists
git tag -l -n9 "v${VERSION}"
```

---

## Step 5: Create GitHub Release

Create the GitHub Release with release notes:

```bash
# Wait for tarball to be generated (usually immediate)
sleep 5

# Get tarball SHA256 for Homebrew (will use in Step 6)
TARBALL_URL="https://github.com/grantcarthew/acon/archive/refs/tags/v${VERSION}.tar.gz"
# macOS:
TARBALL_SHA256=$(curl -sL "$TARBALL_URL" | shasum -a 256 | cut -d' ' -f1)
# Linux:
# TARBALL_SHA256=$(curl -sL "$TARBALL_URL" | sha256sum | cut -d' ' -f1)
echo "Tarball SHA256: $TARBALL_SHA256"

# Create GitHub Release using gh CLI
gh release create "v${VERSION}" \
  --title "Release v${VERSION}" \
  --notes "$(cat <<EOF
## Changes

$(git log ${PREV_VERSION}..v${VERSION} --pretty=format:"- %s" --reverse)

See [README.md](https://github.com/grantcarthew/acon/blob/main/README.md) for documentation.
EOF
)"

# Verify release was created
gh release view "v${VERSION}"
```

**Note**: GitHub automatically attaches source archives (tar.gz, zip) to releases. Homebrew builds from the tar.gz archive.

---

## Step 6: Update Homebrew Tap

Update the Homebrew formula with the new version:

```bash
# Navigate to homebrew-tap directory
cd ~/Projects/homebrew-tap
git pull origin main

# Display tarball info from Step 5
echo "Tarball URL: $TARBALL_URL"
echo "Tarball SHA256: $TARBALL_SHA256"

# Edit Formula/acon.rb and update:
# 1. url line: Update version in URL
# 2. sha256 line: Update with TARBALL_SHA256
# 3. ldflags: Update version in "-X main.version=X.X.X"
# 4. test: Update expected version in assert_match

# After editing, commit and push
git add Formula/acon.rb
git commit -m "acon: update to ${VERSION}"
git push origin main

# Return to acon directory
cd -
```

**Formula example** (Formula/acon.rb):

```ruby
class Acon < Formula
  desc "CLI for Confluence - because the web editor is not it"
  homepage "https://github.com/grantcarthew/acon"
  url "https://github.com/grantcarthew/acon/archive/refs/tags/v1.0.0.tar.gz"
  sha256 "abc123..."  # Use TARBALL_SHA256 value
  license "MIT"

  depends_on "go" => :build

  def install
    ENV["CGO_ENABLED"] = "0"
    system "go", "build", *std_go_args(ldflags: "-X main.version=1.0.0", output: bin/"acon")
  end

  test do
    assert_match "1.0.0", shell_output("#{bin}/acon --version")
  end
end
```

---

## Step 7: Verify Installation

Test the Homebrew installation:

```bash
# Update and reinstall
brew update
brew reinstall grantcarthew/tap/acon

# Verify version
acon --version  # Should show new version

# Test basic functionality (requires env vars)
export CONFLUENCE_BASE_URL="https://your-instance.atlassian.net"
export CONFLUENCE_EMAIL="your-email@example.com"
# Set token if available
acon space list -l 1
```

**Expected results**:

- `acon --version` displays new version
- No errors during installation

**If installation fails**, debug with:

```bash
brew audit --strict grantcarthew/tap/acon
brew install --verbose grantcarthew/tap/acon
```

---

## Step 8: Clean Up

Complete the release:

```bash
# Verify release is live
gh release view "v${VERSION}"

# Check Homebrew tap was updated
cd ~/Projects/homebrew-tap
git log -1
cd -

# Verify clean state
git status
```

**Release is complete!**

Monitor for issues:

- Watch GitHub issues for bug reports
- Monitor Homebrew installation feedback
- Be ready to release a patch if critical issues arise

---

## Rollback Procedure

If critical issues are discovered after release:

**Option 1: Patch Release** (Recommended)

```bash
# Fix the issue, then release patch version (e.g., v1.0.1)
# Follow the standard release process
```

**Option 2: Delete Release** (Last resort - use only for critical security issues)

```bash
# Delete GitHub release
gh release delete "v${VERSION}" --yes

# Delete tags
git push origin --delete "v${VERSION}"
git tag -d "v${VERSION}"

# Revert Homebrew tap
cd ~/Projects/homebrew-tap
git revert HEAD
git push origin main
cd -
```

---

## Quick Reference

One-command release workflow:

```bash
# Set version
export VERSION="1.0.0"

# Get previous version for change summary
PREV_VERSION=$(git tag -l | tail -1)

# 1. Validation
git status  # Should be clean

# 2. Create tag with summary
git log ${PREV_VERSION}..HEAD --oneline  # Review changes
SUMMARY="Your summary here"
git tag -a "v${VERSION}" -m "Release v${VERSION} - ${SUMMARY}"
git push origin "v${VERSION}"

# 3. Create GitHub Release
gh release create "v${VERSION}" --title "Release v${VERSION}" \
  --notes "$(git log ${PREV_VERSION}..v${VERSION} --pretty=format:'- %s')"

# 4. Get tarball SHA256
TARBALL_SHA256=$(curl -sL "https://github.com/grantcarthew/acon/archive/refs/tags/v${VERSION}.tar.gz" | shasum -a 256 | cut -d' ' -f1)
echo "SHA256: $TARBALL_SHA256"

# 5. Update Homebrew (edit Formula/acon.rb with VERSION and SHA256)
cd ~/Projects/homebrew-tap
# Edit Formula/acon.rb
git add Formula/acon.rb
git commit -m "acon: update to ${VERSION}"
git push origin main
cd -

# 6. Test
brew update && brew reinstall grantcarthew/tap/acon
acon --version
```

---

## Troubleshooting

**Tarball not available**

- Wait 1-2 minutes after pushing tag
- Verify tag exists: `git ls-remote --tags origin | grep v${VERSION}`
- Check: https://github.com/grantcarthew/acon/tags

**Homebrew formula issues**

- Audit: `brew audit --strict grantcarthew/tap/acon`
- Common: Incorrect SHA256, wrong URL format, Ruby syntax
- Fix and push updated formula

**Installation fails**

- Verbose output: `brew install --verbose grantcarthew/tap/acon`
- View formula: `brew cat grantcarthew/tap/acon`
- Verify tarball: `curl -I https://github.com/grantcarthew/acon/archive/refs/tags/v${VERSION}.tar.gz`

---

## Related Documents

- `README.md` - User-facing documentation
- `docs/code-review.md` - Code review checklist

---

**End of Release Process**
