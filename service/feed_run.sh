#!/usr/local/bin/bash
echo 'bitfinex coinbase kraken' > env/FEEDS
echo 'BTCUSD ETHUSD XLMUSD XRPUSD' > env/TRADING_PAIRS
echo 'BTCXLM CNYXLM ETHXLM SLTXLM USDXLM XRPXLM' > env/TRADING_PAIRS_SDEX
echo 'sdex feed started' > env/SDEX_FEED_STARTED
echo '10' > env/FEED_HISTORY_LIMIT

. util/common.sh

touch ./syserr; chgrp admin ./syserr; chmod 640 ./syserr;

log started
exec envdir ./env ./feed.sh
log "exiting with $?..."
