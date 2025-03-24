#!/usr/bin/env bash
set -euo pipefail

if [ -n "$(git status --porcelain --untracked-files=no)" ]; then
  printf "\nError: repo has uncommitted changes\n\n"
  exit 1
fi

GIT_TAG_LATEST=$(git tag --sort=version:refname | tail -n 1)

if [ -z "$GIT_TAG_LATEST" ]; then
  GIT_TAG_LATEST="v0.0.0"
fi

# Strip prefix 'v' from the tag to easily increment
GIT_TAG_LATEST=$(echo "$GIT_TAG_LATEST" | sed 's/^v//')

VERSION_TYPE="${1-}"
VERSION_NEXT=""

if [ "$VERSION_TYPE" = "patch" ]; then
  # Increment patch version
  VERSION_NEXT="$(echo "$GIT_TAG_LATEST" | awk -F. '{$NF++; print $1"."$2"."$NF}')"
elif [ "$VERSION_TYPE" = "minor" ]; then
  # Increment minor version
  VERSION_NEXT="$(echo "$GIT_TAG_LATEST" | awk -F. '{$2++; $3=0; print $1"."$2"."$3}')"
elif [ "$VERSION_TYPE" = "major" ]; then
  # Increment major version
  VERSION_NEXT="$(echo "$GIT_TAG_LATEST" | awk -F. '{$1++; $2=0; $3=0; print $1"."$2"."$3}')"
else
  # Print error for unknown versioning type
  printf "\nError: invalid VERSION_TYPE arg passed, must be 'patch', 'minor' or 'major'\n\n"
  # Exit with error code
  exit 1
fi

git tag -a "v$VERSION_NEXT" -m "Release: v$VERSION_NEXT"

git push origin main --follow-tags
