---
description: |
  Automated maintainer audit workflow that cycles through maintainers in round-robin
  fashion, auditing their contributions across KubeStellar repositories. Runs every 6
  hours and uses cache-memory to track progress. Generates HTML email reports via
  Postmark with activity metrics for the last 60 days.

on:
  schedule:
    # Every 6 hours: 3am, 9am, 3pm, 9pm UTC
    - cron: "0 3,9,15,21 * * *"
  workflow_dispatch:

permissions: read-all

tools:
  github:
    read-only: true
  bash:
  cache-memory: true
  edit:

safe-outputs:
  jobs:
    send-maintainer-email:
      description: "Send maintainer audit report via Postmark"
      runs-on: ubuntu-latest
      output: "Email sent successfully!"
      inputs:
        subject:
          description: "Email subject line"
          required: true
          type: string
        html_body:
          description: "HTML email content"
          required: true
          type: string
        username:
          description: "GitHub username being audited"
          required: true
          type: string
      permissions:
        contents: read
      steps:
        - name: Send email via Postmark
          uses: actions/github-script@v7
          env:
            POSTMARK_API_TOKEN: "${{ secrets.POSTMARK_API_TOKEN }}"
            POSTMARK_FROM_EMAIL: "${{ secrets.POSTMARK_FROM_EMAIL }}"
            POSTMARK_TO_EMAIL: "${{ secrets.POSTMARK_TO_EMAIL }}"
          with:
            script: |
              const fs = require('fs');
              const postmarkToken = process.env.POSTMARK_API_TOKEN;
              const fromEmail = process.env.POSTMARK_FROM_EMAIL;
              const toEmail = process.env.POSTMARK_TO_EMAIL;
              const isStaged = process.env.GH_AW_SAFE_OUTPUTS_STAGED === 'true';
              const outputContent = process.env.GH_AW_AGENT_OUTPUT;
              
              if (!postmarkToken || !fromEmail || !toEmail) {
                core.setFailed('Missing Postmark secrets: POSTMARK_API_TOKEN, POSTMARK_FROM_EMAIL, POSTMARK_TO_EMAIL');
                return;
              }
              
              if (!outputContent) {
                core.info('No GH_AW_AGENT_OUTPUT environment variable found');
                return;
              }
              
              let agentOutputData;
              try {
                const fileContent = fs.readFileSync(outputContent, 'utf8');
                agentOutputData = JSON.parse(fileContent);
              } catch (error) {
                core.setFailed(`Error reading agent output: ${error instanceof Error ? error.message : String(error)}`);
                return;
              }
              
              if (!agentOutputData.items || !Array.isArray(agentOutputData.items)) {
                core.info('No valid items in agent output');
                return;
              }
              
              const emailItems = agentOutputData.items.filter(item => item.type === 'send_maintainer_email');
              
              if (emailItems.length === 0) {
                core.info('No send_maintainer_email items found');
                return;
              }
              
              core.info(`Found ${emailItems.length} email(s) to send`);
              
              for (let i = 0; i < emailItems.length; i++) {
                const item = emailItems[i];
                const { subject, html_body, username } = item;
                
                if (!subject || !html_body || !username) {
                  core.warning(`Email ${i + 1}: Missing required fields, skipping`);
                  continue;
                }
                
                if (isStaged) {
                  let summaryContent = "## 📧 Staged Mode: Email Preview\n\n";
                  summaryContent += "**Subject:** " + subject + "\n\n";
                  summaryContent += "**Username:** @" + username + "\n\n";
                  summaryContent += "**HTML Preview (first 500 chars):**\n\n```html\n" + html_body.substring(0, 500) + "...\n```\n\n";
                  await core.summary.addRaw(summaryContent).write();
                  core.info("📝 Email preview written to step summary");
                  continue;
                }
                
                core.info(`Sending email ${i + 1}/${emailItems.length} for @${username}`);
                
                try {
                  const response = await fetch('https://api.postmarkapp.com/email', {
                    method: 'POST',
                    headers: {
                      'Accept': 'application/json',
                      'Content-Type': 'application/json',
                      'X-Postmark-Server-Token': postmarkToken
                    },
                    body: JSON.stringify({
                      From: fromEmail,
                      To: toEmail,
                      Subject: subject,
                      HtmlBody: html_body,
                      MessageStream: 'outbound'
                    })
                  });
                  
                  if (!response.ok) {
                    const errorText = await response.text();
                    core.setFailed(`Postmark API error: ${response.status} - ${errorText}`);
                    return;
                  }
                  
                  const result = await response.json();
                  core.info(`✅ Email sent! MessageID: ${result.MessageId}`);
                } catch (error) {
                  core.setFailed(`Failed to send email ${i + 1}: ${error instanceof Error ? error.message : String(error)}`);
                  return;
                }
              }
---

# Maintainer Audit Report

You are conducting automated maintainer audits for the KubeStellar organization.

## Your Mission

Audit ONE maintainer per run using a round-robin system, tracking progress in cache-memory.

## Maintainer List (Round-Robin Order)

The complete list of maintainers to audit:

1. clubanderson
<!-- 2. mikespreitzer
3. dumb0002
4. waltforme
5. pdettori
6. francostellari
7. kproche
8. nupurshivani
9. onkar717
10. kunal-511
11. mavrick-1
12. gaurab-khanal
13. naman9271
14. btwshivam
15. rxinui
16. vedansh-5
17. sagar2366
18. oksaumya
19. rupam-it -->

## Audit Criteria (Last 60 Days)

Calculate these metrics using GitHub search tools across KubeStellar org repositories:

