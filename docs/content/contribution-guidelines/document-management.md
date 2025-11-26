## Kubestellar Web Documentation Structure

**The KubeStellar website is upgrading from collection of mkdocs-based individual sites for each of its components to a unified Nextra-based site which can display documents imported from all of the component Github repositories in the KubeStellar project** 

The new site is going live in November 2025, and this document is therefore being revised extensively and actively. 
[Information about the legacy website built via mkdocs is still viewable directly in the Github Repository](https://github.com/kubestellar/kubestellar/blob/release-0.29.0/docs/content/contribution-guidelines/operations/mkdocs-management.md)

## Websites

### One Website for All: Kubestellar.io

The new Nextra-based [**Kubestellar.io**](https://kubestellar.io) site has migrated away from a mkdocs-rendered static site to a site using next.js which can support active content. The documentation source files for this new site are actually distributed across the multiple repos in the [Github Kubestellar organization](https://github.com/kubestellar)

---
###  Design Repository: kubestellar/docs

The [KubeStellar/docs](https://github.com/kubestellar/docs) repository contains the Nextra components and source files to generate the main navigation pages and layout of the site.
It uses the remote file capacity of the Nextra engine to pull and render MarkDown (MD) and MDX documentation files from documentation files in the component repositorys of KubeStellar (kubestellar/kubestellar, kubestellar/ui, kubestellar/kubeflex, kubestellar/a2a, etc)

_Note: this means that to modify the style and overall structure of the KubeStellar site, one must work with the files in the docs repository, but to update documentation for Kubestellar or any of it's particular components, one must work on files in that component repository_

Doing a GitHub Pull Request on the docs repository itself will build and deploy a preview of the resulting site via Netlify. Updates the component repository documentation files will not be rendered on the main site until a cron job and/or an admin triggers a refresh.

### Component repositories: kubestellar, kubeflex, etc.

All of the component documentation is updated via commits to [the individual Github repositories](https://github.com/kubestellar) of the KubeStellar project. There may be additional information, scripts, etc. accessible via Github which are not directly exposed on the Nextra-based site.

### Future migration: contribution information

At this time the documentation and instructions, best practices, etc., for contributing to the KubeStellar project are part of the [kubestellar/kubestellar](htpps://github.com/kubestellar/kubestellar) repository. We anticipate moving those files into the kubestellar/docs repository in the future as they have global application to the whole KubeStellar open-source project.

---
### Dual-use documentation sources

The documentation that is rendered from the repositiories to the website is designed so that it can also be usefully viewed
directly at GitHub. For example, you can view this page both (a) on
the website at 
[/docs/contribution-guidelines/document-management/](/docs/contribution-guidelines/document-management/)
and (b) directly in GitHub at
[https://github.com/kubestellar/kubestellar/blob/main/docs/content/contribution-guidelines/document-management.md](https://github.com/kubestellar/kubestellar/blob/main/docs/content/contribution-guidelines/document-management.md). Unfortunately, occasionally
those two uses interpret the documentation sources a bit
differently. To the degree possible, we choose to write the
documentation sources in a way that is rendered consistently in those
two views (for example: indenting with four spaces instead of
three). When that is not possible, considerations for the website
rendering take precedence.

### Style Guide

With more contributors writing pages for our documentation, we are implementing a [Style Guide](./docs-styleguide.md) to help ensure more usable documenations with a consistent style and voice.

---
## Local Development: Preparing to contribute to the KubeStellar website

If you are interested in working on our site, your local system must first be properly configured with node.js and next.js. These instructions are copied from the Local Development section of [README file for kubestellar/docs](https://github.com/kubestellar/docs/blob/main/README.md)

This documentation site is built with [Next.js](https://nextjs.org/), providing a modern, performant documentation experience.

### Prerequisites

- **Node.js** v18.0.0 or higher ([Download](https://nodejs.org/))
- **npm** or **yarn** package manager

**Note for WSL users:** search for hints on the correct procedure to install node.js, npm, and yarn in your environment. 

Verify your Node.js installation:

```bash
node --version
```

### Setup Instructions

1. **Clone the repository:**

   ```bash
   git clone https://github.com/kubestellar/kubestellar.git
   cd docd
   ```

2. **Install dependencies:**

   ```bash
   npm install
   # or
   yarn install
   ```

3. **Start the development server:**

   ```bash
   npm run dev
   # or
   yarn dev
   ```

   The site will be available at `http://localhost:3000` with hot-reload enabled for instant feedback.

4. **Build for production:**

   ```bash
   npm run build
   # or
   yarn build
   ```

5. **Preview the production build:**

   ```bash
   npm start
   # or
   yarn start
   ```

## We look forward to seeing your contributions!

