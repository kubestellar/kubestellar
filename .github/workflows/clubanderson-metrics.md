---
description: |
  Simple metrics tracker for clubanderson. Checks 3 criteria over last 60 days:
  - Help-wanted issues created (≥2)
  - Unique PRs commented on (≥8)
  - Merged PRs authored (≥3)
  Sends pass/fail email to andy@clubanderson.com

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
          mkdir -p /tmp/metrics-data
          
          # Search 1: Help-wanted issues created
          gh search issues \
            --owner kubestellar \
            --label "help wanted" \
            --author clubanderson \
            --created ">=${{ steps.dates.outputs.date_60 }}" \
            --json number,title,url,createdAt,labels \
            --jq '{total_count: length, items: .}' \
            > /tmp/metrics-data/help-wanted-created.json
          
          # Search 2: PRs commented on
          gh search prs \
            --owner kubestellar \
            --commenter clubanderson \
            --updated ">=${{ steps.dates.outputs.date_60 }}" \
            --json number,title,url,state \
            --jq '{total_count: length, items: .}' \
            > /tmp/metrics-data/prs-commented.json
          
          # Search 3: Merged PRs authored
          gh search prs \
            --owner kubestellar \
            --author clubanderson \
            --merged ">=${{ steps.dates.outputs.date_60 }}" \
            --json number,title,url,closedAt,labels \
            --jq '{total_count: length, items: .}' \
            > /tmp/metrics-data/prs-merged.json
          
          # Search 4: All open issues in active repos
          gh search issues \
            --owner kubestellar \
            --repo docs \
            --repo ui \
            --repo ui-plugins \
            --state open \
            --json number,title,url,repository,labels,createdAt \
            --jq '{total_count: length, items: .}' \
            > /tmp/metrics-data/open-issues.json
          
          # Search 5: All open PRs in active repos
          gh search prs \
            --owner kubestellar \
            --repo docs \
            --repo ui \
            --repo ui-plugins \
            --state open \
            --json number,title,url,repository,labels,createdAt \
            --jq '{total_count: length, items: .}' \
            > /tmp/metrics-data/open-prs.json
          
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

safe-outputs:
  jobs:
    send-email:
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

# Clubanderson Metrics Tracker

Your task is to **generate ONE metrics email for clubanderson** using pre-downloaded GitHub data.

## Pre-Downloaded Data Files

All GitHub search data has been downloaded for you. The following JSON files are available in `/tmp/metrics-data/`:

1. **help-wanted-created.json** - Help-wanted issues created by clubanderson (last 60 days)
2. **prs-commented.json** - PRs clubanderson commented on (last 60 days)  
3. **prs-merged.json** - Merged PRs authored by clubanderson (last 60 days)
4. **open-issues.json** - All open issues in docs/ui/ui-plugins repos
5. **open-prs.json** - All open PRs in docs/ui/ui-plugins repos

## Your Task

Read these 5 JSON files and analyze the data:

**Calculate metrics:**
- Count items in `help-wanted-created.json` → must be ≥2
- Count unique PR numbers from `prs-commented.json` → must be ≥8
- Count items in `prs-merged.json` → must be ≥3

**Detect expertise:**
From `prs-merged.json`, look at PR titles/labels to identify if clubanderson works on CI/CD, docs, or UI.

**Generate recommendations:**
- From `open-issues.json`: Find 3 issues with "help wanted" label
- From `open-prs.json`: Find 3 PRs that need review  
- From `open-issues.json`: Find 3 recent issues (created in last 30 days) that need PRs

**Generate email:**
Create a plain-text email with:
- Pass/fail for each metric
- Top 3 recommendations in each category
- Simple formatting (no markdown headings)

**Output:**
Create a safe-output JSON with type "send_email" containing the email subject and body.

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
