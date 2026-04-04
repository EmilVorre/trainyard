#!/usr/bin/env bash
# Creates the standard Trainyard labels on the GitHub repo.
# Requires: gh CLI authenticated (gh auth login)
# Usage: ./scripts/setup-labels.sh

set -euo pipefail

REPO="Emilvorre/trainyard"

create_label() {
  local name="$1"
  local color="$2"
  local description="$3"

  if gh label list --repo "$REPO" --json name | grep -q "\"$name\""; then
    echo "  (exists) $name"
  else
    gh label create "$name" \
      --repo "$REPO" \
      --color "$color" \
      --description "$description"
    echo "  created: $name"
  fi
}

echo "Setting up labels for $REPO..."
echo

# Trigger label
create_label "preview"     "0075ca" "Deploy a preview environment for this PR"

# Standard issue labels
create_label "bug"         "d73a4a" "Something isn't working"
create_label "enhancement" "a2eeef" "New feature or improvement"
create_label "question"    "d876e3" "Further information is requested"
create_label "docs"        "0075ca" "Documentation change"
create_label "chore"       "e4e669" "Maintenance, dependency updates"

# Triage labels
create_label "good first issue" "7057ff" "Good for newcomers"
create_label "help wanted"      "008672" "Extra attention is needed"
create_label "wontfix"          "ffffff" "This will not be worked on"
create_label "duplicate"        "cfd3d7" "This issue or PR already exists"
create_label "invalid"          "e4e669" "This doesn't seem right"

# Component labels
create_label "cli"         "bfd4f2" "Affects the yard CLI"
create_label "helm"        "bfd4f2" "Affects the Helm charts"
create_label "workflows"   "bfd4f2" "Affects GitHub Actions workflows"

echo
echo "Done."
