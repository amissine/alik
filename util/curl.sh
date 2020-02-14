#!/usr/local/bin/bash
# curl.sh sources {{{1

. util/common.sh

# Locals {{{1
ch='Accept: text/event-stream' # curl header
url='https://horizon.stellar.org'
bs='buying_asset_type=native&selling_asset_type=credit_alphanum4&cursor=now'
batp='base_asset_type=credit_alphanum4&base_asset_code=' # bat (see below) prefix
bats='&counter_asset_type=native&limit=1&order=desc' # bat suffix
cs='--silent --no-buffer' # curl suffix
gopts='--line-buffered --only-matching' # grep opts
gfilter='\"counter_asset_type\":\"native\"' # preventing trades with

# A separate process is started for each asset traded for XLM on SDEX. {{{1
# The process monitors the order book for the asset by calling curl. A call returns
# one or more order book updates. Presently, an order book consists of one ask and
# one bid (limit=1 below, local asset).
#
# Each order book update consists of one line (--line-buffered above) and triggers
# another curl call that returns some latest trades of the asset for XLM in the
# descending order. Presently, ONE latest trades are being piped to the feed, the 
# latest one first (limit=2&order=desc above, bat suffix bats). The trades are being
# piped out only if they differ from the previous bunch of trades. Then we pipe the 
# order book update (echo "$REPLY") to the feed.

sdex_ob () { # {{{1
  local ASSET=$1
  local asset="selling_asset_code=$ASSET&selling_asset_issuer=$ai&limit=1"
  local URL="$url/order_book?$bs&$asset"
  local rc

  # See also:
  # - https://www.stellar.org/developers/horizon/reference/resources/orderbook.html
  #
  curl -H "$ch" "$URL" $cs | grep $gopts '{.*}$'; rc=$?
  [ $rc -eq 0 ] || echo "{\"internal_error\":\"sdex_ob $ASSET rc $rc\"}"
} 

sdex_t () { # {{{1
  local ASSET=$1
  local bat="$batp$ASSET&base_asset_issuer=$ai$bats"
  local URL="$url/trades?$bat"
  local q
  
  q=$(curl -H "$ch" "$URL" $cs | grep $gopts '{.*}$')
  [ "$q" = "$sdex_t_p" ] || { sdex_t_p="$q"; echo "$q"; }
}

bitfinex_t () { # {{{1
  local url="https://api-pub.bitfinex.com/v2/trades/t$1USD/hist"
  local start
  local json

  # See also: {{{2
  # - https://docs.bitfinex.com/reference#rest-public-trades
  # - https://gist.github.com/magnetikonline/90d6fe30fc247ef110a1
  #
  bitfinex_t_rate_ok || return 0

  # Call curl, set json and  next start. {{{2
  if [ "$bitfinex_t_data" ]; then
    start=$bitfinex_t_data # milliseconds
    bitfinex_t_data=$(curl "$url?start=$start" $cs)
  else
    bitfinex_t_data=$(curl "$url" $cs)
  fi
  json=$bitfinex_t_data
  bitfinex_t_data=${bitfinex_t_data#*\,}
  bitfinex_t_data=${bitfinex_t_data%%\,*} # next start, milliseconds

  # Remove duplicate trades. {{{2
  #echo "json              '$json'"
  #echo "bitfinex_t_json_p '$bitfinex_t_json_p'"
  if [ "$bitfinex_t_json_p" -a "$json" = "$bitfinex_t_json_p" ]; then return 0; fi
  if [ "$start" ]; then
    #echo $start
    #echo $bitfinex_t_data
    if [ "$start" = "$bitfinex_t_data" ]; then
      bitfinex_t_json_p="$json"
      return 0
    fi
    json=${json%%$start\,*}
    json=${json%\,\[*}
    json="${json}]" # duplicate trades removed
    bitfinex_t_json_p="$json"
  fi
  echo $json # }}}2
}

bitfinex_t_rate_ok () { # {{{1
  # Ratelimit: 30 req/min
  if [ "$bitfinex_t_data" ]; then
    let duration=$SECONDS-$bitfinex_t_p
    if [ $duration -ge 2 ]; then
      bitfinex_t_p=$SECONDS
      return 0
    else
      return 1
    fi
  else
    bitfinex_t_p=$SECONDS
    return 0
  fi
}

coinbase_t () { # {{{1
  local url="https://api.pro.coinbase.com/products/$1-USD/trades"
  local before
  local json
  local headers

  # See also: {{{2
  # - https://docs.pro.coinbase.com/#get-trades
  # - https://docs.pro.coinbase.com/#pagination
  #
  coinbase_t_rate_ok || return 0

  # Call curl, set json and next before. {{{2
  if [ "$coinbase_t_data" ]; then
    before=$coinbase_t_data # retrieving trades that are newer than before
    coinbase_t_data="$(curl -i "$url?before=$before" $cs)"
    #echo "- ? $?; coinbase_t_data '$coinbase_t_data'"
  else
    coinbase_t_data="$(curl -i "$url" $cs)"
  fi
  headers="${coinbase_t_data%\[*\]}"
  let json_length=${#coinbase_t_data}-${#headers}
  if [ $json_length -eq 2 ]; then # no new trades
    coinbase_t_data=$before
    #echo "- json_length $json_length; coinbase_t_data '$coinbase_t_data'"
    return 0
  fi
  json="${coinbase_t_data:${#headers}:$json_length}"
  echo $json
  coinbase_t_data=$(echo "$headers" | grep 'cb-before: ')
  coinbase_t_data=${coinbase_t_data#cb-before: }
  let coinbase_t_data_length=${#coinbase_t_data}-1
  coinbase_t_data=${coinbase_t_data:0:$coinbase_t_data_length} # }}}2
}
coinbase_t_rate_ok () { # {{{1
  # We throttle public endpoints by IP: 3 requests per second, 
  # up to 6 requests per second in bursts.
  if [ "$coinbase_t_data" ]; then
    let duration=$SECONDS-$coinbase_t_p
    if [ $duration -ge 1 ]; then
      coinbase_t_p=$SECONDS
      return 0
    else
      return 1
    fi
  else
    coinbase_t_p=$SECONDS
    return 0
  fi
}

kraken_t () { # {{{1
  local url="https://api.kraken.com/0/public/Trades?pair=$1USD"
  local last
  local json

  # See also: {{{2
  # - https://support.kraken.com/hc/en-us/articles/360000919986-Public-endpoint-examples-you-can-try-them-directly-in-a-web-browser-
  #
  kraken_t_rate_ok || return 0

  # Call curl, set json and next last. {{{2
  if [ "$kraken_t_data" ]; then
    last=$kraken_t_data # retrieving trades since the last one
    #echo "- last          '$last'"
    kraken_t_data=$(curl "$url&since=$last" $cs)
  else
    kraken_t_data=$(curl "$url" $cs)
  fi
  json=$kraken_t_data
  kraken_t_data=${kraken_t_data##*\"last\"\:\"}
  kraken_t_data=${kraken_t_data%%\"*}
  #echo "- kraken_t_data '$kraken_t_data'"
  if [ "$kraken_t_data" = "$last" ]; then # no new trades
    return 0
  fi
  echo $json
}

kraken_t_rate_ok () { # {{{1
  if [ "$_krakent_data" ]; then
    let duration=$SECONDS-$kraken_t_p
    if [ $duration -ge 1 ]; then
      kraken_t_p=$SECONDS
      return 0
    else
      return 1
    fi
  else
    kraken_t_p=$SECONDS
    return 0
  fi
}
