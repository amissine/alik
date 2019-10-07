#!/usr/local/bin/bash
# feed.sh sources {{{1

. util/common.sh

# If called with arguments (feed, trading_pair), get the latest trades {{{1
if [ $# -gt 0 ]; then
  log $@
  echo '{"Result":"Ok"}'
  exit 0
fi

# Locals for sdex {{{1
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

  log sdex $ASSET started
  while true; do
    curl -H "$ch" "$url/order_book?$bs&$asset" $cs | grep $gopts '{.*}$' || break
  done | {
    while true; do
      read || break
      curl -H "$ch" "$url/trades?$bat" $cs | grep $gopts '{.*}$'
      echo "$REPLY"
    done
  } | ./feed 'sdex' $ASSET "$FEEDS" "$TRADING_PAIRS" 2>>./syserr
  log "sdex exiting with $?..."
} 

# Set traps, start sdex processes, and wait {{{1
onSIGCONT () { # {{{2
  log onSIGCONT received SIGCONT
  cat ./syserr | grep 'sdex feed started' | kill $(awk '{print $3}')
} # }}}2

trap "{ log received SIGTERM; }" SIGTERM
trap onSIGCONT SIGCONT

for TRADING_PAIR in $TRADING_PAIRS_SDEX; do
  sdex ${TRADING_PAIR:0:3} &
done

wait
