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

## Understanding the Documentation Architecture

### Overview

This documentation website is a **separate repository** from the main KubeStellar codebase. Here's the key architecture:

```
┌─────────────────────────────────────────────────────────────┐
│  Main KubeStellar Repository                                 │
│  github.com/kubestellar/kubestellar                          │
│                                                               │
│  📁 docs/content/                                            │
│     ├── readme.md                                            │
│     ├── architecture.md                                      │
│     ├── direct/                                              │
│     │   ├── binding.md                                       │
│     │   └── wds.md                                           │
│     └── ... (all documentation content)                      │
└─────────────────────────────────────────────────────────────┘
                          ↓
                    (Fetched via GitHub API)
                          ↓
┌─────────────────────────────────────────────────────────────┐
│  Docs Website Repository (THIS REPO)                         │
│  github.com/kubestellar/docs                                 │
│                                                               │
│  📁 src/app/docs/                                            │
│     ├── page-map.ts     ← Defines navigation structure      │
│     ├── layout.tsx       ← Nextra theme integration         │
│     └── [...slug]/page.tsx  ← Renders fetched content       │
│                                                               │
│  📁 next.config.ts      ← Nextra configuration              │
│  📁 mdx-components.js   ← MDX component mappings            │
└─────────────────────────────────────────────────────────────┘
                          ↓
                    (Built & Deployed)
                          ↓
┌─────────────────────────────────────────────────────────────┐
│  Live Documentation Website                                  │
│  https://kubestellar.io                                      │
└─────────────────────────────────────────────────────────────┘
```

**Important Concepts:**
- ✅ **Content lives in the main KubeStellar repo** (`docs/content/`)
- ✅ **This repo only contains the website framework** (Next.js + Nextra)
- ✅ **Content is fetched dynamically** via GitHub API at build time
- ✅ **Navigation is defined in `page-map.ts`** (not auto-generated from files)

### How Nextra Integration Works

This documentation site is built using **Nextra**, a powerful Next.js-based documentation framework that provides:

- **Static Site Generation (SSG)** for fast loading
- **MDX Support** for rich, interactive documentation
- **Built-in Search** functionality
- **Theme Customization** with dark/light modes
- **Automatic Navigation** generation

#### Key Files and Their Roles

1. **`next.config.ts`** - Main configuration file that:
   - Imports and configures Nextra with `nextra()` function
   - Enables LaTeX support for mathematical expressions
   - Configures search settings
   - Integrates with `next-intl` for internationalization
   - Sets up redirects for various KubeStellar links

2. **`src/app/docs/layout.tsx`** - Docs layout component that:
   - Imports `Layout` from `nextra-theme-docs`
   - Imports the Nextra theme styles
   - Configures custom navbar, footer, and banner components
   - Sets up the sidebar with page map and repository links
   - Enables dark mode and collapsible sidebar sections

3. **`src/app/docs/page-map.ts`** - Dynamic page map builder that:
   - Fetches documentation files from the main KubeStellar GitHub repository
   - Constructs navigation structure dynamically based on GitHub content
   - Supports multiple branches/versions of documentation
   - Filters and organizes content into logical categories
   - Generates routes for each documentation page

4. **`src/app/docs/[...slug]/page.tsx`** - Dynamic page renderer that:
   - Fetches MDX content from GitHub on-demand
   - Compiles and evaluates MDX with custom components
   - Supports version switching via query parameters
   - Handles Mermaid diagrams and other custom components

5. **`mdx-components.js`** - Component mapping file that:
   - Exports MDX components from Nextra theme
   - Allows customization of how markdown elements render
   - Enables adding custom React components to MDX files

### How to Add Documentation from the Main KubeStellar Repository

The documentation content is **NOT stored in this repository**. Instead, it's dynamically fetched from the main KubeStellar repository at build time and runtime.

#### Content Location

All documentation content lives in the main KubeStellar repository:
- **Repository**: `https://github.com/kubestellar/kubestellar`
- **Content Path**: `/docs/content/`
- **Branches**: Supports multiple branches for versioning (e.g., `main`, `release-0.23.0`)

#### How Content is Fetched

The `buildPageMapForBranch()` function in `page-map.ts`:

1. Makes API calls to GitHub to fetch the repository tree
2. Filters for `.md` and `.mdx` files in the `docs/content/` directory
3. Organizes files according to the `CATEGORY_MAPPINGS` structure
4. Creates navigation entries and routes for each file
5. Caches results for performance

#### Adding New Content

To add new documentation pages:

