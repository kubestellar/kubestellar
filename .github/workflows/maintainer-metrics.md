---
description: |
  Metrics tracker for KubeStellar maintainers. Checks 3 criteria over last 60 days:
  - Help-wanted issues created (≥2)
  - Unique PRs commented on (≥8)
  - Merged PRs authored (≥3)
  Run manually for individual maintainers via dispatch dropdown

on:
  workflow_dispatch:
    inputs:
      maintainer:
        description: "Select maintainer to audit"
        required: true
        type: choice
        options:
          - btwshivam
          - clubanderson
          - dumb0002
          - francostellari
          - gaurab-khanal
          - kproche
          - kunal-511
          - mavrick-1
          - mikespreitzer
          - naman9271
          - nupurshivani
          - oksaumya
          - onkar717
          - pdettori
          - rupam-it
          - rxinui
          - sagar2366
          - vedansh-5
          - waltforme

run-name: "Maintainer Metrics Tracker - ${{ inputs.maintainer }}"

permissions: read-all

jobs:
  fetch-data:
    name: Fetch GitHub Data
    runs-on: ubuntu-latest
    outputs:
      data-ready: ${{ steps.fetch.outputs.ready }}
    steps:
      - name: Calculate dates
        id: dates
        run: |
          echo "date_60=$(date -d '60 days ago' '+%Y-%m-%d')" >> $GITHUB_OUTPUT
          echo "date_30=$(date -d '30 days ago' '+%Y-%m-%d')" >> $GITHUB_OUTPUT

      - name: Fetch all GitHub data
        id: fetch
        env:
          GH_TOKEN: ${{ secrets.GH_AUDIT_TOKEN }}
        run: |
          # Fetch data for selected maintainer
          username="${{ github.event.inputs.maintainer }}"
          echo "Fetching data for $username..."
          mkdir -p /tmp/metrics-data/$username

          # Search 1: Help-wanted issues created
          gh search issues \
            --owner kubestellar \
            --label "help wanted" \
            --author $username \
            --created ">=${{ steps.dates.outputs.date_60 }}" \
            --limit 100 \
            --json number,title,url,createdAt,labels \
            --jq '{total_count: length, items: .}' \
            > /tmp/metrics-data/$username/help-wanted-created.json

          # Search 2: PRs commented/reviewed on (merged)
          gh search prs \
            --owner kubestellar \
            --commenter $username \
            --merged \
            --updated ">=${{ steps.dates.outputs.date_60 }}" \
            --limit 1000 \
            --json number,title,url,state \
            --jq '{total_count: length, items: .}' \
            > /tmp/metrics-data/$username/prs-commented-merged.json

          # Search 3: PRs commented/reviewed on (open)
          gh search prs \
            --owner kubestellar \
            --commenter $username \
            --state open \
            --updated ">=${{ steps.dates.outputs.date_60 }}" \
            --limit 1000 \
            --json number,title,url,state \
            --jq '{total_count: length, items: .}' \
            > /tmp/metrics-data/$username/prs-commented-open.json

          # Search 4: Merged PRs authored
          echo "Searching for PRs merged by $username since ${{ steps.dates.outputs.date_60 }}"
          gh search prs \
            --owner kubestellar \
            --author $username \
            --merged \
            --merged-at ">=${{ steps.dates.outputs.date_60 }}" \
            --limit 1000 \
            --json number,title,url,closedAt,labels \
            --jq '{total_count: length, items: .}' \
            > /tmp/metrics-data/$username/prs-merged.json

          # Shared data for all maintainers (put in shared location)
          mkdir -p /tmp/metrics-data/shared

          # Search: All open issues in active repos (for recommendations)
          gh search issues \
            --owner kubestellar \
            --repo docs \
            --repo ui \
            --repo ui-plugins \
            --state open \
            --limit 1000 \
            --json number,title,url,repository,labels,createdAt \
            --jq '{total_count: length, items: .}' \
            > /tmp/metrics-data/shared/open-issues.json

          # Search: All open PRs in active repos (for recommendations)
          gh search prs \
            --owner kubestellar \
            --repo docs \
            --repo ui \
            --repo ui-plugins \
            --state open \
            --limit 1000 \
            --json number,title,url,repository,labels,createdAt \
            --jq '{total_count: length, items: .}' \
            > /tmp/metrics-data/shared/open-prs.json

          # Copy shared files to selected maintainer's folder
          username="${{ github.event.inputs.maintainer }}"
          cp /tmp/metrics-data/shared/open-issues.json /tmp/metrics-data/$username/
          cp /tmp/metrics-data/shared/open-prs.json /tmp/metrics-data/$username/

          echo "ready=true" >> $GITHUB_OUTPUT

      - name: Upload data as artifact
        uses: actions/upload-artifact@v4
        with:
          name: metrics-data
          path: /tmp/metrics-data/
          retention-days: 1

