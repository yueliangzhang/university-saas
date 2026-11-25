#!/usr/bin/env bash
set -euo pipefail

root=/srv/gva
current=$(readlink -f "$root/current")
if [[ $# -gt 1 ]]; then
  printf 'Usage: %s [release directory]\n' "$0" >&2
  exit 2
fi
target=$(readlink -f "${1:-$(cat "$current/previous-release")}")
if [[ "$target" != "$root/releases/"* ]] ||
    [[ ! -x "$target/server/gva-server" || ! -s "$target/web/dist/index.html" ]]; then
  printf 'Invalid release: %s\n' "$target" >&2
  exit 1
fi
sudo -v

switch_release() {
  ln -sfnT "$1" "$root/.current-$$" &&
    mv -Tf "$root/.current-$$" "$root/current"
}

verify() {
  sudo -n systemctl restart gva.service &&
    [[ $(curl -fsS --max-time 5 --retry 8 --retry-delay 1 --retry-connrefused http://127.0.0.1:8888/health) == '"ok"' ]] &&
    [[ $(curl -fkSs --max-time 5 --noproxy '*' --connect-to 43.128.51.87:443:127.0.0.1:443 https://43.128.51.87/api/health) == '"ok"' ]] &&
    curl -fkSs --max-time 5 --noproxy '*' --connect-to 43.128.51.87:443:127.0.0.1:443 https://43.128.51.87/ |
      cmp -s - "$target/web/dist/index.html"
}

switch_release "$target"
if verify; then
  printf 'Current release: %s\nPrevious release: %s\n' "$target" "$current"
else
  printf 'Health check failed for %s. Restoring %s.\n' "$target" "$current" >&2
  switch_release "$current"
  sudo -n systemctl restart gva.service
  exit 1
fi