1. **In the Main KubeStellar Repository:**
   - Add your `.md` or `.mdx` file to `/docs/content/` directory
   - Organize it in an appropriate subdirectory
   - Use standard Markdown or MDX syntax
   - Commit and push to the desired branch

2. **In This Docs Repository:**
   - Update `src/app/docs/page-map.ts`
   - Find the appropriate category in `CATEGORY_MAPPINGS`
   - Add an entry for your new file:
     ```typescript
     { file: 'your-new-file.md' }
     // or with custom title
     { 'Custom Title': 'your-new-file.md' }
     ```
   - The file path is relative to `docs/content/` in the main repo

#### Example: Adding a New Getting Started Guide

```typescript
// In src/app/docs/page-map.ts
['Install & Configure', [
  { file: 'pre-reqs.md' },
  { 'Quick Start': 'getting-started.md' },  // Add this line
  { file: 'start-from-ocm.md' },
  // ... rest of the entries
]]
```

#### Adding Nested Sections

For hierarchical navigation:

```typescript
{
  'Section Name': [
    { 'Subsection 1': 'path/to/file1.md' },
    { 'Subsection 2': 'path/to/file2.md' },
    {
      'Nested Section': [
        { 'Deep File': 'path/to/nested/file.md' }
      ]
    }
  ]
}
```

#### External Links

You can also add external documentation links:

```typescript
{ 'API Reference (new tab)': 'https://pkg.go.dev/github.com/kubestellar/kubestellar/api/control/v1alpha1' }
```

### Version Management

The documentation supports multiple versions through the `versions.ts` config:

- **Default Version**: Set in `getDefaultVersion()`
- **Branch Mapping**: Map versions to Git branches in `getBranchForVersion()`
- **Version Switching**: Users can switch versions via query parameter: `?version=0.23.0`

### Testing Your Changes

1. **Local Development:**
   ```sh
   npm run dev
   ```
   - Test navigation and page rendering
   - Verify new pages appear in the correct location
   - Check that links work properly

2. **Build Test:**
   ```sh
   npm run build
   ```
   - Ensure no build errors
   - Verify static generation works
   - Check that all pages are accessible

3. **Content Verification:**
   - Ensure the content file exists in the main KubeStellar repo
   - Verify the file path in `page-map.ts` matches exactly
   - Check that the category structure makes logical sense

### Common Issues

1. **Page Not Appearing:**
   - Verify file exists in main KubeStellar repo
   - Check file path spelling and case sensitivity
   - Ensure file has `.md` or `.mdx` extension
   - Rebuild the page map

2. **Navigation Issues:**
   - Check `CATEGORY_MAPPINGS` structure syntax
   - Ensure proper nesting of arrays and objects
   - Verify route generation logic

3. **Content Not Updating:**
   - Clear Next.js cache: `npm run clean`
   - Rebuild: `npm run build`
   - Check GitHub API rate limits
   - Verify `GITHUB_TOKEN` environment variable if needed

### Working with MDX

MDX allows you to use React components in Markdown:

```mdx
# My Documentation

<Callout type="info">
  This is an info callout!
</Callout>

<Tabs items={['npm', 'yarn', 'pnpm']}>
  <Tabs.Tab>npm install kubestellar</Tabs.Tab>
  <Tabs.Tab>yarn add kubestellar</Tabs.Tab>
  <Tabs.Tab>pnpm add kubestellar</Tabs.Tab>
</Tabs>
```

Available components:
- `Callout` - For notes, warnings, and tips
- `Tabs` - For tabbed content
- `Mermaid` - For diagrams (custom component)

### Quick Reference: Common Workflows

#### Workflow 1: Adding a New Documentation Page

```sh
# Step 1: Add content to main KubeStellar repo
cd /path/to/kubestellar
echo "# My New Page" > docs/content/my-new-page.md
git add docs/content/my-new-page.md
git commit -m "Add new documentation page"
git push

# Step 2: Update navigation in docs repo
cd /path/to/docs
# Edit src/app/docs/page-map.ts to add your page
# Add: { file: 'my-new-page.md' } in appropriate category

# Step 3: Test locally
npm run dev
# Visit http://localhost:3000/docs to verify

# Step 4: Commit and push
git add src/app/docs/page-map.ts
git commit -m "Add my-new-page to navigation"
git push
```

#### Workflow 2: Reorganizing Navigation

