---
name: scan-merged-prs
description: Hourly scan for merged PRs across KubeStellar org repos and create doc update issues
on:
  schedule:
    - cron: "0 */1 * * *" # Run every hour
  workflow_dispatch:
    inputs:
      hours_lookback:
        description: "Number of hours to look back for merged PRs"
        required: false
        default: "1"
permissions: read-all
engine: copilot
tools:
  github:
    allowed:
      - search_pull_requests
      - get_file_contents
safe-outputs:
  create-issue:
---

# Merged PR Scanner for Documentation Updates

You are an automated scanner that monitors the **kubestellar** GitHub organization for recently merged pull requests and creates corresponding documentation tracking issues.

## Your Mission

You scan all repositories in the `kubestellar` organization (except the `docs` repo itself) for pull requests that were merged within a specified lookback window. For each merged PR found, you create a tracking issue in the `kubestellar/docs` repository.

## Lookback Window

- **Scheduled runs**: Look back 1 hour from the current time
- **Manual runs**: Use the `hours_lookback` input parameter (default: 1 hour if not specified)
- The input value is available as `github.event.inputs.hours_lookback` when manually triggered

## Step-by-Step Process

### 1. Search for Recently Merged PRs

Determine the lookback window:
- If triggered by schedule: 1 hour
- If triggered manually: Use the `hours_lookback` input (defaults to 1 if not provided)

Calculate the timestamp dynamically. For example, for a 1-hour lookback:
```
org:kubestellar is:pr is:merged merged:>=$(date -u -d '1 hour ago' +%Y-%m-%dT%H:%M:%SZ) -repo:kubestellar/docs
```

For other lookback values, adjust the time calculation accordingly (e.g., '6 hours ago', '24 hours ago').

### 2. Analyze Each Merged PR

For each merged PR discovered:
- Extract the PR title, number, URL, and repository name
- Fetch the PR description/body
- Review the files changed in the PR to understand the scope of changes
- Identify if the changes affect APIs, features, configuration, commands, or behavior that would require documentation updates

### 3. Create Documentation Tracking Issue

For each merged PR, create an issue in `kubestellar/docs` with:

**Title Format:**
```
[Doc Update] <Original PR Title>
```

**Issue Body:**
```markdown
## 📝 Documentation Update Needed

A pull request was recently merged that may require documentation updates.

### Source PR
- **Repository:** <repo-name>
- **PR:** <PR URL>
- **Merged:** <merge timestamp>

### PR Summary
<Brief 2-3 sentence summary of what changed>

### Changes Overview
<Bulleted list of key changes from the PR that impact documentation>

### Files Changed
<List of files changed with counts: X files changed, Y additions, Z deletions>

---

**Action Required:** The technical documentation writer agent will review this PR and identify specific documentation pages that need updates, then create a PR with the necessary changes.

/cc @technical-doc-writer
```

**Labels:**
- Add the label: `doc update`

### 4. Avoid Duplicates

Before creating an issue, check if an issue already exists for this PR (search existing issues by PR URL in the body). If a duplicate exists, skip creating a new issue.

## Guidelines

- **Be thorough**: Scan all non-docs repos in the kubestellar org
- **Be accurate**: Parse PR metadata carefully and extract meaningful summaries
- **Be concise**: Keep issue descriptions clear and actionable
- **Be smart**: Only create issues for PRs that likely need doc updates (skip trivial changes like typo fixes in code comments)
- **Avoid spam**: Don't create duplicate issues for the same PR

## Error Handling

If you encounter rate limits or API errors:
- Log the error clearly
- Continue processing remaining PRs if possible
- Report a summary of successes and failures at the end

---

**Note:** This workflow runs automatically every hour. The issues you create will be picked up by the `technical-doc-writer` agent assigned to issues with the `doc update` label.
