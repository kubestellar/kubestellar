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

Track 3 simple metrics for clubanderson over the last 60 days and send a pass/fail email.

## Task

**Run only 5 searches total, then analyze results for everything else:**

1. Calculate dates once: `date_60_days_ago` and `date_30_days_ago`

2. **Run these 5 searches ONCE (do not repeat):**
   - **Search 1** - Help-wanted issues created: `org:kubestellar is:issue label:"help wanted" author:clubanderson created:>={date_60_days_ago}`
   - **Search 2** - PRs commented (both merged + open): `org:kubestellar is:pr commenter:clubanderson updated:>={date_60_days_ago}`
   - **Search 3** - Merged PRs authored: `org:kubestellar is:pr is:merged author:clubanderson merged:>={date_60_days_ago}`
   - **Search 4** - All open issues in active repos: `org:kubestellar is:issue is:open repo:kubestellar/docs repo:kubestellar/ui repo:kubestellar/ui-plugins`
   - **Search 5** - All open PRs in active repos: `org:kubestellar is:pr is:open repo:kubestellar/docs repo:kubestellar/ui repo:kubestellar/ui-plugins`

3. **STOP searching. Now analyze the data you have:**
   - Count Search 1 results for help-wanted issues metric
   - Count unique PR numbers from Search 2 for PR comments metric
   - Count Search 3 results for merged PRs metric

4. **Check metrics** against criteria:
   - Help-wanted issues ≥ 2? ✅ / ❌
   - Unique PRs commented ≥ 8? ✅ / ❌
   - Merged PRs ≥ 3? ✅ / ❌

5. **Analyze Search 3 results** to detect clubanderson's expertise:
   - Look at file paths in recent merged PRs to identify: CI/CD (workflows, Docker), Docs (*.md files), UI (kubestellar/ui)

6. **Reuse Search 4 & 5 results** to generate recommendations (NO additional searches):
   - **Help-Wanted Suggestions**: From Search 4, find 3 issues with `help wanted` label in docs/ui/ui-plugins
   - **PRs Needing Review**: From Search 5, find 3 PRs with `review:required` or lacking reviews, matching his expertise
   - **PR Opportunities**: From Search 4, find 3 recent issues (last 30 days) without help-wanted label that need PRs in his domains

7. Generate simple plain-text email with metrics + recommendations

8. Output the safe-output JSON and **STOP**

**IMPORTANT: After step 2, DO NOT calculate dates again. DO NOT repeat searches. Use the data you already collected.**

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