```sh
# Edit src/app/docs/page-map.ts
# Modify CATEGORY_MAPPINGS array
# Example: Move a page to different category
npm run dev  # Test changes
npm run build  # Verify build succeeds
git commit -am "Reorganize documentation navigation"
```

#### Workflow 3: Updating Nextra Configuration

```sh
# Edit next.config.ts for Nextra settings
# Example: Enable/disable features
npm run dev  # Test configuration
npm run build  # Verify no errors
git commit -am "Update Nextra configuration"
```

#### Workflow 4: Adding Custom MDX Components

```sh
# Step 1: Create your component
echo 'export function MyComponent() { return <div>Hello</div> }' > src/components/MyComponent.tsx

# Step 2: Export from mdx-components.js
# Add: import { MyComponent } from './src/components/MyComponent'
# Add to export: MyComponent

# Step 3: Use in documentation (main repo)
# In any .mdx file: <MyComponent />

npm run dev  # Test component
```

### Environment Variables

For development and production, you may need these environment variables:

```sh
# .env.local (optional)
GITHUB_TOKEN=ghp_your_token_here  # For higher GitHub API rate limits
GH_TOKEN=ghp_your_token_here      # Alternative name
GITHUB_PAT=ghp_your_token_here    # Alternative name
```

**When is GITHUB_TOKEN needed?**
- When fetching content frequently during development
- To avoid GitHub API rate limiting (60 requests/hour without token, 5000 with token)
- Not required for basic local development

### Key Files Summary

| File | Purpose | When to Edit |
|------|---------|--------------|
| `src/app/docs/page-map.ts` | Navigation structure | Adding/removing/reorganizing pages |
| `next.config.ts` | Nextra & Next.js config | Changing Nextra settings, redirects |
| `src/app/docs/layout.tsx` | Docs page layout | Modifying sidebar, theme, or layout |
| `mdx-components.js` | MDX component mappings | Adding custom React components to MDX |
| `src/config/versions.ts` | Version management | Adding new documentation versions |
| `src/middleware.ts` | Route handling | Changing i18n behavior, route matching |
| `package.json` | Dependencies & scripts | Adding new packages or commands |

### Debugging Tips

**Problem: Page not showing up**
```sh
# Check if file exists in main repo
curl https://api.github.com/repos/kubestellar/kubestellar/contents/docs/content/your-file.md

# Verify page-map.ts entry
grep -r "your-file.md" src/app/docs/page-map.ts

# Clear Next.js cache
npm run clean
npm run dev
```

**Problem: Build fails**
```sh
# Check TypeScript errors
npm run type-check

# Check linting
npm run lint

# View detailed build output
npm run build 2>&1 | tee build.log
```

**Problem: Styling issues**
```sh
# Check Tailwind classes
npm run build

# Inspect global CSS
cat src/app/globals.css

# Check theme styles
cat node_modules/nextra-theme-docs/style.css
```

### Contributing Checklist

Before submitting your PR, ensure:

- [ ] Code follows existing style and conventions
- [ ] All links work correctly
- [ ] Navigation structure is logical
- [ ] Local build succeeds (`npm run build`)
- [ ] No TypeScript errors (`npm run type-check`)
- [ ] No linting errors (`npm run lint`)
- [ ] Changes are documented in PR description
- [ ] Related issue is referenced (if applicable)
- [ ] Screenshots included for UI changes (if applicable)

---

## Need Help?

If you have questions, open an issue or ask in the community channels:

- **Slack**: [#kubestellar-dev](https://cloud-native.slack.com/archives/C097094RZ3M)
- **GitHub Issues**: [kubestellar/docs](https://github.com/kubestellar/docs/issues)
- **Community Meetings**: Check the [community calendar](https://calendar.google.com/calendar/event?action=TEMPLATE&tmeid=MWM4a2loZDZrOWwzZWQzZ29xanZwa3NuMWdfMjAyMzA1MThUMTQwMDAwWiBiM2Q2NWM5MmJlZDdhOTg4NGVmN2ZlOWUzZjZjOGZlZDE2ZjZmYjJmODExZjU3NTBmNTQ3NTY3YTVkZDU4ZmVkQGc)

### Additional Resources

- **Nextra Documentation**: [https://nextra.site](https://nextra.site)
- **Next.js Documentation**: [https://nextjs.org/docs](https://nextjs.org/docs)
- **MDX Documentation**: [https://mdxjs.com](https://mdxjs.com)
- **Main KubeStellar Repo**: [https://github.com/kubestellar/kubestellar](https://github.com/kubestellar/kubestellar)

Thank you for contributing to our documentation! 🚀
