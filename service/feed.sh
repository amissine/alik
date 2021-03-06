#!/usr/local/bin/bash
# feed.sh sources {{{1

. util/common.sh

# See also:
#   https://docs.google.com/document/d/1h5P9SaulgMFryKERavy7s5dIiI7incNpcCiUxKBPQaI

#set -e # TODO remove when done debugging

# Locals {{{1
SDEX_STARTED="sdex started for"

sdex () { # {{{1
  local asset
  local ASSET=$1
  local q
  . util/$ASSET.sh

  log "$SDEX_STARTED $ASSET"
  while true; do
    sdex_ob $ASSET || break
  done | {
    while true; do
      read; q="$REPLY"
      for t in sdex_t bitfinex_t coinbase_t kraken_t; do
        asset=$ASSET
        [ "$t" != 'sdex_t' -a "$asset" = 'CNY' ] && continue
        [ "$t" != 'sdex_t' -a "$asset" = 'SLT' ] && continue
        if [ "$t" = 'kraken_t' -a "$asset" = 'BTC' ]; then asset='XBT'
        elif [ "$t" != 'sdex_t' -a "$asset" = 'USD' ]; then asset='XLM'; fi
        $t $asset
      done
      echo "$q"
    done
  } | ./feed $ASSET 2>>./syserr
  log "sdex $ASSET exiting with $?"
} 

# Set traps, start sdex processes, and wait {{{1
onSIGCONT () { # {{{2
  log 'received SIGCONT, killing feeds'
  ps -ef | grep "/usr/local/bin/bash ./feed.sh" | kill $(awk '{print $2}')
} # }}}2

# sudo svc -d /service/feed ==> SIGTERM, SIGCONT
trap "{ log 'received SIGTERM'; }" SIGTERM
trap onSIGCONT SIGCONT

for TRADING_PAIR in $TRADING_PAIRS_SDEX; do
  sdex ${TRADING_PAIR:0:3} &
  log "sdex ${TRADING_PAIR:0:3} pid $! started"
done

wait
