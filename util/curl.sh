#!/usr/local/bin/bash
# curl.sh sources {{{1

. util/common.sh

# Locals for sdex_ob {{{1
ch='Accept: text/event-stream' # curl header
url='https://horizon.stellar.org'
bs='buying_asset_type=native&selling_asset_type=credit_alphanum4'
batp='base_asset_type=credit_alphanum4&base_asset_code=' # bat (see below) prefix
bats='&counter_asset_type=native&limit=2&order=desc' # bat suffix
cs='--silent --no-buffer' # curl suffix
gopts='--line-buffered --only-matching' # grep opts

sdex_ob () { # {{{1
  local ASSET=$1
  local asset="selling_asset_code=$ASSET&selling_asset_issuer=$ai&limit=1"

  # See also:
  # - https://www.stellar.org/developers/horizon/reference/resources/orderbook.html
  #
  curl -H "$ch" "$url/order_book?$bs&$asset" $cs | grep $gopts '{.*}$'
} 

sdex_t () { # {{{1
  local ASSET=$1
  local bat="$batp$ASSET&base_asset_issuer=$ai$bats"
  
  curl -H "$ch" "$url/trades?$bat" $cs | grep $gopts '{.*}$'
}
