# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.3.0] - 2026-04-15

### Added

The following was contribute by @lucastansini on  PR-[#1]:

- **Default mode unchanged:** `./claude-print "What is 2+2?"` — output on stdout, no behaviour difference from v0.2.0
- **`--stream-json` JSON only:** `./claude-print --stream-json "What is 2+2?" 2>/dev/null` — only JSON lines on stdout
- **`--stream-json` display only:** `./claude-print --stream-json "What is 2+2?" 1>/dev/null` — only visual progress on stderr
- **stdin prompt:** `echo "What is 2+2?" | ./claude-print` — prompt read from pipe, normal display output
- **stdin + stream-json together:** `echo "What is 2+2?" | ./claude-print --stream-json 2>/dev/null` — JSON on stdout


## [0.2.0] - 2026-02-19

### Changed

- Align verbose mode display with normal mode for consistent output formatting
- Add blank-line separator between consecutive tool call headers for improved readability

## [0.1.0] - 2025-01-01

### Added

- Initial release
