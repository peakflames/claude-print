# Release Protocol

Perform a release of claude-print.

**You MUST follow every step in order. Do NOT skip steps. Confirm each step completes before proceeding.**

## Prerequisites

- Must be on `develop` branch. If not, abort and tell the user.
- Check for uncommitted changes. If any exist, abort and tell the user.
- Read the version from `cmd/claude-print/main.go` (the `var version = "X.Y.Z"` line). This is the **release version** for all steps below. Refer to it as `VERSION` throughout.
- If the user provided a version argument, verify it matches the source version. If it doesn't match, abort and tell the user.
- Confirm the version with the user before proceeding.

## Step 1: Finalize CHANGELOG

- In `CHANGELOG.md`, replace `## [Unreleased]` with `## [Unreleased]\n\n## [VERSION] - YYYY-MM-DD` using today's date.
- If there are items listed under `[Unreleased]`, move them under the new version heading.
- Commit: `chore: release vVERSION`
- Push: `git push`

## Step 2: Merge to main

```bash
git checkout main && git pull origin main
git merge develop --no-ff -m "Merge branch 'develop' into main for release vVERSION"
git push
```

## Step 3: Tag the release (on main)

Ask the user for a brief release description, then:

```bash
git tag -a vVERSION -m "Release vVERSION - <description>"
git push origin vVERSION
```

## Step 4: Merge back to develop

```bash
git checkout develop && git merge main --no-ff -m "Merge branch 'main' back into develop after vVERSION release"
```

## Step 5: Post-release version bump (on develop)

Calculate the next minor version by incrementing the minor component of `VERSION` (e.g. `0.3.0` → `0.4.0`).

- Update `cmd/claude-print/main.go`: change `var version = "VERSION"` to `var version = "<next_version>"`
- Ensure `CHANGELOG.md` has an empty `## [Unreleased]` section at the top (below the header)
- Run: `uv run make.py fmt && uv run make.py vet`
- Commit: `chore: bump version to <next_version> for next development cycle`
- Push: `git push`

## Completion

Report a summary of what was done: version released, tag pushed, new development version.

## Reminder

**You MUST follow every step in order. Do NOT skip steps. Confirm each step completes before proceeding.**

