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

   **Creating a Classic Personal Access Token**:
   
   a. Go to GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)
      - Direct link: https://github.com/settings/tokens
   
   b. Click "Generate new token" → "Generate new token (classic)"
   
   c. Configure the token:
      - **Note**: "raypaste-cli auto-release" (or any descriptive name)
      - **Expiration**: Choose based on your preference (90 days, 1 year, or no expiration)
      - **Select scopes** (checkboxes):
        - ✅ `repo` (Full control of private repositories)
          - This grants `contents: write` access
          - This includes: `repo:status`, `repo_deployment`, `public_repo`, `repo:invite`, `security_events`
        - ✅ `workflow` (Update GitHub Action workflows)
          - This grants `actions: write` access
          - Required to trigger the release workflow from another workflow
      
      **What you should see in the UI**:
      ```
      Select scopes
      
      [✓] repo               Full control of private repositories
          [✓] repo:status    Access commit status
          [✓] repo_deployment Access deployment status
          [✓] public_repo    Access public repositories
          [✓] repo:invite    Access repository invitations
          [✓] security_events Read and write security events
      
      [✓] workflow           Update GitHub Action workflows
      ```
   
   d. Click "Generate token" at the bottom
   
   e. **Important**: Copy the token immediately (it won't be shown again)
   
   **Adding the token as a repository secret**:
   
   a. Go to your repository → Settings → Secrets and variables → Actions
      - Direct link: https://github.com/raypaste/raypaste-cli/settings/secrets/actions
   
   b. Click "New repository secret"
   
   c. Configure the secret:
      - **Name**: `RELEASE_PAT` (must be exactly this name)
      - **Secret**: Paste the token you copied
   
   d. Click "Add secret"
   
   e. The workflow will automatically use this token for tag pushes, which will trigger the release workflow

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

If the release workflow doesn't trigger automatically (before `RELEASE_PAT` is configured):

### Option 1: Configure RELEASE_PAT (Recommended - one-time setup)

Follow the instructions in the "Solutions" section above to create and add the `RELEASE_PAT` secret. Once configured, future releases will work automatically.

### Option 2: Manually Trigger Release Workflow

For immediate release (or if PAT setup is not desired):

1. Navigate to: https://github.com/raypaste/raypaste-cli/actions/workflows/release.yml
2. Click "Run workflow" button (top right)
3. Ensure branch is set to: `main`
4. Enter the tag in the input field (e.g., `v0.2.1`)
5. Click the green "Run workflow" button

**Note**: The workflow file must be merged to main before manual dispatch will work.

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

### Token Permissions Explained

Classic Personal Access Tokens use different scope names than fine-grained permissions:

| Required Permission | Classic Token Scope | What it enables |
|---------------------|---------------------|-----------------|
| `contents: write` | ✅ `repo` | Push tags, create releases, modify repository contents |
| `actions: write` | ✅ `workflow` | Trigger workflow runs from other workflows |

**Why both are needed**:
- `repo` scope: Allows the workflow to push git tags
- `workflow` scope: Ensures tag pushes from the workflow can trigger other workflows

Without the `workflow` scope, tags pushed by the workflow won't trigger the release workflow due to GitHub's security policy preventing recursive workflow triggers.

### Release workflow didn't trigger after tag push

**Cause**: Tags pushed by `GITHUB_TOKEN` don't trigger workflows

**Fix**: Configure `RELEASE_PAT` secret OR manually trigger the release workflow

### "Tag already exists" error

**Cause**: The version in CHANGELOG.md matches an existing tag

**Fix**: Bump the version in CHANGELOG.md to a new version number

### GoReleaser fails on macOS

**Cause**: CGO dependencies for clipboard support require specific build flags

**Fix**: Check `.goreleaser.yml` for proper CGO flags and macOS SDK configuration
