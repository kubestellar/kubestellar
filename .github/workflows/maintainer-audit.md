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

You are conducting automated maintainer audits for the KubeStellar organization.

## Your Mission

Audit ONE maintainer per run using a round-robin system, tracking progress in cache-memory.

## Maintainer List (Round-Robin Order)

The complete list of maintainers to audit with their email addresses:

| Index | Username       | Email                      |
| ----- | -------------- | -------------------------- |
| 0     | clubanderson   | andy@clubanderson.com      |
| 1     | mikespreitzer  | mspreitz@us.ibm.com        |
| 2     | dumb0002       | Braulio.Dumba@ibm.com      |
| 3     | waltforme      | jun.duan@ibm.com           |
| 4     | pdettori       | dettori@us.ibm.com         |
| 5     | francostellari | stellari@us.ibm.com        |
| 6     | kproche        | kproche@us.ibm.com         |
| 7     | nupurshivani   | nupurjha.me@gmail.com      |
| 8     | onkar717       | onkarwork2234@gmail.com    |
| 9     | kunal-511      | yoyokvunal@gmail.com       |
| 10    | mavrick-1      | mavrickrishi@gmail.com     |
| 11    | gaurab-khanal  | khanalgaurab98@gmail.com   |
| 12    | naman9271      | namanjain9271@gmail.com    |
| 13    | btwshivam      | shivam200446@gmail.com     |
| 14    | rxinui         | rainui.ly@gmail.com        |
| 15    | vedansh-5      | vedanshsaini7719@gmail.com |
| 16    | sagar2366      | sagarutekar2366@gmail.com  |
| 17    | oksaumya       | saumyakr2006@gmail.com     |
| 18    | rupam-it       | Mannarupam3@gmail.com      |

**Total: 19 maintainers**

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

### Step 1: Load State from Repository

Load progress from `.github/audit-state.json` in the repository:

- Use `github.get_file_contents` to read the file
- Parse JSON to get `last_index` and increment by 1
- If file doesn't exist or fails to load, start at index 0

### Step 2: Select Next Maintainer

- Get maintainer username and email at current index from the table above (0-18)
- If index >= 19, wrap to 0
- Store both `username` and `email` for later use

### Step 3: Calculate Date Range

Calculate the date 60 days ago from today in YYYY-MM-DD format.

### Step 4: Analyze Maintainer's Interests

Before gathering metrics, understand what the maintainer likes to work on:

**A. Analyze Past PRs (Last 6 months)**

- Search: `org:kubestellar is:pr is:merged author:{username} merged:>={date_180_days_ago}`
- For each PR (up to 10-15 recent ones):
  - Extract file paths changed (look for patterns like `/docs/`, `/src/`, `*_test.*`, `.yaml`, etc.)
  - Note PR labels (e.g., `documentation`, `bug`, `feature`, `testing`)
  - Track which repos they contribute to most
- **Detect patterns:**
  - **Documentation focus:** Many changes in `/docs/`, `*.md` files
  - **Frontend/UI focus:** Changes in `/ui/`, `/src/components/`, `*.tsx`, `*.css`
  - **Backend/Core:** Changes in `/pkg/`, `/cmd/`, `*.go`, `*.java`
  - **Testing:** Changes in `*_test.*`, `/tests/`, `*_spec.*`
  - **DevOps/CI:** Changes in `.github/workflows/`, `Dockerfile`, `*.yaml`
  - **Favorite repos:** Count contributions per repo

### Step 5: Gather Current Metrics

Use the GitHub MCP tools to search and count:

**A. Help-Wanted Issues Created**

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

### Step 6: Find Personalized Opportunities

Based on their interest patterns, find **Top 3** opportunities in each category:

**A. Help-Wanted Issues in Their Favorite Areas**

- Search across all KubeStellar repos: `org:kubestellar is:issue is:open label:"help wanted"`
- Filter/rank by:
  - Repos they've contributed to before (higher priority)
  - File paths/labels matching their detected interests
  - Recent activity (updated within last 14 days)
- Select top 3 issues with direct links

**B. PRs Needing Review in Their Expertise Areas**

- Search: `org:kubestellar is:pr is:open review:required`
- Filter/rank by:
  - Repos they're active in
  - File paths matching their interest patterns (docs, backend, frontend, etc.)
  - PRs without recent review activity
- Select top 3 PRs with direct links

**C. Repos That Could Use Their Skills**

- Based on their detected focus areas (docs, backend, UI, testing, DevOps):
  - Match to KubeStellar repos that align (e.g., docs contributor → kubestellar/docs)
  - Note recent issues/PRs in those repos needing attention
- Suggest top 3 repos with reasons why

### Step 7: Evaluate Criteria

Compare actual counts against requirements:

- Help-wanted issues: actual >= 2 ? ✅ PASS : ❌ FAIL
- Unique PRs commented: actual >= 8 ? ✅ PASS : ❌ FAIL
- Merged PRs: actual >= 3 ? ✅ PASS : ❌ FAIL

