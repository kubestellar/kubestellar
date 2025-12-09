---
description: |
  Metrics tracker for KubeStellar maintainers. Checks 3 criteria over last 60 days:
  - Help-wanted issues created (≥2)
  - Unique PRs commented on (≥8)
  - Merged PRs authored (≥3)
  Sends reports to clubanderson and kproche

on:
  schedule:
    - cron: "0 */6 * * *" # Every 6 hours
  workflow_dispatch:

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
          # Fetch data for BOTH maintainers
          for username in clubanderson kproche; do
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
            gh search prs \
              --owner kubestellar \
              --author $username \
              --merged \
              --merged-at ">=${{ steps.dates.outputs.date_60 }}" \
              --limit 100 \
              --json number,title,url,closedAt,labels \
              --jq '{total_count: length, items: .}' \
              > /tmp/metrics-data/$username/prs-merged.json
          done
          
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
          
          # Copy shared files to each maintainer's folder
          for username in clubanderson kproche; do
            cp /tmp/metrics-data/shared/open-issues.json /tmp/metrics-data/$username/
            cp /tmp/metrics-data/shared/open-prs.json /tmp/metrics-data/$username/
          done
          
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
              const toEmail = "andy@clubanderson.com";
              
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

Your task is to **generate TWO metrics emails** using pre-downloaded GitHub data.

Process these maintainers IN ORDER:
1. **clubanderson** → andy@clubanderson.com
2. **kproche** → kproche@us.ibm.com

For EACH maintainer, the data files are in `/tmp/metrics-data/{username}/`:

## Pre-Downloaded Data Files

For each maintainer, these 6 JSON files are available in `/tmp/metrics-data/{username}/`:

1. **help-wanted-created.json** - Help-wanted issues created by the user (last 60 days)
2. **prs-commented-merged.json** - Merged PRs the user commented on (last 60 days)  
3. **prs-commented-open.json** - Open PRs the user commented on (last 60 days)
4. **prs-merged.json** - Merged PRs authored by the user (last 60 days)
5. **open-issues.json** - All open issues in docs/ui/ui-plugins repos
6. **open-prs.json** - All open PRs in docs/ui/ui-plugins repos

## Your Task

For EACH maintainer (clubanderson, then kproche):

**Calculate metrics:**
- **Help-wanted count**: Read `help-wanted-created.json` - the file has a `total_count` field at the top
- **PR reviews count**: Read both `prs-commented-merged.json` and `prs-commented-open.json` - each has `items` array with PR numbers. Count unique numbers across both files.
- **Merged PRs count**: Read `prs-merged.json` - the file has a `total_count` field at the top

**IMPORTANT - Keep it simple:**
- ✅ DO: Read files with `cat` and manually count/parse the visible data
- ✅ DO: The JSON files are small enough to read entirely
- ✅ DO: Look for `"total_count"` field in the JSON output
- ✅ DO: Look for `"number"` fields in the items array
- ❌ DO NOT use jq, python, node, or complex bash commands
- ❌ If you see "Permission denied" - just read the file with `cat` and parse visually

**Detect expertise:**
From `prs-merged.json`, look at PR titles/labels to identify their work areas (CI/CD, docs, UI, etc).

**Generate recommendations (WITH URLs):**
- From `open-issues.json`: Find 3 issues with "help wanted" label
- From `open-prs.json`: Find 3 PRs that need review  
- From `open-issues.json`: Find 3 recent issues (created in last 30 days) that need PRs

**IMPORTANT - All recommendations MUST include clickable URLs:**
- Format: "Title (repo #number) - https://github.com/org/repo/issues/number"
- Example: "Fix UI bug (kubestellar/ui #2275) - https://github.com/kubestellar/ui/issues/2275"
- The URL field is in the JSON as `url` - use it directly

**Generate email:**
Create a plain-text email with:
- Pass/fail for each metric
- Top 3 recommendations in each category (with URLs!)
- Simple formatting (no markdown headings)

**Output:**
Use the **send_email** MCP tool to send the metrics email. Call it like this:

```
send_email(subject="KubeStellar Metrics - @{username} - PASS", body="Hey {username},...")
```

Do NOT print JSON. Do NOT use echo. Use the MCP tool directly.

**DO NOT use any GitHub search tools. Only read the pre-downloaded JSON files.**

## Email Format

Keep it simple and clear:

```
Subject: KubeStellar Metrics - @clubanderson - [PASS/FAIL]

Hey clubanderson,

Here are your KubeStellar metrics for the last 60 days:

✅ Help-Wanted Issues: X created (required: ≥2)
✅/❌ PR Reviews: Y unique PRs (required: ≥8)
✅ Merged PRs: Z merged (required: ≥3)

Overall: PASS [3/3] or FAIL [1/3]

[If FAIL: Brief encouragement to focus on the missing criteria]

---

🏷️ Help-Wanted Suggestions for You:
1. [Specific area] in [repo] - [brief reason based on low activity + your expertise]
2. [Specific area] in [repo] - [brief reason]
3. [Specific area] in [repo] - [brief reason]

👀 PRs Needing Your Review:
1. [PR title + link] - [repo] • [why it matches your skills]
2. [PR title + link] - [repo] • [why it matches your skills]
3. [PR title + link] - [repo] • [why it matches your skills]

🔨 PR Opportunities in Your Areas:
1. [Issue title + link] - [repo] • [why this needs a PR in your domain]
2. [Issue title + link] - [repo] • [why this needs a PR in your domain]
3. [Issue title + link] - [repo] • [why this needs a PR in your domain]

---
Automated metrics check • {date}
```

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
