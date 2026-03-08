# Git Hooks Configuration

This project uses git hooks for automated checks before commits.

## Setup

To enable the pre-commit hook:

```bash
chmod +x .hooks/pre-commit
ln -sf ../../.hooks/pre-commit .git/hooks/pre-commit
```

Or on Windows:

```powershell
copy .hooks\pre-commit .git\hooks\pre-commit
```

## Available Hooks

- **pre-commit**: Runs `go vet` and `go test` before each commit

## Branch Protection

For remote repositories, configure branch protection rules:

1. **main** branch:
   - Require pull request reviews before merging
   - Require status checks to pass
   - Require linear history (no merge commits)
   - Restrict who can push

2. **develop** branch (if using GitFlow):
   - Require status checks to pass
   - Allow squash merging

## Recommended Git Workflow

```bash
# Create feature branch
git checkout -b feature/your-feature

# Make changes and commit
git add .
git commit -m "feat: add your feature"

# Push and create PR
git push origin feature/your-feature
```
