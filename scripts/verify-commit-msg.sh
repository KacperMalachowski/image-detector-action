#!/usr/bin/env bash

set -euo pipefail

MSG_FILE="${1:-.git/COMMIT_EDITMSG}"
SUBJECT="$(head -n1 "$MSG_FILE" | tr -d '\n')"

echo "Commit subject: '$SUBJECT'"

if [[ "$SUBJECT" =~ ^Merge ]]; then
    echo "Merge commit detected, skipping verification."
    exit 0
fi

# Conventional Commits v1.0.0 subject line pattern:
# type(scope)!: short summary
# where type ∈ {build,chore,ci,docs,feat,fix,perf,refactor,revert,style,test}
# scope is optional, "!" marks a breaking change, summary must exist, keep subject ≲ 72 chars.
re='^(build|chore|ci|docs|feat|fix|perf|refactor|revert|style|test)(\([a-z0-9._-]+\))?(!)?: [^[:space:]].{0,72}$'

if [[ ! "$SUBJECT" =~ $re ]]; then
    echo "✗ Commit message does not follow Conventional Commits v1.0.0 format."
    echo "  See https://www.conventionalcommits.org/en/v1.0.0/"
    echo "  Got:    $SUBJECT"
    echo "  Expect: type(scope)!: short summary"
    echo "  Types: build|chore|ci|docs|feat|fix|perf|refactor|revert|style|test"
    exit 1
fi

# Warn if body has 'BREAKING CHANGE: ' but subject does not have '!'
if grep -qiE '^breaking change:' "$MSG_FILE" && [[ ! "$SUBJECT" =~ !: ]]; then
    echo "⚠️  Commit body contains 'BREAKING CHANGE:' but subject line does not indicate a breaking change with '!'."
    echo "  Consider adding '!' before the colon in the subject line."
fi

echo "✓ Commit message follows Conventional Commits v1.0.0 format."
