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
    github-token: ${{ secrets.GH_AUDIT_TOKEN }}
    allowed:
      - get_file_contents
      - search_issues
      - search_pull_requests
      - list_commits
  bash:

safe-outputs:
  jobs:
    update-audit-state:
      description: "Update audit state file via PR"
      runs-on: ubuntu-latest
      output: "State updated successfully!"
      inputs:
        state_content:
          description: "JSON content for .github/audit-state.json"
          required: true
          type: string
        username:
          description: "Username just audited"
          required: true
          type: string
      permissions:
        contents: write
        pull-requests: write
      steps:
        - uses: actions/checkout@v4
        - name: Update state file
          run: |
            mkdir -p .github
            echo '${{ inputs.state_content }}' > .github/audit-state.json
        - name: Create PR with state update
          uses: peter-evans/create-pull-request@v6
          with:
            commit-message: "chore: update maintainer audit progress to @${{ inputs.username }}"
            branch: "audit-state-${{ inputs.username }}-${{ github.run_number }}"
            title: "chore: audit state update for @${{ inputs.username }}"
            body: |
              Automated audit state update.

              **Maintainer:** @${{ inputs.username }}
              **Run:** ${{ github.run_number }}
            delete-branch: true
            add-paths: |
              .github/audit-state.json

    send-maintainer-email:
      description: "Send maintainer audit report via Postmark"
      runs-on: ubuntu-latest
      output: "Email sent successfully!"
      inputs:
        subject:
          description: "Email subject line"
          required: true
          type: string
        markdown_body:
          description: "Markdown email content"
          required: true
          type: string
        username:
          description: "GitHub username being audited"
          required: true
          type: string
        email:
          description: "Maintainer's email address"
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
          with:
            script: |
              const fs = require('fs');
              const postmarkToken = process.env.POSTMARK_API_TOKEN;
              const fromEmail = process.env.POSTMARK_FROM_EMAIL;
              const isStaged = process.env.GH_AW_SAFE_OUTPUTS_STAGED === 'true';
              const outputContent = process.env.GH_AW_AGENT_OUTPUT;

              if (!postmarkToken || !fromEmail) {
                core.setFailed('Missing Postmark secrets: POSTMARK_API_TOKEN, POSTMARK_FROM_EMAIL');
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
                const { subject, markdown_body, username, email } = item;
                
                if (!subject || !markdown_body || !username || !email) {
                  core.warning(`Email ${i + 1}: Missing required fields, skipping`);
                  continue;
                }
                
                if (isStaged) {
                  let summaryContent = "## 📧 Staged Mode: Email Preview\n\n";
                  summaryContent += "**Subject:** " + subject + "\n\n";
                  summaryContent += "**To:** " + email + "\n\n";
                  summaryContent += "**Username:** @" + username + "\n\n";
                  summaryContent += "**Markdown Preview (first 500 chars):**\n\n```markdown\n" + markdown_body.substring(0, 500) + "...\n```\n\n";
                  await core.summary.addRaw(summaryContent).write();
                  core.info("📝 Email preview written to step summary");
                  continue;
                }
                
                core.info(`Sending email ${i + 1}/${emailItems.length} to ${email} (@${username})`);
                
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
                      To: email,
                      Subject: subject,
                      TextBody: markdown_body,
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

Generate a personalized maintainer audit report for **clubanderson** (`andy@clubanderson.com`).

## Task Overview

Analyze clubanderson's contributions over the last 60 days and create an encouraging email report with personalized recommendations. This is a **single-pass analysis** - gather all data once, generate the email, output the safe-output JSON, then stop.

## Required Data Collection

Calculate dates once at the start:
- `date_60_days_ago`: 60 days before today (YYYY-MM-DD)
- `date_180_days_ago`: 180 days before today (for interest analysis)
- `date_30_days_ago`: 30 days before today (for repo health)

Then collect the following data **one time only** using GitHub searches:

### 1. Interest Analysis (Last 6 Months)
Search once: `org:kubestellar is:pr is:merged author:clubanderson merged:>={date_180_days_ago}`
- Examine 10-15 recent PRs
- Detect patterns: CI/CD, docs, UI, testing, DevOps
- Note favorite repos

### 2. Current Metrics (Last 60 Days)
- **Help-wanted issues created**: Search once: `org:kubestellar is:issue label:"help wanted" author:clubanderson created:>={date_60_days_ago}` (count total)
- **Unique PRs commented**: Search once for merged + once for open, deduplicate PR numbers:
  - `org:kubestellar is:pr is:merged commenter:clubanderson updated:>={date_60_days_ago}`
  - `org:kubestellar is:pr is:open commenter:clubanderson updated:>={date_60_days_ago}`
- **Merged PRs authored**: Search once: `org:kubestellar is:pr is:merged author:clubanderson merged:>={date_60_days_ago}` (count total)

### 3. Opportunities
- **Help-wanted issues**: Search once: `org:kubestellar is:issue is:open label:"help wanted"` (pick top 3 matching interests)
- **PRs needing review**: Search once: `org:kubestellar is:pr is:open review:required` (pick top 3 matching expertise)

### 4. Repo Health (for help-wanted suggestions)
Use the **existing search results** from steps 1-3 above to identify:
- Which repos clubanderson is most active in
- Areas with low recent activity (cold spots)
- Don't perform additional searches - analyze the data you already have

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

## Evaluation Criteria

Compare collected metrics against requirements:
- ✅ Help-wanted issues ≥ 2
- ✅ Unique PRs commented ≥ 8  
- ✅ Merged PRs ≥ 3

## Output Format

Generate a personalized Markdown email following this structure (NO markdown headings `##`):

```markdown
Hey @clubanderson! 👋

Here's your KubeStellar impact snapshot for the last 60 days.

---

📊 **Quick Stats**

✅ **Help-Wanted Issues:** X created (required: ≥2)  
✅/❌ **PR Reviews:** Y unique PRs (required: ≥8)  
✅ **PRs Merged:** Z merged (required: ≥3)

**Overall:** Pass/Fail

---

🎯 **Your Impact Areas**

[Detected interests from interest analysis]

---

🌟 **Where You Can Help Next**

🏷️ **Help-Wanted Issues Perfect For You**
[Top 3 with links]

👀 **PRs That Need Your Review**
[Top 3 with links]

🎯 **Repos Looking for Your Skills**
[Top 3 with reasons]

---

✨ **Consider Creating Help-Wanted Issues**

[Top 3 suggestions based on repo health + expertise]

---

💪 **[Encouraging closing based on pass/fail]**

---

_Automated by GitHub Agentic Workflows • {timestamp}_
```

## Final Step: Output Safe-Output

Create a **valid JSON object** for the safe-output system. Use the `safeoutputs` tool to output:

```json
{
  "type": "send_maintainer_email",
  "subject": "🌟 Your KubeStellar Impact Report - @clubanderson",
  "markdown_body": "<YOUR_GENERATED_EMAIL_HERE>",
  "username": "clubanderson",
  "email": "andy@clubanderson.com"
}
```

**Important**:
- Replace `<YOUR_GENERATED_EMAIL_HERE>` with the actual email markdown you generated
- Escape any quotes or special characters in the markdown_body
- Output this as a single valid JSON object
- Do NOT output intermediate JSON attempts

**After successfully outputting this JSON object, stop immediately.**

