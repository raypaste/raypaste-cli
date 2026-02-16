# GitHub Actions Workflows

## Overview

This repository uses a three-stage automated release process:

1. **PR Validation** (`check-version.yml`) - Validates that CHANGELOG.md has a new version entry
2. **Tag Creation** (`tag-release-on-merge.yml`) - Creates and pushes a git tag when PR merges to main
3. **Release Build** (`release.yml`) - Builds binaries and creates GitHub release for the new tag

## Workflow Details

### Check Version (`check-version.yml`)

**Trigger**: Pull requests to main  
**Purpose**: Ensures CHANGELOG.md has a valid, new version entry before merge

### Tag Release On Merge (`tag-release-on-merge.yml`)

**Trigger**: Push to main branch  
**Purpose**: Automatically creates and pushes a git tag based on CHANGELOG.md version

**Important**: By default, GitHub prevents workflows triggered by `GITHUB_TOKEN` from starting other workflows (to prevent recursive triggers). This means the release workflow may not automatically trigger.

**Solutions**:

1. **Recommended**: Add a `RELEASE_PAT` secret
   - Create a Personal Access Token with `contents: write` and `actions: write` scopes
   - Add it as a repository secret named `RELEASE_PAT`
   - The workflow will use this token to push tags, which will trigger the release workflow

2. **Fallback**: Manual workflow dispatch
   - If no `RELEASE_PAT` is configured, the workflow attempts to manually trigger the release
   - This requires the default `GITHUB_TOKEN` to have workflow trigger permissions
   - If this fails, manually trigger the release workflow from the Actions tab

### Release (`release.yml`)

**Triggers**:
- Push of tags matching `v*` (when triggered by PAT)
- Manual workflow dispatch (fallback mechanism)

**Purpose**: Builds cross-platform binaries and creates GitHub release with changelog notes

**Inputs** (for manual dispatch):
- `tag`: The tag to release (e.g., `v0.2.1`)

## Manual Release (Workaround)

If the release workflow doesn't trigger automatically:

1. Go to Actions → Release workflow
2. Click "Run workflow"
3. Select branch: `main`
4. Enter the tag (e.g., `v0.2.1`)
5. Click "Run workflow"

## Testing Changes

When modifying these workflows, test in a feature branch:

```bash
# Create test branch
git checkout -b test-workflow-changes

# Make your changes to .github/workflows/*

# Test by creating a test tag
git tag -a v0.0.0-test -m "Test release"
git push origin v0.0.0-test

# Check Actions tab to verify workflows run correctly

# Clean up
git tag -d v0.0.0-test
git push --delete origin v0.0.0-test
```

## Troubleshooting

### Release workflow didn't trigger after tag push

**Cause**: Tags pushed by `GITHUB_TOKEN` don't trigger workflows

**Fix**: Configure `RELEASE_PAT` secret OR manually trigger the release workflow

### "Tag already exists" error

**Cause**: The version in CHANGELOG.md matches an existing tag

**Fix**: Bump the version in CHANGELOG.md to a new version number

### GoReleaser fails on macOS

**Cause**: CGO dependencies for clipboard support require specific build flags

**Fix**: Check `.goreleaser.yml` for proper CGO flags and macOS SDK configuration
