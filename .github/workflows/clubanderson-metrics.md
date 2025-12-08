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

1. Calculate the date 60 days ago from today (YYYY-MM-DD format)
2. Run exactly 3 GitHub searches (org-wide across kubestellar):
   - Help-wanted issues created: `org:kubestellar is:issue label:"help wanted" author:clubanderson created:>={date_60_days_ago}`
   - PRs commented on (merged): `org:kubestellar is:pr is:merged commenter:clubanderson updated:>={date_60_days_ago}`
   - PRs commented on (open): `org:kubestellar is:pr is:open commenter:clubanderson updated:>={date_60_days_ago}`
   - Merged PRs: `org:kubestellar is:pr is:merged author:clubanderson merged:>={date_60_days_ago}`
3. Count the results and check against criteria:
   - Help-wanted issues ≥ 2? ✅ / ❌
   - Unique PRs commented ≥ 8? (deduplicate PR numbers from merged + open searches) ✅ / ❌
   - Merged PRs ≥ 3? ✅ / ❌
4. Generate a simple plain-text email with results
5. Output the safe-output JSON

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
