#!/usr/local/bin/bash
echo 'bitfinex coinbase kraken' > env/FEEDS
echo 'BTCUSD ETHUSD XLMUSD XRPUSD' > env/TRADING_PAIRS
echo 'BTCXLM CNYXLM ETHXLM SLTXLM USDXLM XRPXLM' > env/TRADING_PAIRS_SDEX

. util/common.sh

touch ./syserr; chgrp admin ./syserr; chmod 640 ./syserr;

log started
exec envdir ./env ./feed.sh
log "exiting with $?..."
