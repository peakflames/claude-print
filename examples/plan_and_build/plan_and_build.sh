#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROXY="${SCRIPT_DIR}/../../claude-print"

# Use .exe on Windows (MINGW/MSYS)
case "$(uname -s)" in
  MINGW*|MSYS*|CYGWIN*) PROXY="${PROXY}.exe" ;;
esac

usage() {
  cat <<EOF
Usage: $(basename "$0") <task-description>

Two-phase headless workflow using claude-print:
  1. Planning  — Claude creates a detailed execution plan
  2. Building  — Claude executes the plan

Options:
  -h, --help    Show this help message

Examples:
  $(basename "$0") "Create a hello world HTML page"
  $(basename "$0") "Build a REST API with Go that serves TODO items"
  $(basename "$0") "Refactor the auth module to use JWT tokens"
EOF
  exit 0
}

# Show usage if no args or help flag
[[ $# -lt 1 || "$1" == "-h" || "$1" == "--help" ]] && usage

TASK="$1"

# Step 1: Generate plan filename via Claude Haiku
echo ""
echo ""..
echo "Generating plan filename..."
PLAN_FILE=$(claude -p "Suggest a short kebab-case filename (no extension) that ends with the suffix '-plan' for a implementation plan about: '${TASK}'. Reply with ONLY the filename, nothing else." --model claude-haiku-4-5-20251001)
PLAN_FILE="${SCRIPT_DIR}/${PLAN_FILE}.md"
echo "Plan will be saved to: ${PLAN_FILE}"


# Step 2: Planning phase
echo ""
echo ""
echo ""
echo "=== Planning Phase ==="
$PROXY "<operating_env>You are operating in a non-interactive, headless CLI script. The user expects fully automous, non-interactive operation. The outputs generate during this script will be used as inputs to other scripts. The working directory is ${SCRIPT_DIR}.</operating_env> \
        <instruction>Create a detailed execution plan that will handed off to development team of mix experience levels for the task: '${TASK}'.</instruction> \
        <critical_endstate>**CRITICAL INSTRUCTION:** Immediately copy the generated plan file to ${PLAN_FILE} and ensure the copied plan exists at the target path.</critical_endstate>" \
  --permission-mode plan \
  --allowedTools "Read,Glob,Grep,WebFetch,WebSearch,Task,Write,Bash" \
  --model opus

echo ""
echo "Plan created: ${PLAN_FILE}"


# Step 3: Build phase
echo ""
echo ""
echo ""
echo "=== Build Phase ==="
$PROXY "<operating_env>You are operating in a non-interactive, headless CLI script. The user expects fully automous, non-interactive operation. The outputs generate during this script will be used as inputs to other scripts.</operating_env> \
        <instruction>Read ${PLAN_FILE} and execute the plan. Use Task agents for parallel work where possible.</instruction> " \
  --dangerously-skip-permissions \
  --model sonnet

echo ""
echo "Done!"