steps:
  - name: Download metrics data
    uses: actions/download-artifact@v4
    with:
      name: metrics-data
      path: /tmp/metrics-data/

tools:
  bash:
    - "cat *"
    - "jq *"
    - "wc *"
    - "grep *"
    - "head *"
    - "tail *"
    - "mkdir *"
    - "ls *"
    - "echo *"

safe-outputs:
  jobs:
    send_email:
      description: "Send metrics email via Postmark"
      runs-on: ubuntu-latest
      output: "Email sent!"
      inputs:
        subject:
          description: "Email subject"
          required: true
          type: string
        body:
          description: "Plain text email body"
          required: true
          type: string
      permissions:
        contents: read
      steps:
        - name: Send via Postmark
          uses: actions/github-script@v7
          env:
            POSTMARK_API_TOKEN: "${{ secrets.POSTMARK_API_TOKEN }}"
            POSTMARK_FROM_EMAIL: "${{ secrets.POSTMARK_FROM_EMAIL }}"
          with:
            script: |
              const postmarkToken = process.env.POSTMARK_API_TOKEN;
              const fromEmail = process.env.POSTMARK_FROM_EMAIL;

              // Maintainer email mapping
              const maintainerEmails = {
                'btwshivam': 'shivam200446@gmail.com',
                'clubanderson': 'andy@clubanderson.com',
                'dumb0002': 'Braulio.Dumba@ibm.com',
                'francostellari': 'stellari@us.ibm.com',
                'gaurab-khanal': 'khanalgaurab98@gmail.com',
                'kproche': 'kproche@us.ibm.com',
                'kunal-511': 'yoyokvunal@gmail.com',
                'mavrick-1': 'mavrickrishi@gmail.com',
                'mikespreitzer': 'mspreitz@us.ibm.com',
                'naman9271': 'namanjain9271@gmail.com',
                'nupurshivani': 'nupurjha.me@gmail.com',
                'oksaumya': 'saumyakr2006@gmail.com',
                'onkar717': 'onkarwork2234@gmail.com',
                'pdettori': 'dettori@us.ibm.com',
                'rupam-it': 'Mannarupam3@gmail.com',
                'rxinui': 'rainui.ly@gmail.com',
                'sagar2366': 'sagarutekar2366@gmail.com',
                'vedansh-5': 'vedanshsaini7719@gmail.com',
                'waltforme': 'jun.duan@ibm.com'
              };

              const maintainer = '${{ github.event.inputs.maintainer }}';
              const toEmail = maintainerEmails[maintainer];

              if (!toEmail) {
                core.setFailed(`No email found for maintainer: ${maintainer}`);
                return;
              }

              const fs = require('fs');
              const outputFile = process.env.GH_AW_AGENT_OUTPUT;

              if (!postmarkToken || !fromEmail) {
                core.setFailed('Missing Postmark secrets');
                return;
              }

              if (!outputFile) {
                core.info('No agent output file found');
                return;
              }

              const fileContent = fs.readFileSync(outputFile, 'utf8');
              const agentOutput = JSON.parse(fileContent);

              const emailItems = agentOutput.items?.filter(item => item.type === 'send_email') || [];

              if (emailItems.length === 0) {
                core.info('No email items to send');
                return;
              }

              for (const item of emailItems) {
                const { subject, body } = item;
                
                core.info(`Sending email: ${subject}`);
                
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
                    Cc: 'andy@clubanderson.com',
                    Subject: subject,
                    TextBody: body,
                    MessageStream: 'outbound'
                  })
                });
                
                if (!response.ok) {
                  const errorText = await response.text();
                  core.setFailed(`Postmark error: ${response.status} - ${errorText}`);
                  return;
                }
                
                const result = await response.json();
                core.info(`✅ Email sent! MessageID: ${result.MessageId}`);
              }
---

# Maintainer Metrics Tracker

Your task is to **generate ONE metrics email** for the selected maintainer using pre-downloaded GitHub data.

**Selected maintainer:** ${{ github.event.inputs.maintainer }}

