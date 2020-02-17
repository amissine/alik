package json // {{{1

import ( // {{{1
	"bufio"
	"encoding/json"
	"log"
	"os"
	"sort"
	"strconv"
	"time"
)

const ENCODER_BUFFER_SIZE = 1024 * 1024 // {{{1
var w = bufio.NewWriterSize(os.Stdout, ENCODER_BUFFER_SIZE)
var enc = json.NewEncoder(w)
var s2e = make(map[string][]interface{}) // asset-keyed slices to encode {{{2
// An slice to encode for a given asset has the following 5 elements:
// - the asset code, for example "XLM";
// - trades for the asset since the previous s2e[asset] has been encoded;
// - the time now;
// - the latest SDEX asks (of type OrderBook);
// - the latest SDEX bids (of type OrderBook).
//
// The trades element is an slice of trades on different exchanges. Each trade is
// a slice of the following 4 elements:
// - the time of the trade;
// - the exchange code, for example "kraken";
// - the amount of the asset bought/sold, +/-;
// - the price.
//
// On all exchanges other than SDEX, the price of the asset is given in USD.
// On SDEX, the price is given in XLM. For the "XLM" asset on SDEX, amount is
// actually the amount of USD that has been bought or sold for the price in XLM.

func encodeUMF(p *UMF) { // {{{2
	q, ok := s2e[p.AE.Asset]
	if !ok {
		q = make([]interface{}, 5)
		s2e[p.AE.Asset] = q
		q[0] = p.AE.Asset
		q[1] = make([]interface{}, 0)
	}
	if p.IsTrade() {
		trade := make([]interface{}, 0)
		trade = append(trade, p.Feed[0], p.AE.Exchange, p.Feed[1], p.Feed[2])
		q[1] = append(q[1].([]interface{}), trade)
	} else {
		// Sort trades, add OrderBook time, asks, and bids. {{{3
		trades := q[1].([]interface{})
		sort.Slice(trades, func(i, j int) bool {
			timeI := trades[i].([]interface{})[0].(time.Time)
			timeJ := trades[j].([]interface{})[0].(time.Time)
			return timeI.Before(timeJ)
		})
		q[1] = trades
		q[2] = p.UTC
		q[3] = p.Feed[0]
		q[4] = p.Feed[1] // }}}3
		if err := enc.Encode(&q); err != nil {
			log.Fatal(os.Getpid(), "encodeUMF", err)
		}
		w.Flush()
		q[1] = make([]interface{}, 0) // cleanup {{{3
		q[2] = nil
		q[3] = nil
		q[4] = nil // }}}3
	}
} // }}}2

type UMF struct { // Unified Market Feed {{{1
	AE   AE
	Feed []interface{} // either [asks, bids] or tradeAsArray or tradeBitfinex...
	UTC  time.Time
}

type AE struct { // Asset, Exchange {{{1
	Asset, Exchange string
}

type OrderBook []struct { // {{{1
	Amount, Price string // Price in XLM
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
	a := asset
	if a == "USD" {
		a = "XLM"
	}
	v := *p
	asks, _ := v["asks"]
	bids, _ := v["bids"]
	feed := make([]interface{}, 2)
	feed[0], feed[1] = ob(asks.([]interface{})), ob(bids.([]interface{}))
	encodeUMF(newUMF(a, "sdex", feed))
}

func SdexTradeToUMF(asset string, p *map[string]interface{}) { // {{{1
	a := asset
	if a == "USD" {
		a = "XLM"
	}
	encodeUMF(newUMF(a, "sdex", tradeAsArray(p)))
}

func BitfinexTradesToUMF(asset string, p *[]interface{}) { // {{{1
	a := asset
	if a == "USD" {
		a = "XLM"
	}
	array := *p
	for i := len(array) - 1; i >= 0; i-- {
		trade := array[i]
		encodeUMF(newUMF(a, "bitfinex", tradeBitfinex(asset, trade.([]interface{}))))
	}
}

func CoinbaseTradesToUMF(asset string, p *[]interface{}) { // {{{1
	a := asset
	if a == "USD" {
		a = "XLM"
	}
	array := *p
	for i := len(array) - 1; i >= 0; i-- {
		trade := array[i]
		encodeUMF(newUMF(a, "coinbase", tradeCoinbase(asset, trade.(map[string]interface{}))))
	}
}

func KrakenTradesToUMF(asset string, p *map[string]interface{}) { // {{{1
	a := asset
	if a == "USD" {
		a = "XLM"
	}
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
		encodeUMF(newUMF(a, "kraken", tradeKraken(asset, trade.([]interface{}))))
	}
}

func (this *UMF) IsTrade() bool { // {{{1
	return len(this.Feed) > 2
}
