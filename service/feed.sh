#!/usr/local/bin/bash

. util/common.sh

#touch ./sysin; chgrp admin ./sysin; chmod 660 ./sysin;
touch ./syserr; chgrp admin ./syserr; chmod 640 ./syserr;

sdex_ob () { # {{{1
  local ASSET=$1
  local URL='https://horizon.stellar.org/order_book'
  local BUY='buying_asset_type=native'
  . util/$ASSET.sh

  log "sdex_ob: $ASSET started" >>./syserr
  while true; do
    {
      let rc=0
      while [ $rc -eq 0 ]; do
        curl -H "Accept: text/event-stream" "$URL?$BUY&$SELL&limit=$DEPTH" \
          --silent --no-buffer | grep --line-buffered --only-matching '{.*}$'
        rc=$?
      done
      log "curl: $ASSET rc $rc" >>./syserr
    } | ./feed $ASSET  2>>./syserr
  done
} 

for FEED in $FEEDS; do
  for TRADING_PAIR in $TRADING_PAIRS; do
    if [ "$FEED" = "sdex" ]; then
      if [ "${TRADING_PAIR:0:3}" = "XLM" ]; then
        sdex_ob ${TRADING_PAIR:3:6} &
      fi
    else
      if [ "${TRADING_PAIR:3:6}" = "USD" ]; then
        log "feed: skip $FEED/$TRADING_PAIR" >>./syserr
      fi
    fi
  done
done
wait
log "feed: exiting..." >>./syserr