**Email mapping:**

- btwshivam → shivam200446@gmail.com
- clubanderson → andy@clubanderson.com
- dumb0002 → Braulio.Dumba@ibm.com
- francostellari → stellari@us.ibm.com
- gaurab-khanal → khanalgaurab98@gmail.com
- kproche → kproche@us.ibm.com
- kunal-511 → yoyokvunal@gmail.com
- mavrick-1 → mavrickrishi@gmail.com
- mikespreitzer → mspreitz@us.ibm.com
- naman9271 → namanjain9271@gmail.com
- nupurshivani → nupurjha.me@gmail.com
- oksaumya → saumyakr2006@gmail.com
- onkar717 → onkarwork2234@gmail.com
- pdettori → dettori@us.ibm.com
- rupam-it → Mannarupam3@gmail.com
- rxinui → rainui.ly@gmail.com
- sagar2366 → sagarutekar2366@gmail.com
- vedansh-5 → vedanshsaini7719@gmail.com
- waltforme → jun.duan@ibm.com

The data files are in `/tmp/metrics-data/${{ github.event.inputs.maintainer }}/`:

## Pre-Downloaded Data Files

For each maintainer, these 6 JSON files are available in `/tmp/metrics-data/{username}/`:

1. **help-wanted-created.json** - Help-wanted issues created by the user (last 60 days)
2. **prs-commented-merged.json** - Merged PRs the user commented on (last 60 days)
3. **prs-commented-open.json** - Open PRs the user commented on (last 60 days)
4. **prs-merged.json** - Merged PRs authored by the user (last 60 days)
5. **open-issues.json** - All open issues in docs/ui/ui-plugins repos
6. **open-prs.json** - All open PRs in docs/ui/ui-plugins repos

## Your Task

For the selected maintainer (${{ github.event.inputs.maintainer }}):

**Calculate metrics (READ CAREFULLY - DO NOT MISCOUNT):**

First, display the raw data so we can verify:

```bash
cat /tmp/metrics-data/$username/help-wanted-created.json
cat /tmp/metrics-data/$username/prs-merged.json
```

**CRITICAL - Read the total_count field EXACTLY:**

Each JSON file has this structure:

```json
{"total_count": 4, "items": [...]}
```

**Step 1: Display the raw data for verification**

```bash
echo "=== HELP-WANTED DATA ==="
cat /tmp/metrics-data/${{ github.event.inputs.maintainer }}/help-wanted-created.json

echo "=== MERGED PRS DATA ==="
cat /tmp/metrics-data/${{ github.event.inputs.maintainer }}/prs-merged.json

echo "=== COMMENTED MERGED PRS DATA ==="
cat /tmp/metrics-data/${{ github.event.inputs.maintainer }}/prs-commented-merged.json

echo "=== COMMENTED OPEN PRS DATA ==="
cat /tmp/metrics-data/${{ github.event.inputs.maintainer }}/prs-commented-open.json
```

**Step 2: Extract metrics EXACTLY from total_count fields**

⚠️ **CRITICAL INSTRUCTION - DO NOT SKIP THIS:**

**RULE #1:** ONLY read the exact number after `"total_count":` in the JSON. DO NOT calculate, filter, count items, or modify in any way.

**RULE #2:** If you see `{"total_count": 11,` then the metric is **11**. Not 3. Not 12. Not the length of items array. **Exactly 11**.

**RULE #3:** Display the metrics you extracted before generating the email so we can verify they match the JSON.

Extract these three numbers using ONLY the total_count field:

1. **Help-wanted count** = `total_count` from help-wanted-created.json
   - Look for the line: `"total_count": X`
   - Use X directly (no math, no processing)
   - Example: `{"total_count": 11,` → metric is **11**

2. **Merged PRs count** = `total_count` from prs-merged.json
   - Look for the line: `"total_count": X`
   - Use X directly (no math, no processing)
   - Example: `{"total_count": 26,` → metric is **26**

3. **PR reviews count** = `total_count` from prs-commented-merged.json + `total_count` from prs-commented-open.json
   - Find `"total_count": X` in prs-commented-merged.json
   - Find `"total_count": Y` in prs-commented-open.json
   - PR reviews = X + Y (simple addition)
   - Example: if merged has `"total_count": 59` and open has `"total_count": 9` → 59 + 9 = **68**

**Before generating email, display:**

