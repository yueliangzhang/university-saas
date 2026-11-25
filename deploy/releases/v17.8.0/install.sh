#!/usr/bin/env bash
set -euo pipefail

if [[ $# != 1 ]]; then
  printf 'Usage: %s <V17.8.0 commit or tag>\n' "$0" >&2
  exit 2
fi

root=/srv/gva
export PATH="$HOME/.local/opt/go1.26.11/bin:$HOME/.local/opt/node24/bin:/usr/local/go/bin:$PATH"
source_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
cd "$root/repo"
git fetch origin feature/university-saas --tags
commit=$(git rev-parse --verify --end-of-options "$1^{commit}")
git merge-base --is-ancestor "$commit" FETCH_HEAD
previous=$(readlink -f "$root/current")
test -x "$previous/server/gva-server"
release=$(mktemp -d "$root/releases/${commit:0:8}.XXXXXX")
chmod 755 "$release"
git archive "$commit" | tar -x -C "$release"

cd "$release/server"
go build -o gva-server .
ln -sfnT /var/lib/gva/uploads uploads
cd "$release/web"
if cmp -s package.json "$previous/web/package.json" &&
    test -x "$previous/web/node_modules/.bin/vite"; then
  cp -a "$previous/web/node_modules" .
else
  npm install --prefer-offline --no-audit
fi
VITE_APP_NAME="高校管理平台" npm run build
test "$(readlink -f "$root/current")" = "$previous"
printf '%s\n' "$previous" > "$release/previous-release"
bash "$source_dir/rollback.sh" "$release"
