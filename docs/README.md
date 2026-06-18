# These Are Not The Docs You Are Looking For

The documentation for KubeStellar has been moved to a separate repository, [https://github.com/kubestellar/docs](https://github.com/kubestellar/docs) to be rendered as part of the new (2026) consolidated kubestellar.io site.

**Do NOT open issues or PRs against anything in the docs folder of this repository.**

The previous docs folder contents in this repository have been moved into a docs-to-be-deleted folder  as a precaution while we confirm that there are no omissions in the files copied into the docs repository

## How to make a docs PR for Kubestellar

### A. The easy way

For simple edits to a single page

1. Sign into Github in your browser.
2.  Open a second tab and visit the page in the website you wish to modify. (Make sure have selected the specific version of the docs with the dropdown in the masthead)
3. Find and click on the Edit This Page (Pencil) icon near the upper right page
4. A Github editor session will open for you and when you commit your changes, you will be presented with the option to create a corresponding PR. 
5. You may have to make some adjustments to the PR title, etc to follow our rules.

### B. The complicated way

For less simple edits, for edits across multiple files, or for editing the docs site structure/navigation, you will have to go the more traditional GitHub route of:
1. creating a fork of the docs repository, 
2. editing the files
3. committing changes to the branch
   _be sure to both sign off (-s option) for DCO    and sign (-S option) your commits_
3. pushing those changes up to your fork 
4. and then doing a standard Pull Request.

## Don't Waste Your or the Reviewers' Time

Docs PRS _for the website_ submitted from the **kubestellar/kubestellar** repository instead of the **kubestellar/docs** repository will be closed without further review. 