```
Extracted metrics:
- Help-wanted: [number from help-wanted-created.json total_count]
- Merged PRs: [number from prs-merged.json total_count]
- PR Reviews: [merged total_count] + [open total_count] = [sum]
```

**DO NOT:**

- ❌ Count items manually
- ❌ Try to deduplicate PR numbers
- ❌ Filter or analyze the items array
- ❌ Use jq, python, or any processing tools
- ❌ Modify the total_count numbers in any way

**DO:**

- ✅ Find the number after `"total_count":` in the displayed JSON
- ✅ Use that exact number in the email
- ✅ Add the two PR review counts together (merged + open)

**Detect expertise:**
From `prs-merged.json`, look at PR titles/labels to identify their work areas (CI/CD, docs, UI, testing, frontend, backend, etc).

**Generate Help-Wanted Suggestions (WHERE TO CREATE NEW ISSUES):**

⚠️ **CRITICAL:** This section is NOT about assigning the maintainer to existing help-wanted issues. It's about suggesting WHERE they should CREATE NEW help-wanted issues to grow the community.

Generate 3 suggestions using this framework:

1. **[Expertise-Based]** - Where their domain knowledge can guide new contributors
   - If they're a docs expert: "Create help-wanted issues for expanding API documentation in kubestellar/docs - Outline specific sections that need contributor help"
   - If they're a CI/CD expert: "Create help-wanted issues for workflow standardization across kubestellar/\* repos - Identify automation patterns that could be templates for contributors"
   - If they're a UI expert: "Create help-wanted issues for component refactoring in kubestellar/ui - Break down UI technical debt into approachable tasks"

2. **[Project Growth]** - Where the project needs maturity/advancement
   - Look at repos they're active in from `prs-merged.json`
   - Suggest: "Create help-wanted issues to advance [specific area needing growth] in [repo] - Your experience with [their work] can help identify gaps"
   - Examples:
     - Testing coverage: "Create help-wanted issues for E2E test scenarios in kubestellar/kubestellar - Define test cases contributors can implement"
     - Documentation gaps: "Create help-wanted issues for troubleshooting guides in kubestellar/docs - Outline common issues that need documentation"
     - CI/CD maturity: "Create help-wanted issues for GitHub Actions improvements across repos - Identify workflow patterns to standardize"

3. **[Community Building]** - Breaking down complex work into contributor-friendly chunks
   - Suggest: "Create help-wanted issues that break down [complex feature] into smaller tasks in [repo] - Make [advanced work] accessible to new contributors"
   - Examples:
     - "Create help-wanted issues for UI accessibility improvements in kubestellar/ui - Break A11y audit findings into actionable tasks"
     - "Create help-wanted issues for Helm chart enhancements in kubestellar/kubestellar - Decompose chart improvements into discrete PRs"
     - "Create help-wanted issues for integration test coverage in kubestellar/kubeflex - Define test scenarios for contributors to implement"

**Format:** Each suggestion should be actionable and specific:

- What type of issues to create
- In which repo
- Why it helps the project/community
- How it leverages their expertise

**Generate other recommendations (WITH URLs):**

- From `open-prs.json`: Find 3 PRs that need review (match their expertise)
- From `open-issues.json`: Find 3 recent issues (created in last 30 days) that need PRs (match their expertise)

**IMPORTANT - All recommendations MUST include clickable URLs:**

- Format: "Title (repo #number) - https://github.com/org/repo/issues/number"
- Example: "Fix UI bug (kubestellar/ui #2275) - https://github.com/kubestellar/ui/issues/2275"
- The URL field is in the JSON as `url` - use it directly

**Generate email:**
Create a plain-text email with:

- Pass/fail for each metric
- Top 3 recommendations in each category (with URLs!)
- Simple formatting (no markdown headings)
- Include the 60-day date range in the opening line

**Calculate date range:**

```bash
# Get today's date
date '+%b %-d, %Y'

# Get date 60 days ago
date -d '60 days ago' '+%b %-d, %Y'
```

**Output:**
Use the **send_email** MCP tool to send the metrics email. Format the subject like this:

- If PASS: `🚀 KubeStellar Metrics - {username} - {date} - ✅ PASS`
- If FAIL: `🚀 KubeStellar Metrics - {username} - {date} - ❌ FAIL`

Where {date} is the current date in format "Dec 9, 2025" (use the `date` command: `date '+%b %-d, %Y'`)

Example:

```
send_email(subject="🚀 KubeStellar Metrics - clubanderson - Dec 9, 2025 - ✅ PASS", body="Hey clubanderson,...")
```

