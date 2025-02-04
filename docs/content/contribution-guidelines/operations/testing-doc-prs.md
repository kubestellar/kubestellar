# Testing a KubeStellar documentation PR

If a contributor has _**not**_ created a sharable preview of a documentation PR [as documented in the documents management overview](document-management.md#serving-up-documents-globally-from-a-fork-of-the-repository-via-github) , here are the steps to checkout a git pull request for local testing.

## STEP 1: Checkout the Pull Request**

Helpers: [GitHub](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/reviewing-changes-in-pull-requests/checking-out-pull-requests-locally), [DevOpsCube](https://devopscube.com/checkout-git-pull-request/)

Following is one approach to checking out the branch that a PR asks to merge. Alternatively you could use any other technique that accomplishes the same thing.

### 1.1 Use `git fetch` to get a local copy of the PR's branch (note: be sure to check out the right PR!)

Fetch the reference to the pull request based on its ID number, creating a new branch locally. Replace ID with your PR # and BRANCH_NAME with the desired branch name. The branch name will be used only in your local workspace; you can pick anything you like.

The following command assumes that your local workspace has a "git remote" named "upstream" that refers to the shared repository at `github.com/kubestellar/kubestellar`.

```shell
git fetch upstream pull/ID/head:BRANCH_NAME
```

### 1.2 Switch to the new branch

Checkout the BRANCH_NAME where you have all the changes from the pull request.

```shell
git checkout BRANCH_NAME
```

At this point, you can do anything you want with this branch. You can run some local tests, or merge other branches into the branch.

## STEP 2: Test and Build the Documentation (optional)**

See [Serving up documents locally](document-management.md#serving-up-documents-locally) for how
to view and modify the documentation in the branch that you have checked out.
