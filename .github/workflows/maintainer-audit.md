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

**⚠️ CRITICAL EXECUTION RULES:**

1. **ONE EXECUTION ONLY**: Execute the workflow ONCE. After you complete Step 10, you MUST stop completely.
2. **NO LOOPS**: Do NOT restart from Step 1. Do NOT re-analyze the same user.
3. **LINEAR PROGRESSION**: Steps 1 → 2 → 3 → 4 → 5 → 6 → 7 → 8 → 9 → 10 → STOP.
4. **If you find yourself at Step 1 again, STOP IMMEDIATELY** - you've already completed the audit.

## Your Mission

Audit **clubanderson** (locked for testing). Execute steps 1-11 exactly once. After Step 10 (Output Safe-Output Entry), output the safe-output JSON and STOP. Do not continue. Do not loop back.

## Target Maintainer (TEST MODE - LOCKED)

**clubanderson** (`andy@clubanderson.com`) - Index 0

**For future reference (not used during testing):** The full maintainer list is: clubanderson, mikespreitzer, dumb0002, waltforme, pdettori, francostellari, kproche, nupurshivani, onkar717, kunal-511, mavrick-1, gaurab-khanal, naman9271, btwshivam, rxinui, vedansh-5, sagar2366, oksaumya, rupam-it.

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

**⚠️ CRITICAL: Execution Strategy**

This workflow audits **ONE maintainer per run**. Follow this linear process **without repeating steps**:

1. **Load state** → Get next maintainer index
2. **Select maintainer** → Store username + email
3. **Calculate dates** → Compute 60-day and 180-day cutoffs ONCE
4. **Analyze interests** → ONE search for past PRs, extract patterns
5. **Gather metrics** → THREE searches total (help-wanted, PRs commented, PRs merged)
6. **Find opportunities** → TWO searches (help-wanted issues, PRs needing review)
7. **Suggest help-wanted areas** → Analyze repo health + expertise for creating new issues
8. **Evaluate** → Compare metrics to requirements
9. **Generate email** → Create personalized report
10. **Output results** → Write safe-output entries
11. **Update state** → Save next index

**Do NOT repeat searches or loop over the same user's data.** Each search should happen **exactly once**.

---

### Step 1: Load State from Repository (TESTING MODE - LOCKED)

**🔒 TEST MODE: Always use index 0 (clubanderson)**

For now, ignore the state file and **always audit clubanderson**:
- Set index to `0`
- Do NOT load or increment from `.github/audit-state.json`
- Skip round-robin logic entirely

### Step 2: Select Maintainer (LOCKED TO clubanderson)

**🔒 TEST MODE: Always audit clubanderson**

- Username: `clubanderson`
- Email: `andy@clubanderson.com`
- Index: `0`

Store these values and proceed to Step 3.

### Step 3: Calculate Date Range

Calculate the date 60 days ago from today in YYYY-MM-DD format.

### Step 4: Analyze Maintainer's Interests

Before gathering metrics, understand what the maintainer likes to work on:

**A. Analyze Past PRs (Last 6 months) - ONE SEARCH ONLY**

- Search ONCE: `org:kubestellar is:pr is:merged author:{username} merged:>={date_180_days_ago}`
- Examine up to 10-15 recent PRs from the results:
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
- **Store these patterns and proceed immediately to Step 5**

### Step 5: Gather Current Metrics

**IMPORTANT:** Perform each search **ONCE** and store the results. Do not repeat searches.

Use the GitHub MCP tools to search and count:

**A. Help-Wanted Issues Created**

- Search ONCE: `org:kubestellar is:issue label:"help wanted" author:{username} created:>={date_60_days_ago}`
- Count total results and store the count

**B. Unique PRs Commented On**

- Search merged PRs ONCE: `org:kubestellar is:pr is:merged commenter:{username} updated:>={date_60_days_ago}`
- Search open PRs ONCE: `org:kubestellar is:pr is:open commenter:{username} updated:>={date_60_days_ago}`
- Extract PR numbers from both result sets
- Count unique PR numbers (deduplicate) and store the count
- **Do NOT repeat these searches after getting the results**

**C. Merged PRs Authored**

- Search ONCE: `org:kubestellar is:pr is:merged author:{username} merged:>={date_60_days_ago}`
- Count total results and store the count