**Note:** Do NOT use tick marks around the username (no @clubanderson, just clubanderson)

Do NOT print JSON. Do NOT use echo. Use the MCP tool directly.

**DO NOT use any GitHub search tools. Only read the pre-downloaded JSON files.**

## Email Format

Keep it simple and clear:

```
Subject: 🚀 KubeStellar Metrics - clubanderson - Dec 9, 2025 - ✅ PASS
(or)
Subject: 🚀 KubeStellar Metrics - kproche - Dec 9, 2025 - ❌ FAIL

Hey clubanderson,

Here are your KubeStellar metrics for the last 60 days (Oct 11, 2025 - Dec 10, 2025):

✅/❌ Help-Wanted Issues: X created (required: ≥2)
✅/❌ Merged PRs: Z merged (required: ≥3)
✅/❌ PR Reviews: Y unique PRs (required: ≥8)

Overall: PASS [3/3] or FAIL [1/3]

[If FAIL: Brief encouragement to focus on the missing criteria]

---

🏷️ Help-Wanted Issues You Could Create:
        - [Expertise-based] - Suggest where the maintainer should CREATE a new help-wanted issue in their domain (e.g., "Create help-wanted issues for improving API documentation in kubestellar/docs - Your docs expertise can help outline tasks for new contributors")
        - [Project Growth] - Suggest where the maintainer should CREATE help-wanted issues to advance project maturity (e.g., "Create help-wanted issues to standardize CI across repos in kubestellar/* - Your CI/CD knowledge can help identify automation gaps")
        - [Community Building] - Suggest where the maintainer should CREATE help-wanted issues to break down complex work (e.g., "Create help-wanted issues for UI accessibility improvements in kubestellar/ui - Break down A11y work into contributor-friendly tasks")

🔨 PR Opportunities in Your Areas:
        - Title (repo #123) - https://github.com/... - Brief reason
        - Title (repo #456) - https://github.com/... - Brief reason
        - Title (repo #789) - https://github.com/... - Brief reason

👀 PRs Needing Your Review:
        - Title (repo #123) - https://github.com/... - Brief reason
        - Title (repo #456) - https://github.com/... - Brief reason
        - Title (repo #789) - https://github.com/... - Brief reason

🌍 Growing Our User Base - Ideas for You:
        - [Social Networks] - Specific suggestion for promoting KubeStellar on LinkedIn, Slack communities, or X/Twitter based on their network (e.g., "Share KubeStellar's multi-cluster capabilities in the #kubernetes channel on CNCF Slack - your recent work on [topic] makes you a great advocate")
        - [Professional Network] - Suggestion to introduce KubeStellar at their workplace or to colleagues (e.g., "Introduce KubeStellar to your DevOps team at work - your expertise in [domain] positions you well to demonstrate how it solves [specific pain point]")
        - [Content & Advocacy] - Suggestion to create content or speak about KubeStellar (e.g., "Write a blog post about your experience with [specific feature] - your [CI/CD/docs/UI] background makes you uniquely qualified to explain the benefits")

**CRITICAL FORMATTING INSTRUCTIONS - READ CAREFULLY:**

When you generate the email body, each recommendation line MUST be formatted with EXACTLY 8 spaces of indentation.

**Correct indentation (copy this exactly):**
```

🏷️ Help-Wanted Suggestions for You: - Fix UI bug (kubestellar/ui #2275) - https://github.com/kubestellar/ui/issues/2275 - Matches your UI expertise - Add docs (kubestellar/docs #123) - https://github.com/kubestellar/docs/issues/123 - Documentation work - Improve tests (kubestellar/ui #456) - https://github.com/kubestellar/ui/issues/456 - Testing expertise

```

Count the spaces: "        -" = 8 spaces + "-" + space + text

**DO NOT:**
- ❌ Use tabs
- ❌ Use 3 spaces, 5 spaces, or any other number
- ✅ Use EXACTLY 8 spaces before each numbered item

---
Automated metrics check • {date}
```

**IMPORTANT:** The logo HTML above uses HTML entities (&lt; and &gt;). In your email output, replace these with actual angle brackets (< and >) so the HTML renders properly.

## Output

Create a safe-output JSON object:

```json
{
  "type": "send_email",
  "subject": "KubeStellar Metrics - @clubanderson - PASS",
  "body": "[generated plain text email]"
}
```

Then stop.
