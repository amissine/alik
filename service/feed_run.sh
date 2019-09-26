#!/usr/local/bin/bash
echo 'bitfinex bitstamp gemini kraken sdex' > env/FEEDS
echo 'BTCUSD CNYUSD ETHUSD XLMBTC XLMCNY XLMUSD XLMXRP XRPUSD' > env/TRADING_PAIRS
#echo 'BTCUSD CNYUSD ETHUSD XLMXRP' > env/TRADING_PAIRS

. util/common.sh

touch ./syserr; chgrp admin ./syserr; chmod 640 ./syserr;

log started
exec envdir ./env ./feed.sh
log "exiting with $?..."
