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

tools:
  github:
    github-token: ${{ secrets.GH_AUDIT_TOKEN }}
    allowed:
      - search_issues
      - search_pull_requests
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

Your task is to **generate ONE metrics email for clubanderson** and stop. This is a single-pass workflow.

## What to do

First, calculate two dates:
- 60 days ago from today (YYYY-MM-DD)
- 30 days ago from today (YYYY-MM-DD)

Then run exactly 5 GitHub searches:
- Help-wanted issues created by clubanderson (60 days): `org:kubestellar is:issue label:"help wanted" author:clubanderson created:>={date_60_days_ago}`
- PRs clubanderson commented on (60 days): `org:kubestellar is:pr commenter:clubanderson updated:>={date_60_days_ago}`
- Merged PRs by clubanderson (60 days): `org:kubestellar is:pr is:merged author:clubanderson merged:>={date_60_days_ago}`
- All open issues in docs/ui/ui-plugins: `org:kubestellar is:issue is:open repo:kubestellar/docs repo:kubestellar/ui repo:kubestellar/ui-plugins`
- All open PRs in docs/ui/ui-plugins: `org:kubestellar is:pr is:open repo:kubestellar/docs repo:kubestellar/ui repo:kubestellar/ui-plugins`

After running these 5 searches, **stop searching and analyze the results you collected**:

**Metrics (from searches 1-3):**
- Count help-wanted issues (search 1) → must be ≥2
- Count unique PR numbers from search 2 → must be ≥8  
- Count merged PRs (search 3) → must be ≥3

**Expertise detection (from search 3 PRs):**
Look at file paths to identify if clubanderson works on CI/CD (workflows, Docker), docs (*.md), or UI (kubestellar/ui).

**Recommendations (from searches 4-5):**
- Find 3 help-wanted issues from search 4
- Find 3 PRs needing review from search 5
- Find 3 recent issues (last 30 days) from search 4 that need PRs

Generate a plain-text email with the metrics (pass/fail) and recommendations, then output the safe-output JSON and stop.

**DO NOT re-run the searches. DO NOT calculate dates again. Just use the data you already have.**

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
