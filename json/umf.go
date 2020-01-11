package json

// import {{{1
import (
	"log"
	"strconv"
	"time"
)

type UMF struct { // Unified Market Feed{{{1
	AE   AE
	Feed []interface{} // either [Spread] or tradeAsArray or tradeBitfinex...
	UTC  time.Time
}

type AE struct { // {{{1
	Asset, Exchange string
}

var pUMF = make(map[AE]*UMF)

type Spread struct { // {{{1
	Asks, Bids OrderBook
}

type OrderBook []struct { // {{{1
	Amount, Price string
	Price_r       struct{ D, N float64 }
}

func (this *UMF) MakeSDEX(a string, v *map[string]interface{}) *UMF { // {{{1
	if tradeSDEX(v) {
		return makeTradeSDEX(a, v, this)
	}
	w := *v
	asks, _ := w["asks"]
	bids, _ := w["bids"]
	feed := make([]interface{}, 1)
	feed[0] = Spread{Asks: ob(asks.([]interface{})), Bids: ob(bids.([]interface{}))}
	location, _ := time.LoadLocation("UTC")
	this = &UMF{
		AE:   AE{Asset: a, Exchange: "sdex"},
		Feed: feed,
		UTC:  time.Now().In(location),
	}
	return this
}

func (this *UMF) MakeTrade(e, a string, v *[]interface{}) *UMF { // {{{1
	var trade []interface{}
	loc, _ := time.LoadLocation("UTC")
	switch e {
	case "bitfinex":
		trade = tradeBitfinex(a, v)
	case "coinbase":
		trade = tradeCoinbase(a, v)
	default:
		panic("TODO implement trade") // TODO implement trade for e
	}
	return &UMF{
		AE:   AE{Asset: a, Exchange: e},
		Feed: trade,
		UTC:  time.Now().In(loc),
	}
}

func tradeBitfinex(a string, v *[]interface{}) []interface{} { // {{{1
	t := make([]interface{}, 3)
	w := (*v)[0].(map[string]interface{})
	amount := parse(w["amount"].(string))
	loc, _ := time.LoadLocation("UTC")
	t[0] = time.Unix(int64(w["timestamp"].(float64)), 0).In(loc)
	if w["type"].(string) == "buy" {
		t[1] = amount
	} else {
		t[1] = -amount
	}
	t[2] = parse(w["price"].(string)) // price in USD
	return t
}

func tradeCoinbase(a string, v *[]interface{}) []interface{} { // {{{1
	t := make([]interface{}, 3)
	w := (*v)[0].(map[string]interface{})
	size := parse(w["size"].(string))
	t[0] = w["time"].(string)
	if w["side"].(string) == "sell" {
		t[1] = size
	} else {
		t[1] = -size
	}
	t[2] = parse(w["price"].(string)) // price in USD
	return t
}

func parse(s string) float64 { // {{{1
	f, e := strconv.ParseFloat(s, 64)
	if e != nil {
		log.Println("parse ERROR", e)
	}
	return f
}

func tradeSDEX(v *map[string]interface{}) bool { // {{{1
	return (*v)["price"] != nil
}

func makeTradeSDEX(a string, v *map[string]interface{}, t *UMF) *UMF { // {{{1
	if t != nil { // TODO do not ignore previous trades
		return t // v represents a previous trade, we are ignoring it for now
	}
	location, _ := time.LoadLocation("UTC")
	this := &UMF{
		AE:   AE{Asset: a, Exchange: "sdex"},
		Feed: tradeAsArray(v),
		UTC:  time.Now().In(location),
	}
	return this
}

// Presently, tradeAsArray returns: {{{1
//
// [ ledger_close_time, base_amount, priceInXLM ]
//
// where base_amount will be positive when the trade is a buy, and negative when
// the trade is a sell. Occasionally, the fourth element may be present in this
// array. When its value is false, this means the trade is not sequential. In other
// words, trade loss is possible between this trade and the previous trade in the
// market feed sequence.
// }}}1
func tradeAsArray(v *map[string]interface{}) []interface{} { // {{{1
	w := *v
	price := w["price"].(map[string]interface{})
	base_amount := parse(w["base_amount"].(string))
	a := make([]interface{}, 3)
	priceInXLM := price["n"].(float64) / price["d"].(float64)
	a[0] = w["ledger_close_time"]
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

func (this *UMF) SkipSDEX() bool { // {{{1
	if this.Same(pUMF[this.AE]) {
		return true
	}
	pUMF[this.AE] = this
	return false
}

func (this *UMF) SkipTrade() bool { // {{{1
	if this.SameTrade(pUMF[this.AE]) {
		return true
	}
	pUMF[this.AE] = this
	return false
}

func (this *UMF) SameTrade(mf *UMF) bool { // {{{1
	if mf == nil || len(this.Feed) < 3 || len(mf.Feed) < 3 {
		return false
	}
	for i, e := range mf.Feed {
		if this.Feed[i] != e {
			return false
		}
	}
	return true
}

func (this *UMF) IsTrade() bool { // {{{1
	return len(this.Feed) > 2
}
func (this *UMF) Same(mf *UMF) bool { // {{{1
	if this.IsTrade() && mf.IsTrade() {
		return this.SameTrade(mf)
	}
	if this.IsTrade() || mf.IsTrade() {
		return false
	}
	if this.AE.Asset != mf.AE.Asset || len(this.Ob.Asks) != len(mf.Ob.Asks) ||
		len(this.Ob.Bids) != len(mf.Ob.Bids) {
		return false
	}
	for i, o := range mf.Ob.Asks {
		if this.Ob.Asks[i].Amount != o.Amount || this.Ob.Asks[i].Price != o.Price {
			return false
		}
	}
	for i, o := range mf.Ob.Bids {
		if this.Ob.Bids[i].Amount != o.Amount || this.Ob.Bids[i].Price != o.Price {
			return false
		}
	}
	return true
}

func (this *Umf) Duplicate(mf *Umf) bool { // {{{1
	if mf == nil {
		return false
	}
	if !this.UTC.Equal(mf.UTC) {
		return false
	}
	return this.Same(mf)
}
