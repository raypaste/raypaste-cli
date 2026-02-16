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

**Requirements**:

- Top `CHANGELOG.md` release heading must match `## [X.Y.Z] - YYYY-MM-DD`
- The derived tag `vX.Y.Z` must not already exist

**Important**: By default, GitHub prevents workflows triggered by `GITHUB_TOKEN` from starting other workflows. This means the `Release` workflow may not automatically trigger when tag push is done with the default token.

**Solutions**:

1. **Recommended**: Add a `RELEASE_PAT` secret
   - Create a Personal Access Token with `contents: write` and `actions: write` scopes
   - Add it as a repository secret named `RELEASE_PAT`
   - The workflow will use this token to push tags, which will trigger the release workflow

2. **Fallback**: Manual workflow dispatch
   - If no `RELEASE_PAT` is configured, the workflow attempts to manually trigger `release.yml`
   - This requires the default `GITHUB_TOKEN` to have workflow trigger permissions
   - If this fails, manually trigger the `Release` workflow from the Actions tab

### Release (`release.yml`)

**Triggers**:
- Push of tags matching `v*` (when triggered by PAT)
- Manual workflow dispatch (fallback mechanism)

**Purpose**: Builds cross-platform binaries and creates GitHub release with changelog notes

**Inputs** (for manual dispatch):
- `tag`: The tag to release (e.g., `v0.2.1`)

## Manual Release (Fallback)

If the release workflow doesn't trigger automatically:

1. Go to Actions → Release workflow
2. Click "Run workflow"
3. Select branch: `main`
4. Enter the tag (e.g., `v0.2.1`)
5. Click "Run workflow"

If manual release fails with `A branch or tag with the name 'vX.Y.Z' could not be found`, create and push the missing tag first:

```bash
git checkout main
git pull
git tag vX.Y.Z
git push origin vX.Y.Z
```

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

### Tag was not created after merge to main

**Cause**: `tag-release-on-merge.yml` failed (for example, invalid workflow syntax or changelog/version guard failure)

**Fix**:

1. Check the failed `tag-release-on-merge` run on `main`
2. Fix workflow/changelog issue and merge to `main`
3. Backfill the tag:
   ```bash
   git checkout main
   git pull
   git tag vX.Y.Z
   git push origin vX.Y.Z
   ```

### "Tag already exists" error

**Cause**: The version in CHANGELOG.md matches an existing tag

**Fix**: Bump the version in CHANGELOG.md to a new version number

### Invalid workflow file error (run fails in 0s)

**Cause**: YAML parse/evaluation error in the workflow file itself

**Fix**: Correct the workflow on a branch and merge to `main`; then re-run by pushing a new commit or backfill the missed tag if needed

### GoReleaser fails on macOS

**Cause**: CGO dependencies for clipboard support require specific build flags

**Fix**: Check `.goreleaser.yml` for proper CGO flags and macOS SDK configuration
