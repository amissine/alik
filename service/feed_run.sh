#!/bin/sh
echo 'bitfinex bitstamp gemini kraken sdex' > env/FEEDS
echo 'BTCUSD ETHUSD XLMBTC XLMUSD XLMXRP XRPUSD' > env/TRADING_PAIRS
exec envdir ./env ./feed.sh