Overall: PASS if all three criteria pass, otherwise FAIL

### Step 8: Generate Personalized Markdown Email

Create an **encouraging, actionable** Markdown email:

**Structure:**

1. **Warm greeting** with maintainer's username
2. **Quick stats summary** (metric results with ✅/❌)
3. **"Your Impact Areas"** section:
   - Detected interests from past PRs (e.g., "You love working on documentation and testing!")
   - Most active repos
4. **"Where You Can Help Next"** - Personalized recommendations:
   - **🏷️ Help-Wanted Issues for You:** Top 3 issues matching their interests with direct links
   - **👀 PRs That Need Your Review:** Top 3 PRs in their expertise areas with direct links
   - **🎯 Repos Looking for Your Skills:** Top 3 repos with explanations
5. **Encouraging closing:**
   - If PASS: Celebrate their contributions and suggest maintaining momentum
   - If FAIL: Focus on opportunities, frame as "here's how to get back on track"
6. **Footer:** Automation note with timestamp

**Tone:** Supportive and constructive, not punitive

### Step 9: Output Safe-Output Entry

Create a JSON entry for the email safe-output job:

```json
{
  "type": "send_maintainer_email",
  "subject": "🌟 Your KubeStellar Impact Report - @{username}",
  "markdown_body": "{pure_markdown_content}",
  "username": "{username}",
  "email": "{email}"
}
```

Where `email` is the maintainer's email address from the table in Step 2.

**Note:** Keep it as pure Markdown - no HTML conversion needed. Postmark will send as plain text.

### Step 10: Output State Update

Create a safe-output entry to update the state file:

```json
{
  "type": "update_audit_state",
  "state_content": "{\"last_index\": {next_index}, \"last_username\": \"{username}\", \"last_audit_date\": \"{iso_timestamp}\", \"last_result\": \"PASS or FAIL\"}",
  "username": "{username}"
}
```

Where `next_index` is the current index + 1 (or 0 if wrapping to start of list).

The safe-output job will create a PR to update `.github/audit-state.json`, ensuring state persists across runs.

## Example Markdown Email Structure

```markdown
Hey @clubanderson! 👋

Here's your KubeStellar impact snapshot for the last 60 days.

---

## 📊 Quick Stats

✅ **Help-Wanted Issues:** 5 created (required: ≥2)  
❌ **PR Reviews:** 6 unique PRs (required: ≥8) — _Let's boost this!_  
✅ **PRs Merged:** 4 merged (required: ≥3)

**Overall:** 2 of 3 criteria met

---

## 🎯 Your Impact Areas

Based on your recent contributions, you're passionate about:

- 📝 **Documentation** (60% of your PRs touch `/docs/`)
- 🧪 **Testing** (noticed several `*_test.go` files)
- Most active in: **kubestellar/docs**, **kubestellar/kubestellar**

---

## 🌟 Where You Can Help Next

### 🏷️ Help-Wanted Issues Perfect For You

1. **[Improve Getting Started Guide](https://github.com/kubestellar/docs/issues/123)**  
   `kubestellar/docs` • Labels: documentation, good-first-issue

2. **[Add Integration Test Coverage](https://github.com/kubestellar/kubestellar/issues/456)**  
   `kubestellar/kubestellar` • Labels: testing, help-wanted

3. **[Document API Reference](https://github.com/kubestellar/docs/issues/789)**  
   `kubestellar/docs` • Labels: documentation, help-wanted

### 👀 PRs That Need Your Review

1. **[Update deployment docs for v0.25](https://github.com/kubestellar/docs/pull/234)**  
   `kubestellar/docs` • Docs changes, no reviews yet

2. **[Add E2E test for multi-cluster](https://github.com/kubestellar/kubestellar/pull/567)**  
   `kubestellar/kubestellar` • Testing PR, needs expert eyes

3. **[Fix typos in contributor guide](https://github.com/kubestellar/docs/pull/890)**  
   `kubestellar/docs` • Quick review needed

### 🎯 Repos Looking for Your Skills

1. **kubestellar/docs** — Your top repo! Several open doc issues need attention.
2. **kubestellar/kubestellar** — Core repo could use more test coverage (your strength!).
3. **kubestellar/kubeflex** — Growing repo, needs documentation help.

---

## 💪 Keep Up the Great Work!

You're making a real difference in KubeStellar! To hit all 3 criteria next time, focus on reviewing a couple more PRs in areas you love. Your expertise in docs and testing is invaluable. 🙌

---

_Automated by GitHub Agentic Workflows • 2025-12-05 20:18 UTC_
```

## Important Notes

- **Pure Markdown** - no HTML conversion, sent as plain text via Postmark
- **Encouraging tone** - focus on opportunities, not failures
- **Personalized recommendations** - based on detected interests from PR history
- **Actionable links** - direct links to specific issues/PRs they can tackle
- **Context-aware** - adapt tone based on PASS/FAIL (celebrate or motivate)
