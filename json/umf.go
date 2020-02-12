package json

// import {{{1
import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"
)

var w = bufio.NewWriterSize(os.Stdout, 65536) // {{{1
var enc = json.NewEncoder(w)

func encodeUMF(q *UMF) {
	if err := enc.Encode(q); err != nil {
		log.Fatal(os.Getpid(), "encodeUMF", err)
	}
}

type UMF struct { // Unified Market Feed {{{1
	AE   AE
	Feed []interface{} // either [asks, bids] or tradeAsArray or tradeBitfinex...
	UTC  time.Time
}

type AE struct { // Asset, Exchange {{{1
	Asset, Exchange string
}

type OrderBook []struct { // {{{1
	Amount, Price string
	Price_r       struct{ D, N float64 }
}

func tradeBitfinex(a string, w []interface{}) []interface{} { // {{{1
	t := make([]interface{}, 3)
	mts := int64(w[1].(float64)) // millisecond time stamp
	sec := mts / 1000
	nsec := mts % 1000 * 1000000
	loc, _ := time.LoadLocation("UTC")
	t[0] = time.Unix(sec, nsec).In(loc)
	t[1] = w[2].(float64)
	t[2] = w[3].(float64) // price in USD
	return t
}

func tradeCoinbase(a string, w map[string]interface{}) []interface{} { // {{{1
	t := make([]interface{}, 3)
	size := parse(w["size"].(string))
	loc, _ := time.LoadLocation("UTC")
	t[0], _ = time.ParseInLocation(time.RFC3339, w["time"].(string), loc)
	if w["side"].(string) == "sell" {
		t[1] = size
	} else {
		t[1] = -size
	}
	t[2] = parse(w["price"].(string)) // price in USD
	return t
}

func tradeKraken(a string, w []interface{}) []interface{} { // {{{1
	t := make([]interface{}, 3)
	ts := w[2].(float64)
	sec := int64(ts)
	nsec := int64(1000000000 * (ts - float64(sec)))
	loc, _ := time.LoadLocation("UTC")
	t[0] = time.Unix(sec, nsec).In(loc)
	amount := parse(w[1].(string))
	if w[3].(string) == "s" {
		amount = -amount
	}
	t[1] = amount
	t[2] = parse(w[0].(string)) // price in USD
	return t
}
func parse(s string) float64 { // {{{1
	f, e := strconv.ParseFloat(s, 64)
	if e != nil {
		log.Println("parse ERROR", e)
	}
	return f
}

func newUMF(a, e string, f []interface{}) *UMF { // {{{1
	location, _ := time.LoadLocation("UTC")
	return &UMF{
		AE:   AE{Asset: a, Exchange: e},
		Feed: f,
		UTC:  time.Now().In(location),
	}
}

// Presently, tradeAsArray returns: {{{1
//
// [ ledger_close_time, base_amount, priceInXLM ]
//
// where base_amount will be positive when the trade is a buy, and negative when
// the trade is a sell. }}}1
func tradeAsArray(v *map[string]interface{}) []interface{} { // {{{1
	w := *v
	price := w["price"].(map[string]interface{})
	base_amount := parse(w["base_amount"].(string))
	a := make([]interface{}, 3)
	priceInXLM := price["n"].(float64) / price["d"].(float64)
	loc, _ := time.LoadLocation("UTC")
	a[0], _ = time.ParseInLocation(time.RFC3339, w["ledger_close_time"].(string), loc)
	if w["base_is_seller"].(bool) {
		a[1] = base_amount
	} else {
		a[1] = -base_amount
	}
	a[2] = priceInXLM
	return a
}

func ob(b []interface{}) OrderBook { // {{{1
	c := make(OrderBook, len(b))
	for i, v := range c {
		v.Amount = b[i].(map[string]interface{})["amount"].(string)
		v.Price = b[i].(map[string]interface{})["price"].(string)
		v.Price_r.D = b[i].(map[string]interface{})["price_r"].(map[string]interface{})["d"].(float64)
		v.Price_r.N = b[i].(map[string]interface{})["price_r"].(map[string]interface{})["n"].(float64)
		c[i] = v
	}
	return c
}

func SdexOrderbookToUMF(asset string, p *map[string]interface{}) { // {{{1
	v := *p
	asks, _ := v["asks"]
	bids, _ := v["bids"]
	feed := make([]interface{}, 2)
	feed[0], feed[1] = ob(asks.([]interface{})), ob(bids.([]interface{}))
	encodeUMF(newUMF(asset, "sdex", feed))
	w.Flush()
}

func SdexTradeToUMF(asset string, p *map[string]interface{}) { // {{{1
	encodeUMF(newUMF(asset, "sdex", tradeAsArray(p)))
	w.Flush()
}

func BitfinexTradesToUMF(asset string, p *[]interface{}) { // {{{1
	array := *p
	for i := len(array) - 1; i >= 0; i-- {
		trade := array[i]
		encodeUMF(newUMF(asset, "bitfinex", tradeBitfinex(asset, trade.([]interface{}))))
	}
	w.Flush()
}

func CoinbaseTradesToUMF(asset string, p *[]interface{}) { // {{{1
	array := *p
	for i := len(array) - 1; i >= 0; i-- {
		trade := array[i]
		encodeUMF(newUMF(asset, "coinbase", tradeCoinbase(asset, trade.(map[string]interface{}))))
	}
	w.Flush()
}

func KrakenTradesToUMF(asset string, p *map[string]interface{}) { // {{{1
	v := *p
	if err, _ := v["error"].([]interface{}); len(err) > 0 {
		panic(err)
	}
	result := v["result"].(map[string]interface{})
	var trades []interface{}
	for key := range result {
		if key != "last" {
			trades = result[key].([]interface{})
			break
		}
	}
	for _, trade := range trades {
		encodeUMF(newUMF(asset, "kraken", tradeKraken(asset, trade.([]interface{}))))
	}
	w.Flush()
}
