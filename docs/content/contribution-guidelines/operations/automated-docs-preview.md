# Automated Documentation PR Preview System

This document explains the automated documentation preview system implemented to address [GitHub Issue #3189](https://github.com/kubestellar/kubestellar/issues/3189).

## Overview

The automated preview system eliminates the manual process of creating and including preview links in documentation PRs. When contributors create PRs that modify documentation from their forks, the system automatically:

1. **Detects documentation changes** in PR from forks
2. **Generates and adds preview links** to PR descriptions automatically  
3. **Triggers documentation builds** on the contributor's fork (for `doc-` branches)
4. **Provides guidance** for optimal setup

## How It Works

### For Contributors (Fork-based PRs)

**Recommended workflow:**

1. **Fork KubeStellar repository** (make sure to include the `gh-pages` branch)
2. **Create a branch starting with `doc-`**:
   ```bash
   git checkout -b doc-my-documentation-fix
   ```
3. **Make your documentation changes** in the `docs/` directory
4. **Push your branch** to your fork
5. **Create a Pull Request** - automation handles the rest!

**What happens automatically:**

- ✅ Preview link added to PR description
- 🔄 Docs build triggered on your fork (for `doc-` branches)
- 📝 Setup instructions provided if needed
- 🤖 Bot comment added with preview status

**Preview URL format:** `https://your-username.github.io/kubestellar/your-branch-name`

### Workflow Components

The automated system consists of two main GitHub Actions workflows:

#### 1. `auto-docs-preview.yml` (NEW)
- **Triggers:** When documentation PRs are opened/updated from forks
- **Functions:**
  - Detects if PR is from a fork and modifies documentation
  - Automatically adds preview section to PR description
  - Attempts to trigger docs build on fork for `doc-` branches
  - Adds informative bot comments

#### 2. `check-docs-pr-preview.yml` (UPDATED)
- **Triggers:** When documentation PRs are opened/updated  
- **Functions:**
  - Validates that preview links are present (now more lenient for forks)
  - Provides different validation for fork PRs vs direct PRs
  - Offers guidance based on branch naming conventions

## Branch Naming Convention

For automatic preview generation, name your branch starting with `doc-`:

✅ **Good examples:**
- `doc-fix-typos`
- `doc-new-section`
- `doc-api-updates`
- `doc-123-issue-fix`

❌ **Won't auto-generate:**
- `fix-docs`
- `feature/documentation`  
- `update-readme`

**Note:** Even if your branch doesn't start with `doc-`, the system will still add preview links and provide setup instructions.

## Fork Setup Requirements

For the automation to work properly, your fork needs:

1. **Include `gh-pages` branch** when forking
2. **Enable GitHub Pages** in your fork's settings
3. **Set GitHub Pages source** to the `gh-pages` branch

### Setting Up Your Fork

If you already have a fork without `gh-pages`:

```bash
# Add upstream remote (if not already added)
git remote add upstream https://github.com/kubestellar/kubestellar.git

# Fetch and push gh-pages branch to your fork
git fetch upstream gh-pages
git checkout upstream/gh-pages
git push -f origin gh-pages

# Go to your fork's Settings → Pages
# Set Source to "Deploy from branch" and select "gh-pages"
```

## Preview Link Formats

The automation generates preview links in this format:

**PR Description:**
```markdown
## 📖 Documentation Preview

A preview of the documentation changes will be available at:
**Preview: https://your-username.github.io/kubestellar/your-branch-name**

*Note: The preview will be automatically generated after the docs build completes on your fork. This may take a few minutes.*

✅ Your branch name (`doc-branch-name`) follows the `doc-` naming convention for automatic preview generation.
```

**Bot Comment:**
```markdown
## 📖 Documentation Preview Generated

✅ Your branch name (`doc-branch-name`) follows the `doc-` naming convention.

🔄 A preview is being generated at: **https://your-username.github.io/kubestellar/doc-branch-name**

⏱️ The preview should be available within a few minutes after the docs build completes on your fork.
```

## Troubleshooting

### Preview Not Generated

If your preview isn't generated automatically:

1. **Check branch name** - does it start with `doc-`?
2. **Verify fork setup** - is `gh-pages` branch present and GitHub Pages enabled?
3. **Check Actions tab** - is the "Generate and push docs" workflow running on your fork?
4. **Manual trigger** - you can manually run the docs workflow on your fork

### Preview Link Not Added

If the preview link isn't automatically added to your PR:

1. **Verify it's a fork PR** - automation only works for fork-based PRs
2. **Check file paths** - does your PR modify files in the documentation paths?
3. **Review bot comments** - check for any error messages or guidance

### Manual Fallback

If automation fails, you can always:

1. **Manually trigger** the "Generate and push docs" workflow on your fork
2. **Add the preview link** manually to your PR description
3. **Follow the existing documentation** in [document-management.md](./document-management.md)

## Technical Implementation

### File Changes Made

1. **New workflow:** `.github/workflows/auto-docs-preview.yml`
   - Handles automatic preview generation for fork PRs
   - Uses `actions/github-script` to update PR descriptions and add comments
   - Attempts to trigger docs workflows on forks when possible

2. **Updated workflow:** `.github/workflows/check-docs-pr-preview.yml`
   - Modified validation logic to work with automated system
   - Different handling for fork PRs vs direct PRs
   - More lenient validation when automation is expected to handle previews

3. **Updated template:** `.github/pull_request_template.md`
   - Added information about automatic preview generation
   - Clarified different processes for fork vs direct PRs

4. **Updated documentation:** `docs/content/contribution-guidelines/operations/document-management.md`
   - Added new sections explaining automated process
   - Updated existing sections to reflect new workflow
   - Added quick-start guide for contributors

### Security Considerations

- Uses approved GitHub Actions from `.gha-reversemap.yml`
- Only modifies PR descriptions and adds comments (no code execution)
- Respects GitHub's permissions model for cross-repository access
- Gracefully handles failures when fork access is restricted

## Benefits

### For Contributors
- ✅ **Simplified workflow** - no manual preview link creation
- ✅ **Automatic setup** - preview links added automatically
- ✅ **Clear guidance** - instructions provided when needed
- ✅ **Immediate feedback** - instant preview status in PR

### For Maintainers  
- ✅ **Consistent previews** - all doc PRs get preview links
- ✅ **Reduced review overhead** - no need to ask for previews
- ✅ **Better quality control** - preview validation is automated
- ✅ **Enhanced collaboration** - easier to review doc changes

### For the Project
- ✅ **Improved documentation quality** - easier preview process encourages better docs
- ✅ **Faster review cycles** - previews available immediately  
- ✅ **Lower barrier to contribution** - simplified process for new contributors
- ✅ **Consistent standards** - automated enforcement of preview requirements

## Future Enhancements

Potential improvements that could be added:

1. **Preview status updates** - track when previews are ready
2. **Direct preview links** in comments when builds complete
3. **Preview diff highlighting** - show what changed in the preview
4. **Integration with review process** - link previews to approval workflow
5. **Automatic cleanup** - remove old preview versions after PR merge

## Related Documentation

- [Document Management Guide](./document-management.md)
- [Contributing Guidelines](../CONTRIBUTING.md)
- [GitHub Issue #3189](https://github.com/kubestellar/kubestellar/issues/3189)