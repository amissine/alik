#!/usr/local/bin/bash
# feed.sh sources {{{1

. util/common.sh

# Locals {{{1
ch='Accept: text/event-stream' # curl header
url='https://horizon.stellar.org'
bs='buying_asset_type=native&selling_asset_type=credit_alphanum4'
batp='base_asset_type=credit_alphanum4&base_asset_code=' # bat (see below) prefix
bats='&counter_asset_type=native&limit=2&order=desc' # bat suffix
cs='--silent --no-buffer' # curl suffix
gopts='--line-buffered --only-matching' # grep opts

sdex () { # {{{1
  local ASSET=$1
  . util/$ASSET.sh # setting ai env var
  local asset="selling_asset_code=$ASSET&selling_asset_issuer=$ai&limit=1"
  local bat="$batp$ASSET&base_asset_issuer=$ai$bats"

  log "sdex: $ASSET started"
  while true; do
    curl -H "$ch" "$url/order_book?$bs&$asset" $cs | grep $gopts '{.*}$' || break
  done | {
    while true; do
      read || break
      curl -H "$ch" "$url/trades?$bat" $cs | grep $gopts '{.*}$'
      echo "$REPLY"
    done
  } | ./feed 'sdex' $ASSET 2>>./syserr
  log "sdex exiting with $?..."
} 

for FEED in $FEEDS; do # {{{1
  for TRADING_PAIR in $TRADING_PAIRS; do
    if [ "$FEED" = "sdex" ]; then
      if [ "${TRADING_PAIR:0:3}" = "XLM" ]; then
        sdex ${TRADING_PAIR:3:6} &
      fi
    else
      if [ "${TRADING_PAIR:3:6}" = "USD" ]; then
        log "feed: skip $FEED/$TRADING_PAIR"
      fi
    fi
  done
done

# Set traps and wait {{{1
kill_pids () { # {{{2
  pids=' '
  while true; do
    shift 2; if [ ! $? -eq 0 ]; then break; fi
    pids="$pids $1"; shift 4
  done
  log "kill_pids: killing pids$pids..."
  kill $pids
}
onSIGCONT () { # {{{2
  log received SIGCONT
  cat ./syserr | grep 'feed started' | kill_pids $(cat)
} # }}}2

trap "{ log received SIGTERM; }" SIGTERM
trap onSIGCONT SIGCONT

wait
