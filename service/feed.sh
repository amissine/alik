#!/usr/local/bin/bash

. util/common.sh

touch ./syserr; chgrp admin ./syserr; chmod 640 ./syserr;

# Locals {{{1
ch='Accept: text/event-stream' # curl header
url='https://horizon.stellar.org'
bs='buying_asset_type=native&selling_asset_type=credit_alphanum4'
cs='--silent --no-buffer' # curl suffix
gopts='--line-buffered --only-matching' # grep opts

sdex () { # {{{1
  local ASSET=$1
  . util/$ASSET.sh # setting ai env var
  local asset="selling_asset_code=$ASSET&selling_asset_issuer=$ai&limit=1"
  local bat="base_asset_type=credit_alphanum4&base_asset_code=$ASSET&base_asset_issuer=$ai&counter_asset_type=native&limit=2&order=desc"

  log $BASHPID "sdex: $ASSET started" >>./syserr
  while true; do
    curl -H "$ch" "$url/order_book?$bs&$asset" $cs | grep $gopts '{.*}$' || break
  done | {
    while true; do
      read || break
      curl -H "$ch" "$url/trades?$bat" $cs | grep $gopts '{.*}$'
      echo $REPLY
    done
  } | ./feed 'sdex' $ASSET 2>>./syserr
  log $BASHPID "sdex exiting with $?..." >>./syserr
} 

for FEED in $FEEDS; do # {{{1
  for TRADING_PAIR in $TRADING_PAIRS; do
    if [ "$FEED" = "sdex" ]; then
      if [ "${TRADING_PAIR:0:3}" = "XLM" ]; then
        sdex ${TRADING_PAIR:3:6} &
      fi
    else
      if [ "${TRADING_PAIR:3:6}" = "USD" ]; then
        log "feed: skip $FEED/$TRADING_PAIR" >>./syserr
      fi
    fi
  done
done
wait
log $BASHPID "feed exiting with $?..." >>./syserr
