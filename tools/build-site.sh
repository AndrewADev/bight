#!/usr/bin/env sh
# Assembles the GitHub Pages site into the given directory.
#
# Layout:
#   index.md               README
#   commands/*.md          generated command reference
#   CHANGELOG.md, DEVELOPING.md
#   *.html                 rendered sibling of every .md, with .md links pointing at .html
#   llms.txt               curated index for LLM agents (docs/llms.txt)
#   llms-full.txt          README plus every command page, concatenated
set -eu

out="${1:?usage: build-site.sh <out-dir>}"
root="$(cd "$(dirname "$0")/.." && pwd)"
tools="$root/tools"

rm -rf "$out"
mkdir -p "$out/commands"

# Repo-relative links in the README become site-relative.
sed 's#docs/commands/#commands/#g' "$root/README.md" > "$out/index.md"
cp "$root/CHANGELOG.md" "$root/DEVELOPING.md" "$root/docs/llms.txt" "$out/"
cp "$root/docs/commands/"*.md "$out/commands/"

{
  cat "$out/index.md"
  for page in "$out/commands/"*.md; do
    printf '\n\n---\n\n'
    cat "$page"
  done
} > "$out/llms-full.txt"

find "$out" -name '*.md' | while read -r page; do
  title="$(grep -m1 '^#' "$page" | sed 's/^#* *//')"
  pandoc --standalone --lua-filter "$tools/site-links.lua" \
    --metadata "pagetitle=${title:-bight}" \
    "$page" -o "${page%.md}.html"
done
