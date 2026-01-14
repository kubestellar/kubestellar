---
sidebar_position: 5
---

# Contributing to A2A

Greetings! 👋 We’re grateful for your interest in contributing to the **A2A project**, part of the [KubeStellar](https://github.com/kubestellar) ecosystem.
Your contributions — whether raising issues, enhancing documentation, fixing bugs, or improving the UI — are essential to our success.

This document adapts the main [KubeStellar contributing guidelines](https://github.com/kubestellar/kubestellar/blob/main/CONTRIBUTING.md) to A2A, focusing on website, UI, and docs contributions.

---

## 📌 General Practices

* Please read and follow our [Code of Conduct](https://github.com/kubestellar/kubestellar/blob/main/CODE_OF_CONDUCT.md).
* Join the discussion on [Slack (KubeStellar-dev)](https://cloud-native.slack.com/archives/C097094RZ3M).
* Most work happens in GitHub (issues, pull requests, discussions).

---

## 🐞 Issues

* Before opening a new issue, **search the [A2A issue tracker](https://github.com/kubestellar/a2a/issues)** (open + closed).
* If no existing issue matches your problem, feel free to [create one](https://github.com/kubestellar/a2a/issues/new).
* We label beginner-friendly items as [good first issue](https://github.com/kubestellar/a2a/labels/good%20first%20issue) and broader items as [help wanted](https://github.com/kubestellar/a2a/labels/help%20wanted).
* To claim an issue, comment `/assign`. To release it, comment `/unassign`.

### Slash Commands

A2A (via Prow bots) supports slash commands:

* **Issue commands**: `/assign`, `/unassign`, `/good-first-issue`, `/help-wanted`
* **PR commands**: `/lgtm`, `/approve`, `/hold`, `/unhold`, `/retest`

---

## 🔀 Git & Branching Workflow

1. Fork this repo and clone your fork.
2. Keep your `main` branch in sync:

   ```bash
   git checkout main
   git pull upstream main
   ```
3. Create a feature/fix branch:

   ```bash
   git checkout -b fix/button-hover
   ```
4. Make your changes, then commit with **sign-off**:

   ```bash
   git add .
   git commit -s -m "🐛 fix: improve hover visibility for 'View on GitHub' button"
   ```
5. Push your branch:

   ```bash
   git push origin fix/button-hover
   ```

---

## 📝 Commit & PR Guidelines

* Use [Conventional Commits](https://www.conventionalcommits.org/) with emoji prefixes:

  * ✨ feat: new feature
  * 🐛 fix: bug fix
  * 📖 docs: documentation
  * 💄 style: UI/style updates
  * ♻️ refactor: refactor
* Keep commits focused and descriptive.
* All commits must be **signed** (`git commit -s`) to comply with CNCF’s [Developer Certificate of Origin](#certificate-of-origin).

When opening a PR:

* Reference the related issue (`Fixes #123`).
* Fill in the PR template (problem, solution, screenshots if UI).
* Keep PRs small and scoped.
* Use an emoji in the title (e.g., `🐛 fix: footer contributing link`).

---

## 🔍 Reviews & Approval

* PRs require **`/lgtm` from a reviewer** and **`/approve` from a maintainer** before merging.
* You cannot `/lgtm` your own PR.
* Reviewers check:

  * Issue is resolved
  * Code follows guidelines
  * UI works in light/dark modes
  * Build/lint/tests pass

---

## 🧪 Testing Locally

Most contributions here are UI and docs. Please:

* Run locally with

  ```bash
  npm install
  npm start
  ```
* Test on desktop, tablet (\~768px), and mobile (\~375px).
* Check both light and dark themes.
* Run build:

  ```bash
  npm run build
  ```

---

## 📜 Licensing

This project is [Apache 2.0 licensed](https://github.com/kubestellar/a2a/blob/main/LICENSE).
By contributing, you agree to the [Developer Certificate of Origin (DCO)](https://github.com/kubestellar/a2a/DCO).


To sign commits:

```bash
git commit -s -m "your message"
```

---

🙌 Thank you for contributing to A2A and helping the KubeStellar community grow!