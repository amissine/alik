#!/usr/local/bin/bash

set -e

echo '=== get XRP order book and trades ===' # {{{1
. util/XRP.sh
for i in 1 2; do
  sdex_ob XRP # || break
done | {
  while true; do
    read || break
    for t in sdex_t bitfinex_t; do $t XRP; done
    echo "$REPLY"
  done
}
echo "=== exiting with $? ==="