### 1. Help-Wanted Issues Created
- **Requirement:** ≥ 2 issues
- **Search Query:** `org:kubestellar is:issue label:"help wanted" author:{username} created:>={date_60_days_ago}`
- Use `github.search_code` or `github.search_issues` to count results

### 2. Unique PRs Commented On  
- **Requirement:** ≥ 8 different PRs
- **Approach:** Search for PRs where user commented, deduplicate by PR number
- **Queries:** 
  - Merged: `org:kubestellar is:pr is:merged commenter:{username} updated:>={date_60_days_ago}`
  - Open: `org:kubestellar is:pr is:open commenter:{username} updated:>={date_60_days_ago}`
- Count unique PR numbers from both searches

### 3. PRs Merged
- **Requirement:** ≥ 3 merged PRs
- **Search Query:** `org:kubestellar is:pr is:merged author:{username} merged:>={date_60_days_ago}`

## KubeStellar Repositories

Scope the search to these repos:
- kubestellar/a2a
- kubestellar/docs
- kubestellar/homebrew-kubectl-multi
- kubestellar/kubectl-multi-plugin
- kubestellar/kubectl-rbac-flatten-plugin
- kubestellar/kubeflex
- kubestellar/kubestellar
- kubestellar/ocm-status-addon
- kubestellar/ocm-transport-plugin
- kubestellar/ui
- kubestellar/ui-plugins

## Process

### Step 1: Check Cache Memory
Load progress from `/tmp/gh-aw/cache-memory/maintainer_audit_progress.json`:
- If found, read `last_index` and increment by 1
- If not found or missing, start at index 0

### Step 2: Select Next Maintainer
- Get maintainer at current index from the list above (0-18)
- If index >= 19, wrap to 0

### Step 3: Calculate Date Range
Calculate the date 60 days ago from today in YYYY-MM-DD format.

### Step 4: Gather Metrics Using GitHub Search

Use the GitHub MCP tools to search and count:

**A. Help-Wanted Issues**
- Search: `org:kubestellar is:issue label:"help wanted" author:{username} created:>={date_60_days_ago}`
- Count total results

**B. Unique PRs Commented On**
- Search merged PRs: `org:kubestellar is:pr is:merged commenter:{username} updated:>={date_60_days_ago}`
- Search open PRs: `org:kubestellar is:pr is:open commenter:{username} updated:>={date_60_days_ago}`
- Extract PR numbers from both result sets
- Count unique PR numbers (deduplicate)

**C. Merged PRs Authored**
- Search: `org:kubestellar is:pr is:merged author:{username} merged:>={date_60_days_ago}`
- Count total results

### Step 5: Evaluate Criteria

Compare actual counts against requirements:
- Help-wanted issues: actual >= 2 ? ✅ PASS : ❌ FAIL
- Unique PRs commented: actual >= 8 ? ✅ PASS : ❌ FAIL  
- Merged PRs: actual >= 3 ? ✅ PASS : ❌ FAIL

Overall: PASS if all three criteria pass, otherwise FAIL

### Step 6: Generate Markdown Email

Create a professional Markdown-formatted email:

**Structure:**
- Header with KubeStellar branding
- Maintainer username and audit metadata
- Three metric sections with:
  - ✅ PASS or ❌ FAIL status
  - Actual count vs requirement
  - GitHub search link for verification
- Overall result summary
- Footer with automation note

**Markdown Format Benefits:**
- Cleaner, more readable in plain text
- Better email client compatibility
- Easier to generate and validate

### Step 7: Output Safe-Output Entry

Create a JSON entry for the email safe-output job:
```json
{
  "type": "send_maintainer_email",
  "subject": "Maintainer Audit Report - @{username} - {PASS/FAIL}",
  "html_body": "{markdown_converted_to_html}",
  "username": "{username}"
}
```

**Note:** Generate the email in Markdown format, then convert to HTML for Postmark compatibility (most email clients render HTML well).

### Step 8: Update Cache Memory

Write to `/tmp/gh-aw/cache-memory/maintainer_audit_progress.json`:
```json
{
  "last_index": {next_index},
  "last_username": "{username}",
  "last_audit_date": "{iso_timestamp}",
  "last_result": "PASS or FAIL"
}
```

Where `next_index` is the current index + 1 (or 0 if wrapping).

## Example Markdown Email Structure

```markdown
# 🔍 Maintainer Audit Report

**KubeStellar Organization**

---

## Audit for: @{username}

**Audit Date:** 2025-12-05  
**Period:** Last 60 days (since 2025-10-06)

---

### ✅ Help-Wanted Issues: PASS
Created **5** issues (required: ≥2)  
[View on GitHub](https://github.com/search?q=org%3Akubestellar+is%3Aissue+label%3A%22help+wanted%22+author%3A{username}+created%3A%3E%3D2025-10-06)

### ❌ Unique PRs Commented: FAIL
Commented on **6** unique PRs (required: ≥8)  
[View Merged PRs](https://github.com/search?q=org%3Akubestellar+is%3Apr+is%3Amerged+commenter%3A{username})

### ✅ Merged PRs: PASS
Merged **4** PRs (required: ≥3)  
[View on GitHub](https://github.com/search?q=org%3Akubestellar+is%3Apr+is%3Amerged+author%3A{username})

---

## Overall Result: ❌ FAIL

**Summary:** 2 of 3 criteria passed. The maintainer needs to increase PR review activity.

---

*Automated by GitHub Agentic Workflows • Run at 2025-12-05 20:00 UTC*
```

## Important Notes

- Generate the email content in **Markdown format** first
- Convert Markdown to simple HTML if needed for Postmark (use basic tags: h1, h2, p, a, strong, em)
- Keep formatting clean and email-client friendly
- Include clickable GitHub search links for transparency
- Use emojis sparingly for visual indicators (✅ ❌ 🔍)
