# Release Protocol

Perform a release of claude-print.

**You MUST follow every step in order. Do NOT skip steps. Confirm each step completes before proceeding.**

## Smart Mode Assumptions (When run with no arguments)

When you run this command without arguments, make these intelligent assumptions:

- **Branch handling**: If not on `develop`, offer to switch automatically. Stash uncommitted changes if needed.
- **Version**: Auto-detect from `cmd/claude-print/main.go`. This is authoritative.
- **Release description**: Scan the `## [Unreleased]` section in CHANGELOG.md. If it has items, suggest a concise description based on them. If empty, ask the user briefly.
- **Next version**: Auto-calculate by incrementing the minor version component (e.g., 0.2.0 → 0.3.0).
- **Confirmation**: Show a summary of detected values and ask for approval before proceeding. Allow the user to override any value.

## Prerequisites

- Must be on `develop` branch. If not on develop:
  - Offer to stash changes and switch to develop
  - Only proceed if user confirms
- Check for uncommitted changes. If any exist:
  - Offer to stash them (they'll be restored after release)
  - Only abort if user declines stashing
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

**⚠️ IMPORTANT: Pushing the tag triggers the CI/CD build workflow!**

When you push this tag, GitHub Actions will automatically:
1. Run quality checks (format, vet, tests)
2. Build binaries for all platforms (Windows, macOS Intel/ARM, Linux)
3. Create a GitHub Release with downloadable binaries

Determine release description:
- If running with `--description` flag, use that value
- If running without args, auto-generate from CHANGELOG `[Unreleased]` section:
  - If section has items, create a one-line summary (e.g., "Improve display formatting and add bug fixes")
  - If section is empty, ask user briefly: "Brief release description?"
- Show the final description to user and ask for approval

Create and push the tag (this starts the build):

```bash
git tag -a vVERSION -m "Release vVERSION - <description>"
git push origin vVERSION
```

**Build Workflow Triggers Automatically**
- Monitor at: `https://github.com/peakflames/claude-print/actions`
- The workflow must pass all quality checks before building
- If it fails, check the logs and fix issues on develop, then re-tag and retry
- On success, binaries appear in the GitHub Release (takes ~2-5 minutes)

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

## Step 6: Verify Build Initiated (CRITICAL)

After pushing the tag, verify that the CI workflow started **within 30 seconds**.

```bash
gh run list --workflow release.yml --limit 1
```

Or visit: `https://github.com/peakflames/claude-print/actions`

### Expected Status
- ✓ "In progress" or "In queue" (build is running)
- ✓ "✓ Completed" (build finished successfully)
- ✗ "✗ Failed" - Review the error and recovery steps below

### If Workflow Didn't Trigger
If you don't see a new run for v0.2.0 within 30 seconds:

1. Verify the tag was actually pushed:
   ```bash
   git ls-remote origin v0.2.0
   ```

2. If tag exists on remote but workflow didn't run:
   - Delete the tag and re-push to force trigger:
   ```bash
   git tag -d vVERSION
   git push origin :vVERSION
   git tag -a vVERSION -m "Release vVERSION - <description>" <commit-hash>
   git push origin vVERSION
   ```

3. Check GitHub Actions settings - ensure workflows are enabled for your repository

### If Quality Checks Fail

The build will be aborted. You'll need to:

1. Fix the issues on develop branch
2. Commit and push the fix
3. Delete the failed tag:
   ```bash
   git tag -d vVERSION
   git push origin :vVERSION
   ```
4. Re-tag and push the corrected version from step 3 above

## Completion

Report a summary of what was done:
- Version released: vVERSION
- Tag pushed to remote (and build workflow triggered automatically)
- New development version: <next_version>
- Build status: [Link to GitHub Actions run or "monitor at https://github.com/peakflames/claude-print/actions"]

Junior engineers should understand that the release is not "complete" until the CI workflow succeeds and binaries are published.

## Agent Implementation Notes

These are guidelines for the agent executing this protocol:

### Smart Defaults Strategy

1. **Version detection**: Parse `cmd/claude-print/main.go` to extract the version. This is always authoritative—never ask the user to provide it.

2. **Changelog scanning**: Read `CHANGELOG.md` to see what's in `[Unreleased]`:
   - If it contains items (features, fixes, changes), summarize them into 1-2 sentences as the default description
   - If empty, ask minimally: "Enter a brief release description (or press Enter to skip)"
   - Show the auto-generated description to the user and let them approve/modify it

3. **Next version calculation**: Always increment the minor version (e.g., 0.2.0 → 0.3.0)

4. **Pre-release checklist**: Show the user a confirmation screen with:
   - Current version being released
   - Suggested release description
   - Next development version
   - All git operations that will be performed
   - "Proceed?" prompt

### Branch and Stash Handling

- If not on `develop`, **offer** to switch rather than abort:
  - "Current branch is X. Switch to develop and proceed?"
- If uncommitted changes exist, **offer** to stash:
  - "Found X uncommitted changes. Stash them temporarily?"
  - After release completes, restore stashed changes with `git stash pop`
  - Only abort if user explicitly declines

### Build Workflow Integration

The **tag push in Step 3 is critical** — it automatically triggers the release workflow that:
- Runs quality checks (code formatting, go vet, tests)
- Builds binaries for all platforms
- Creates GitHub Release with downloadable artifacts

**IMPORTANT**: The workflow MUST start within 30 seconds. If it doesn't appear:
- Check if tag actually exists on remote: `git ls-remote origin vVERSION`
- If tag exists but no workflow run appears, re-push the tag to force trigger
- This is a known GitHub issue - retrying usually fixes it

After tag push, immediately show user:
```
✓ Tag vVERSION pushed successfully
  Build workflow triggered automatically!
  Monitor progress: https://github.com/peakflames/claude-print/actions
  (builds typically complete in 2-5 minutes)

⏳ Checking workflow status...
[Wait 30 seconds and verify workflow appears in Actions tab]
```

In Step 6 (verification), show:
- Workflow run details (status, which stage it's on)
- Direct link to the run
- What to do if workflow didn't appear (re-push recovery steps)
- What to do if workflow failed (debugging steps)

Never proceed to next steps or consider release "done" until:
1. Workflow has started (visible in GitHub Actions)
2. All quality checks passed
3. Binaries were built successfully

### User Experience Goals

- Minimize interaction: Junior engineers should never need to know git commands
- Maximize clarity: Each step should report what happened and why
- Enable recovery: If any step fails, suggest next steps clearly
- Build transparency: Make it obvious that tag push triggers the CI build and when it's complete

## Reminder

**You MUST follow every step in order. Do NOT skip steps. Confirm each step completes before proceeding.**

