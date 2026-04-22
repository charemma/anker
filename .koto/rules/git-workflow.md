# Git Workflow Rules
<!-- Derived from ~/.claude/rules/workflow.md -- keep in sync on major updates -->

## Commits

Follow Conventional Commits: `<type>(<scope>): <description>`.

## Review Fixes

When implementing changes requested in a code review, use fixup commits:

```
git commit --fixup=<original-commit-sha>
```

One fixup per suggestion. Do not amend or squash during review -- reviewers need to see what changed. After final approval, clean up:

```
git rebase --autosquash main
```

## Branches

Feature branches only. Never commit directly to main.