**Once you have all three metrics (A, B, C), proceed immediately to Step 6.**

### Step 6: Find Personalized Opportunities

**IMPORTANT:** Search for opportunities **ONCE** per category. Use the results to pick the top 3.

Based on their interest patterns, find **Top 3** opportunities in each category:

**A. Help-Wanted Issues in Their Favorite Areas**

- Search ONCE across all KubeStellar repos: `org:kubestellar is:issue is:open label:"help wanted"`
- From the results, filter/rank by:
  - Repos they've contributed to before (higher priority)
  - File paths/labels matching their detected interests
  - Recent activity (updated within last 14 days)
- Select top 3 issues with direct links
- **Do NOT search again after getting results**

**B. PRs Needing Review in Their Expertise Areas**

- Search ONCE: `org:kubestellar is:pr is:open review:required`
- From the results, filter/rank by:
  - Repos they're active in
  - File paths matching their interest patterns (docs, backend, frontend, etc.)
  - PRs without recent review activity
- Select top 3 PRs with direct links
- **Do NOT search again after getting results**

**C. Repos That Could Use Their Skills**

- Based on their detected focus areas (docs, backend, UI, testing, DevOps):
  - Match to KubeStellar repos that align (e.g., docs contributor → kubestellar/docs)
  - Note recent issues/PRs in those repos needing attention
- Suggest top 3 repos with reasons why

**Once you have gathered opportunities from A, B, and C above, proceed immediately to Step 7.**

### Step 7: Suggest Areas for Creating Help-Wanted Issues

**IMPORTANT:** Analyze repo health and maintainer expertise to suggest where THEY should create new help-wanted issues.

**A. Repo Health Analysis**

For each KubeStellar repo the maintainer is active in:
- Check recent issue activity (last 30 days): `org:kubestellar is:issue repo:{repo_name} created:>={date_30_days_ago}`
- Check recent PR activity (last 30 days): `org:kubestellar is:pr repo:{repo_name} created:>={date_30_days_ago}`
- Identify "cold spots" - repos with low issue/PR activity that may need help-wanted issues to attract contributors

**B. Maintainer Expertise Matching**

Based on Step 4 (their detected interests):
- Match their expertise areas (docs, testing, UI, backend, DevOps) to repos where they're active
- Identify specific areas where they could create help-wanted issues:
  - **Documentation gaps** - Missing or outdated docs they could outline as help-wanted
  - **Testing coverage** - Areas lacking tests where they could define test scenarios
  - **UI improvements** - UX enhancements they've identified
  - **Technical debt** - Refactoring opportunities in their domain

**C. Generate Top 3 Suggestions**

Create **specific, actionable** suggestions like:
- "Create help-wanted issues for testing coverage in kubestellar/kubestellar `/pkg/` modules"
- "Document API endpoints in kubestellar/docs - outline structure as help-wanted for new contributors"
- "Create UI accessibility issues in kubestellar/ui based on your recent work"

Each suggestion should explain:
- **What** to create (specific issue type/area)
- **Where** (repo + path/component)
- **Why** (repo health gap OR leveraging their expertise)

**Once you have these suggestions, proceed to Step 8.**

### Step 8: Evaluate Criteria

Compare actual counts against requirements:

- Help-wanted issues: actual >= 2 ? ✅ PASS : ❌ FAIL
- Unique PRs commented: actual >= 8 ? ✅ PASS : ❌ FAIL
- Merged PRs: actual >= 3 ? ✅ PASS : ❌ FAIL

Overall: PASS if all three criteria pass, otherwise FAIL

### Step 9: Generate Personalized Markdown Email

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
5. **"Consider Creating Help-Wanted Issues"** - NEW SECTION:
   - **✨ Suggestions from Step 7:** Top 3 areas where they should create help-wanted issues
   - Include **what**, **where**, and **why** for each suggestion
6. **Encouraging closing:**
   - If PASS: Celebrate their contributions and suggest maintaining momentum
   - If FAIL: Focus on opportunities, frame as "here's how to get back on track"
7. **Footer:** Automation note with timestamp

**Tone:** Supportive and constructive, not punitive

