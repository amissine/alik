#!/usr/local/bin/bash
echo 'bitfinex bitstamp gemini kraken sdex' > env/FEEDS
#echo 'BTCUSD CNYUSD ETHUSD XLMBTC XLMCNY XLMUSD XLMXRP XRPUSD' > env/TRADING_PAIRS
echo 'BTCUSD CNYUSD ETHUSD XLMCNY' > env/TRADING_PAIRS
exec envdir ./env ./feed.sh
