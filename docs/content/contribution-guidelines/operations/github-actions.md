# GitHub Action Reference Discipline

For the sake of supply chain security, every reference from a workflow to an action identifies the action's version by a commit hash rather than a tag or branch name. This ensures reproducibility and prevents supply chain attacks through action tampering.

## The Reversemap File

The file `.gha-reversemap.yml` in the root of the repository is the single source of truth for the mapping from action identity (owner/repo and version tag) to commit hash. This file should only be updated when you have confidence in the new or added version.

## Managing Action References

The script `hack/gha-reversemap.sh` provides commands for managing GitHub Action references across workflows.

### Available Commands

| Command | Description |
|---------|-------------|
| `update-action-version` | Updates an action to its latest version in the reversemap file |
| `apply-reversemap` | Distributes the reversemap specifications to all workflow files |
| `verify-mapusage` | Verifies that all workflow files use correct commit hashes |

### Updating an Action

To update an action (e.g., `actions/checkout`) to the latest version:

```shell
hack/gha-reversemap.sh update-action-version actions/checkout
hack/gha-reversemap.sh apply-reversemap
```

The first command updates `.gha-reversemap.yml` with the latest commit hash for the action. The second command propagates this change to all workflow files that reference the action.

### Verifying Action References

To verify that all workflow files use the correct commit hashes:

```shell
hack/gha-reversemap.sh verify-mapusage
```

This command checks all workflow files and reports any discrepancies between the reversemap file and actual workflow references.

## GitHub API Rate Limiting

The `hack/gha-reversemap.sh` script makes calls to the GitHub API, which is rate-limited. If you encounter rate limit errors, you can authenticate using a GitHub token:

```shell
export GITHUB_TOKEN=your_token_here
hack/gha-reversemap.sh update-action-version actions/checkout
```

Authenticated requests have significantly higher rate limits than unauthenticated requests.

## Example Workflow

Here's a typical workflow for updating GitHub Actions:

1. **Check current status**: Run `hack/gha-reversemap.sh verify-mapusage` to see if any actions need updating.

2. **Update actions**: For each action that needs updating:
   ```shell
   hack/gha-reversemap.sh update-action-version owner/action-name
   ```

3. **Apply changes**: Propagate the updates to all workflow files:
   ```shell
   hack/gha-reversemap.sh apply-reversemap
   ```

4. **Verify**: Confirm all references are correct:
   ```shell
   hack/gha-reversemap.sh verify-mapusage
   ```

5. **Commit**: Commit both the updated `.gha-reversemap.yml` and all modified workflow files.

## Why Commit Hashes?

Using commit hashes instead of tags provides several security benefits:

- **Immutability**: Commit hashes cannot be changed, while tags can be moved to point to different commits.
- **Verification**: You can verify exactly what code will run by inspecting the specific commit.
- **Supply Chain Security**: Prevents attacks where a malicious actor compromises an action and moves a tag to point to malicious code.
- **Reproducibility**: Builds are reproducible because the exact same action code runs every time.
