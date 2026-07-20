## Summary
This PR adds a pre-commit hook for documentation spell checking using PySpelling. This runs spellcheck locally to catch typos before code/documentation is committed and pushed.

## Related issue(s)
Fixes #11

## Changes
- **Pre-commit hook**: Added the `pyspelling` hook to `.pre-commit-config.yaml` pointing to the local configuration in `.github/spellcheck/.spellcheck.yml`.
- **Pre-commit setup and dependencies**: Updated `CONTRIBUTING.md` with instructions on installing system dependencies (`aspell` and `pre-commit`) and setting up/running pre-commit hooks.
- **Spellcheck configuration fixes**: Fixed `.github/spellcheck/.spellcheck.yml` to remove the invalid `ignore_classes` parameter and use standard CSS selector ignores instead. Added `markdown.extensions.fenced_code` to exclude triple-backtick code blocks.
- **Typos and Wordlist**: Fixed typos in `monitoring/README.md`, `test/scale-infra/README.md`, `test/scale-infra/INSTRUCTIONS.md`, and `test/performance/latency-controller/README.md`. Appended missing valid technical terms to `.github/spellcheck/.wordlist.txt`.

## Testing
- Installed `aspell` and `pre-commit` and ran `pre-commit run pyspelling --all-files` locally to verify that all spell checks pass.

## Checklist
- [x] Spellcheck pre-commit hook configured in `.pre-commit-config.yaml`
- [x] Pre-commit hook documentation and dependency details added to `CONTRIBUTING.md`
- [x] Spellcheck configuration fixed and exclusions for code blocks/deprecated folders set up
- [x] Existing typos in documentation fixed
- [x] Wordlist updated with missing technical terms
- [x] Local pre-commit hook tests run and passed successfully
