---
description: |
  Maintainer audit report for KubeStellar. Checks 3 criteria over last 60 days for each maintainer:
  - Help-wanted issues created (≥2)
  - Unique PRs commented on (≥8)
  - Merged PRs authored (≥3)
  Sends personalized email reports with recommendations to:
  - clubanderson (andy@clubanderson.com)
  - kproche (kproche@us.ibm.com)

on:
  schedule:
    - cron: "0 */6 * * *" # Every 6 hours
  workflow_dispatch:

permissions: read-all

steps:
  - run: |
      # Will fetch data and send emails for both maintainers
      echo "Processing maintainer audits..."

tools:
  bash:
    - "date *"
    - "cat *"
    - "jq *"
    - "wc *"
    - "grep *"
    - "head *"
    - "tail *"
    - "mkdir *"

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
            POSTMARK_TOKEN: ${{ secrets.POSTMARK_SERVER_TOKEN }}
            FROM_EMAIL: ${{ secrets.MAINTAINER_AUDIT_FROM_EMAIL }}
            TO_EMAIL: ${{ inputs.body }}
          with:
            script: |
              const postmarkToken = process.env.POSTMARK_TOKEN;
              const fromEmail = process.env.FROM_EMAIL || 'noreply@kubestellar.io';
              const { subject, body } = ${{ toJSON(inputs) }};
              
              // Extract email from body (first line after "To:")
              const emailMatch = body.match(/To: (.+@.+)/);
              if (!emailMatch) {
                core.setFailed('Could not extract recipient email from body');
                return;
              }
              const toEmail = emailMatch[1].trim();
              
              // Remove the To: line from body
              const cleanBody = body.replace(/^To: .+\n/, '');
              
              if (!postmarkToken) {
                core.setFailed('POSTMARK_SERVER_TOKEN secret is not set');
                return;
              }
              
              core.info(`Sending email to: ${toEmail}`);
              core.info(`Subject: ${subject}`);
              
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
                  TextBody: cleanBody,
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
---

# Maintainer Audit Report

Process metrics for **TWO maintainers** and send emails to each:

1. **clubanderson** → andy@clubanderson.com
2. **kproche** → kproche@us.ibm.com

For EACH maintainer, follow this process:

## Step 1: Calculate dates

Run once at the start:
```bash
date -d '60 days ago' '+%Y-%m-%d'  # For metrics period
```

## Step 2: Fetch GitHub data for the maintainer

For each maintainer (clubanderson, then kproche), run these 5 searches:

```bash
# 1. Help-wanted issues created
gh search issues --owner kubestellar --label "help wanted" --author {username} --created ">={date_60}" --limit 100 --json number,title,url,createdAt,labels

# 2. Merged PRs commented on
gh search prs --owner kubestellar --commenter {username} --merged --updated ">={date_60}" --limit 1000 --json number,title,url,state

# 3. Open PRs commented on  
gh search prs --owner kubestellar --commenter {username} --state open --updated ">={date_60}" --limit 1000 --json number,title,url,state

# 4. Merged PRs authored
gh search prs --owner kubestellar --author {username} --merged --merged-at ">={date_60}" --limit 100 --json number,title,url,closedAt,labels

# 5. Open issues for recommendations
gh search issues --owner kubestellar --repo docs --repo ui --repo ui-plugins --state open --limit 1000 --json number,title,url,repository,labels,createdAt
```

## Step 3: Calculate metrics

- **Help-wanted count**: Count items from search #1 → must be ≥2
- **PR reviews count**: Count unique PR numbers from searches #2 + #3 → must be ≥8
- **Merged PRs count**: Count items from search #4 → must be ≥3

## Step 4: Generate personalized email

Create email with:
- Pass/fail for each metric
- Brief recommendations (3 help-wanted issues from search #5)
  - **IMPORTANT**: Each recommendation MUST include the full GitHub URL
  - Format: "Title (repo #number) - https://github.com/org/repo/issues/number"
  - Example: "Fix bug in UI (kubestellar/ui #2275) - https://github.com/kubestellar/ui/issues/2275"
- Simple plain text format

## Step 5: Send email

Use send_email tool with subject and body. **IMPORTANT**: Include "To: {email}" as the FIRST line of the body so the safe-output job can extract it:

```
send_email(
  subject="KubeStellar Metrics - @{username} - PASS", 
  body="To: {email}\n\nHey {username},...actual email content..."
)
```

## Important

- Process **clubanderson FIRST**, then **kproche SECOND**
- Each gets their own email with their own metrics
- Keep emails brief and actionable

