# Contributing to the Docs Repository

Thank you for your interest in contributing to the docs repository! We welcome contributions from everyone. Please follow these guidelines to help us maintain a high-quality and collaborative project.

## How to Contribute

1. **Fork the Repository**
   - Click the "Fork" button at the top right of this page to create your own copy of the repository.

2. **Clone Your Fork**
   - Clone your fork to your local machine:
     ```sh
     git clone https://github.com/your-username/docs.git
     ```

3. **Create a Branch**
   - Create a new branch for your changes:
     ```sh
     git checkout -b my-feature-branch
     ```

4. **Make Your Changes**
   - Edit or add documentation files as needed. Please follow the existing style and structure.

5. **Test Your Changes**
   - If applicable, preview your changes locally to ensure everything renders correctly.

6. **Commit and Push**
   - Commit your changes with a clear message:
     ```sh
     git add .
     git commit -m "Describe your changes"
     git push origin my-feature-branch
     ```

7. **Open a Pull Request**
   - Go to the original repository and open a Pull Request from your branch.
   - **Title:** Your PR title should be descriptive. Please prefix it with `major:`, `minor:`, or `patch:` to indicate the scope of the change, following semantic versioning guidelines.
     - Use `major:` for significant, breaking changes or large new features that are not backward-compatible.
     - Use `minor:` for new features or enhancements that are backward-compatible.
     - Use `patch:` for backward-compatible bug fixes, typo corrections, or small documentation updates.
     - _Example: `patch: Fix typo in installation guide`_
   - **Description:** Provide a concise summary of your changes in the PR description. **This summary must not be longer than two lines**, as it is used to automatically generate progress logs.
   - Reference any related issues in the description (e.g., `Fixes #123`).

## 🧩 Pre-commit Checks (Husky)

This project uses **Husky** to enforce code quality before every commit.

### 🔕 Skipping Pre-commit Checks

If you need to skip Husky checks (for example, when committing documentation-only changes), you can bypass them using:

```bash
git commit -n -m "your_commit_message"
```

or equivalently:

```bash
git commit --no-verify -m "your_commit_message"
```

> ⚠️ Use this only when absolutely necessary.

---

## Guidelines

- **Write Clearly:** Use clear, concise language and proper formatting.
- **Stay Consistent:** Follow the existing file structure and naming conventions.
- **Be Respectful:** Review our [Code of Conduct](docs/contribution-guidelines/coc-inc.md) before contributing.

## Need Help?

If you have questions, open an issue or ask in the community channels.

Thank you for helping improve our documentation!
