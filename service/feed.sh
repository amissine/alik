#!/usr/local/bin/bash
# feed.sh sources {{{1

. util/common.sh

# See also:
#   https://docs.google.com/document/d/1h5P9SaulgMFryKERavy7s5dIiI7incNpcCiUxKBPQaI

sdex () { # {{{1
  local ASSET=$1
  . util/$ASSET.sh # setting ai env var
  local asset="selling_asset_code=$ASSET&selling_asset_issuer=$ai&limit=1"
  local bat="$batp$ASSET&base_asset_issuer=$ai$bats"

  # See also:
  # - https://www.stellar.org/developers/horizon/reference/resources/orderbook.html
  #
  while true; do
    curl -H "$ch" "$url/order_book?$bs&$asset" $cs | grep $gopts '{.*}$' || break
  done | {
    while true; do
      read || break
      curl -H "$ch" "$url/trades?$bat" $cs | grep $gopts '{.*}$'
      echo "$REPLY"
    done
  } | ./feed $ASSET 2>>./syserr
  log "sdex exiting with $?..."
} 

# Set traps, start sdex processes, and wait {{{1
onSIGCONT () { # {{{2
  log 'received SIGCONT, killing feeds'
  cat ./syserr | grep "$SDEX_FEED_STARTED" | kill $(awk '{print $3}')
} # }}}2

# sudo svc -d /service/feed ==> SIGTERM, SIGCONT
trap "{ log 'received SIGTERM'; }" SIGTERM
trap onSIGCONT SIGCONT

for TRADING_PAIR in $TRADING_PAIRS_SDEX; do
  sdex ${TRADING_PAIR:0:3} &
done

wait
