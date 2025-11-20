# Contributing to Kubestellar Docs

Thank you for your interest in contributing to our documentation repository! We welcome contributions from everyone. Please follow these guidelines to help maintain a high-quality, consistent, and collaborative project.

---

## Prerequisites

Before contributing, ensure you have:

- [Node.js](https://nodejs.org/) (version 18 or higher) installed
- [npm](https://www.npmjs.com/) installed
- A GitHub account
- Basic knowledge of Markdown and Git

---

## How to Contribute

### 1. Fork the Repository

Click the **Fork** button at the top-right corner of this page to create your own copy of the repository.

### 2. Clone Your Fork

Clone the repository to your local machine:

```sh
git clone https://github.com/your-username/docs.git
```

### 3. Install Dependencies

Navigate into the project directory and install dependencies:

```sh
cd docs
npm install
```

### 4. Create a Branch

Create a new branch for your work:

```sh
git checkout -b my-feature-branch
```

### 5. Make Your Changes

Edit or create documentation files as needed.  
Please follow the existing structure, tone, and formatting style.

### 6. Preview / Test Your Changes

Start the development environment to verify rendering:

```sh
npm run dev
```

> **Tip:** During active documentation contributions, regularly run `npm run dev` to preview updates in real time.

### 7. Commit and Push

Commit your changes with a clear and meaningful message:

```sh
git add .
git commit -m "Describe your changes"
git push origin my-feature-branch
```

### 8. Open a Pull Request

Open a Pull Request (PR) from your branch to the main repository.

#### PR Description

- Provide a summary of what you changed (maximum 2 lines).
- Reference related issues, e.g.:
  ```
  Fixes #123
  ```

---

## Contribution Guidelines

- **Write Clearly:** Use concise language and proper formatting.
- **Stay Consistent:** Maintain the existing structure and style.
- **Respect Internationalization Standards:** Avoid pushing raw UI strings directly; always use i18n references.
- **Be Respectful:** Review our Code of Conduct before contributing.

### Caution With AI-Generated Code

> AI tools (like GitHub Copilot or ChatGPT) are helpful but **not always context-aware**.  
> **Please DO NOT blindly copy-paste AI-generated code.**

Before committing:

- Double-check if the code aligns with our project’s architecture.
- Test thoroughly to ensure it doesn’t break existing functionality.
- Refactor and adapt it as per the codebase standards.

---

## Contribution Commands Guide

This guide helps contributors manage issue assignments and request helpful labels via GitHub comments. These commands are supported through GitHub Actions or bots configured in the repository.

### Issue Assignment

- **To assign yourself to an issue**, comment:
  ```
  /assign
  ```
- **To remove yourself from an issue**, comment:
  ```
  /unassign
  ```

### Label Requests via Comments

You can also request labels to be automatically added to issues using the following commands:

- **To request the `help wanted` label**, comment:
  ```
  /help-wanted
  ```
- **To request the `good first issue` label**, comment:
  ```
  /good-first-issue
  ```

These commands help maintainers manage community contributions effectively and allow newcomers to find suitable issues to work on.

---

## Need Help?

If you have questions, open an issue or ask in the community channels.

Thank you for contributing to our documentation! 🚀
