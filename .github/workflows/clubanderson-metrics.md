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
        description: 'Select maintainer to audit'
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
            --limit 100 \
            --json number,title,url,closedAt,labels > /tmp/prs-merged-raw.json
          
          # Manually count and create JSON
          count=$(jq '. | length' /tmp/prs-merged-raw.json)
          echo "Found $count merged PRs for $username"
          jq --argjson count "$count" '{total_count: $count, items: .}' /tmp/prs-merged-raw.json \
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

**Step 2: Extract metrics from the displayed data**

Look at the JSON output from Step 1 and find:

- **Help-wanted count**: In the first JSON, find the number after `"total_count":`. Example: if you see `{"total_count": 11, "items": [...]` then the count is **11**.
- **Merged PRs count**: In the second JSON, find the number after `"total_count":`. Example: if you see `{"total_count": 26, "items": [...]` then the count is **26**.
- **PR reviews count**: 
  1. Look in the third JSON (commented merged) - list all numbers that appear after `"number":` 
  2. Look in the fourth JSON (commented open) - list all numbers that appear after `"number":`
  3. Combine both lists and count UNIQUE numbers (no duplicates)
  
**Example for PR reviews:**
- If commented-merged shows: `"number": 100`, `"number": 101`, `"number": 102`
- And commented-open shows: `"number": 102`, `"number": 103`
- Unique numbers are: 100, 101, 102, 103 = **4 unique PRs**

**IMPORTANT:**
- ✅ DO: Use `cat` to read the files (you have permission)
- ✅ DO: Look at the visible JSON and extract `total_count` manually
- ✅ DO: Count unique PR numbers by listing them out
- ❌ DO NOT use jq, python, node - they don't work here
- ❌ DO NOT try to execute complex commands

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

Here are your KubeStellar metrics for the last 60 days:

✅/❌ Help-Wanted Issues: X created (required: ≥2)
✅/❌ Merged PRs: Z merged (required: ≥3)
✅/❌ PR Reviews: Y unique PRs (required: ≥8)

Overall: PASS [3/3] or FAIL [1/3]

[If FAIL: Brief encouragement to focus on the missing criteria]

---

🏷️ Help-Wanted Suggestions for You:
        1. Title (repo #123) - https://github.com/... - Brief reason
        2. Title (repo #456) - https://github.com/... - Brief reason  
        3. Title (repo #789) - https://github.com/... - Brief reason

🔨 PR Opportunities in Your Areas:
        1. Title (repo #123) - https://github.com/... - Brief reason
        2. Title (repo #456) - https://github.com/... - Brief reason
        3. Title (repo #789) - https://github.com/... - Brief reason

👀 PRs Needing Your Review:
        1. Title (repo #123) - https://github.com/... - Brief reason
        2. Title (repo #456) - https://github.com/... - Brief reason
        3. Title (repo #789) - https://github.com/... - Brief reason

**CRITICAL FORMATTING INSTRUCTIONS - READ CAREFULLY:**

When you generate the email body, each recommendation line MUST be formatted with EXACTLY 8 spaces of indentation.

**Correct indentation (copy this exactly):**
```
🏷️ Help-Wanted Suggestions for You:
        1. Fix UI bug (kubestellar/ui #2275) - https://github.com/kubestellar/ui/issues/2275 - Matches your UI expertise
        2. Add docs (kubestellar/docs #123) - https://github.com/kubestellar/docs/issues/123 - Documentation work
        3. Improve tests (kubestellar/ui #456) - https://github.com/kubestellar/ui/issues/456 - Testing expertise
```

Count the spaces: "        1." = 8 spaces + "1." + space + text

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
