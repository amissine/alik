#!/usr/local/bin/bash
# feed.sh sources {{{1

. util/common.sh

# If called with arguments (feed, trading_pair), get the latest trades {{{1
if [ $# -gt 0 ]; then
  case $1 in
    "bitfinex")
      curl "https://api.bitfinex.com/v1/trades/$2?limit_trades=2"
      ;;
    "coinbase")
      curl "https://api.pro.coinbase.com/products/$2/trades?limit=2"
      ;;
    *)
      log TODO implement feed $1 # TODO implement feed kraken
      ;;
  esac
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

# A separate process is started for each asset being traded on SDEX for XLM. {{{1
# The process monitors the order book for the asset by calling curl. A call returns
# one or more order book updates. Presently, an order book consists of one ask and
# one bid (limit=1 below, local asset).
#
# Each order book update consists of one line (--line-buffered above) and triggers
# another curl call that returns some latest trades of the asset for XLM. Presently,
# two latest trades are being piped to the feed (limit=2 above, bat suffix bats).
# Then we pipe the order book update (echo "$REPLY" below) to the feed.
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
