# Contributing to KubeStellar

Thank you for your interest in contributing to KubeStellar! We welcome contributions from everyone—whether you’re new to open source or a seasoned maintainer.

---

## ⚡ Quick Start

1. **Fork** the repo
2. **Create a branch:**<br>
   Use a prefix for clarity:
   - `feat/` for new features
   - `fix/` for bug fixes
   - `docs/` for documentation
   - Example:
     ```sh
     git checkout -b docs/update-contributing-guide
     ```
3. **Edit**: Make your changes and add tests.
4. **Commit**: Sign your commits (`git commit -s -m "message"`).
5. **Push & PR**: Push your branch and open a Pull Request.

---

## 📑 Table of Contents
- [How to Contribute](#-how-to-contribute)
- [Code Style & Guidelines](#-code-style--guidelines)
- [DCO Sign-Off](#dco-sign-off)
- [Referencing Issues & EPICs](#referencing-issues--epics)
- [Pull Request Review Process](#-pull-request-review-process)
- [Governance & Decision-Making](#-governance--decision-making)
- [Communication & Etiquette](#-communication--etiquette)
- [Code of Conduct](#-code-of-conduct)
- [Helpful Links](#-helpful-links)

---

## 🚀 How to Contribute

### 1. Open an Issue
- Check [existing issues](https://github.com/kubestellar/kubestellar/issues) to avoid duplicates.
- [Open a new issue](https://github.com/kubestellar/kubestellar/issues/new/choose) for bugs, feature requests, or questions.
- For large changes, discuss your idea in an issue before starting work.

### 2. Fork & Clone the Repository
- Click **Fork** on [kubestellar/kubestellar](https://github.com/kubestellar/kubestellar).
- Clone your fork:
  ```sh
  git clone https://github.com/<your-username>/kubestellar.git
  cd kubestellar
  ```

### 3. Create a Branch
- Use a descriptive branch name:
  ```sh
  git checkout -b fix-typo-in-readme
  ```

### 4. Make Changes
- Follow the [Code Style & Guidelines](#code-style--guidelines).
- Add tests if applicable.

## 🛡️ DCO Sign-Off

All contributions to KubeStellar require a Developer Certificate of Origin (DCO) sign-off. This is a simple way to certify that you wrote or have the right to submit the code you are contributing.

- **How to sign your commits:**
  ```sh
  git commit -s -m "Your commit message"
  ```
- This adds a `Signed-off-by` line to your commit, as required by the [DCO](https://developercertificate.org/).
- Commits without a DCO sign-off will fail automated checks.

---

## 🔗 Referencing Issues & EPICs

- Reference related issues and EPICs in your commit messages and pull requests.
- **Commit message example:**
  ```
  Fix controller bug

  Addresses #1234. See EPIC #5678.
  ```
- In your PR description, use `Fixes #<issue-number>`, `Closes #<issue-number>`, or `Related to #<issue-number>` to automatically link your PR to issues.

---

## 🧑‍💻 Code Style, Testing & Guidelines

### Go
- Use `gofmt`, `goimports`, and `golangci-lint` (run `make lint` before submitting).
- Follow idiomatic Go and [Effective Go](https://go.dev/doc/effective_go).
- Place main executables in `/cmd` and core logic in `/pkg`.
- **Run tests:**
  ```sh
  go test ./...
  ```

### Shell Scripts
- Use `#!/usr/bin/env bash` as the shebang.
- Follow [ShellCheck](https://www.shellcheck.net/) recommendations.
- Prefer `set -euo pipefail` for safety.

### Python
- Follow [PEP 8](https://peps.python.org/pep-0008/) style.
- Use `black` and `flake8` for formatting and linting.
- **Run tests:**
  ```sh
  pytest
  ```

### TypeScript
- Use [Prettier](https://prettier.io/) and [ESLint](https://eslint.org/) for formatting and linting.
- Prefer modern ES6+ syntax.
- **Run tests:**
  ```sh
  npm test
  ```

---

## 🚦 Pull Request Review Process

1. **Automatic Checks:** CI runs lint, test, and DCO checks on every PR.
2. **Review:** At least one maintainer reviews your PR. You may be asked to make changes.
3. **Approval:** Once approved and all checks pass, a maintainer will merge your PR.
4. **Merge:** PRs are usually merged via "Squash and Merge" for a clean history.

---

## 🏛️ Governance & Decision-Making

- KubeStellar is a CNCF Sandbox project. Major decisions are made by project maintainers, following community input and consensus when possible.
- Proposals for significant changes should be discussed in a GitHub issue or in community meetings.
- Maintainers are listed in the [OWNERS](OWNERS) file (if present) or in the repository README.

---

## 🤝 Communication & Etiquette

- Be respectful, inclusive, and constructive in all interactions.
- Use GitHub issues for technical discussions and feature requests.
- Join our community Slack for quick questions or informal discussion.
- Follow the [CNCF Code of Conduct](https://github.com/cncf/foundation/blob/main/code-of-conduct.md) at all times.

---

## 📜 Code of Conduct

Participation in KubeStellar is governed by our [Code of Conduct](./CODE_OF_CONDUCT.md). By contributing, you agree to follow this code and help create a welcoming, inclusive environment for all.

---

## 🔗 Helpful Links

- **Repository:** [kubestellar/kubestellar](https://github.com/kubestellar/kubestellar)
- **Documentation:** [KubeStellar Docs](https://kubestellar.io/docs/)
- **Code of Conduct:** [CODE_OF_CONDUCT.md](./CODE_OF_CONDUCT.md)
- **Open Issues:** [GitHub Issues](https://github.com/kubestellar/kubestellar/issues)
- **Community Slack:** [CNCF Slack #kubestellar](https://slack.cncf.io/) (join and find `#kubestellar`)
- **Mailing List:** [KubeStellar Google Group](https://groups.google.com/g/kubestellar)

---

Thank you for helping make KubeStellar better! If you have questions, feel free to open an issue or reach out on Slack or the mailing list.