**Formatting Rules:**
- DO NOT use markdown headings (##, ###) - they look weird in plain text emails
- Use plain text section labels with emojis and horizontal rules (---) for separation
- Use plain text, NOT inline code backticks for usernames (write `@username` not `` `@username` ``)
- Use **bold** for emphasis, _italic_ for secondary emphasis
- Use standard markdown lists with `-` or numbered `1.`
- Avoid wrapping normal text in backticks unless it's actual code/commands

### Step 10: Output Safe-Output Entry

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

**AFTER creating this safe-output entry, your work is COMPLETE. Stop execution. Do NOT restart from Step 1.**

### Step 11: Output State Update (TESTING MODE - DISABLED)

**🔒 TEST MODE: Do NOT update state**

Skip the state update output entirely. Do NOT create a safe-output entry for `update_audit_state`.

Since we're locked to clubanderson for testing, there's no need to track progress.

## Example Markdown Email Structure

**IMPORTANT:** Follow this example format exactly. Notice NO markdown headings (##), just plain text labels with emojis and horizontal rules.

```markdown
Hey @clubanderson! 👋

Here's your KubeStellar impact snapshot for the last 60 days.

---

📊 **Quick Stats**

✅ **Help-Wanted Issues:** 5 created (required: ≥2)  
❌ **PR Reviews:** 6 unique PRs (required: ≥8) — _Let's boost this!_  
✅ **PRs Merged:** 4 merged (required: ≥3)

**Overall:** 2 of 3 criteria met

---

🎯 **Your Impact Areas**

Based on your recent contributions, you're passionate about:

- 📝 **Documentation** (60% of your PRs touch `/docs/`)
- 🧪 **Testing** (noticed several `*_test.go` files)
- Most active in: **kubestellar/docs**, **kubestellar/kubestellar**

---

🌟 **Where You Can Help Next**

🏷️ **Help-Wanted Issues Perfect For You**

1. **[Improve Getting Started Guide](https://github.com/kubestellar/docs/issues/123)**  
   kubestellar/docs • Labels: documentation, good-first-issue

2. **[Add Integration Test Coverage](https://github.com/kubestellar/kubestellar/issues/456)**  
   kubestellar/kubestellar • Labels: testing, help-wanted

3. **[Document API Reference](https://github.com/kubestellar/docs/issues/789)**  
   kubestellar/docs • Labels: documentation, help-wanted

👀 **PRs That Need Your Review**

1. **[Update deployment docs for v0.25](https://github.com/kubestellar/docs/pull/234)**  
   kubestellar/docs • Docs changes, no reviews yet

2. **[Add E2E test for multi-cluster](https://github.com/kubestellar/kubestellar/pull/567)**  
   kubestellar/kubestellar • Testing PR, needs expert eyes

3. **[Fix typos in contributor guide](https://github.com/kubestellar/docs/pull/890)**  
   kubestellar/docs • Quick review needed

🎯 **Repos Looking for Your Skills**

1. **kubestellar/docs** — Your top repo! Several open doc issues need attention.
2. **kubestellar/kubestellar** — Core repo could use more test coverage (your strength!).
3. **kubestellar/kubeflex** — Growing repo, needs documentation help.

---

✨ **Consider Creating Help-Wanted Issues**

Here are areas where YOU could create help-wanted issues to grow our contributor base:

1. **Testing coverage for `/pkg/` modules in kubestellar/kubestellar**  
   _Why:_ Low recent test activity detected. Your testing expertise could outline specific test scenarios as help-wanted issues for new contributors.

2. **Document API endpoints in kubestellar/docs**  
   _Why:_ You're the docs expert! Create help-wanted issues with outlines for missing API documentation.

3. **Accessibility improvements in kubestellar/ui**  
   _Why:_ Based on your recent UI work, you could identify accessibility gaps and create structured help-wanted issues.

---

💪 **Keep Up the Great Work!**

You're making a real difference in KubeStellar! To hit all 3 criteria next time, focus on reviewing a couple more PRs in areas you love. Your expertise in docs and testing is invaluable. 🙌

---

_Automated by GitHub Agentic Workflows • 2025-12-05 20:18 UTC_
```

---

**END OF WORKFLOW INSTRUCTIONS. Your audit is complete after Step 10. Stop here.**
