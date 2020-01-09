package json

// import {{{1
import (
	"log"
	"strconv"
	"time"
)

type Umf struct { // {{{1
	Asset, Exchange string
	Ob              Spread
	Trade           []interface{}
	UTC             time.Time
}

type AE struct { // {{{1
	Asset, Exchange string
}

var pUmf = make(map[AE]*Umf)

type Spread struct { // {{{1
	Asks, Bids OrderBook
}

type OrderBook []struct { // {{{1
	Amount, Price string
	Price_r       struct{ D, N float64 }
}

func (this *Umf) Make(e, a string, v *map[string]interface{}) *Umf { // {{{1
	if sdexTrade(e, v) {
		return sdexMakeTrade(e, a, v, this)
	}
	w := *v // Make Spread {{{2
	asks, _ := w["asks"]
	bids, _ := w["bids"]
	loc, _ := time.LoadLocation("UTC")
	this.Ob = Spread{
		Asks: ob(asks.([]interface{})),
		Bids: ob(bids.([]interface{})),
	}
	this.UTC = time.Now().In(loc)
	return this // }}}2
}

func (this *Umf) MakeTrade(e, a string, v *[]interface{}) *Umf { // {{{1
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
	return &Umf{
		Asset:    a,
		Exchange: e,
		Trade:    trade,
		UTC:      time.Now().In(loc),
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

func sdexTrade(exchange string, v *map[string]interface{}) bool { // {{{1
	return exchange == "sdex" && (*v)["price"] != nil
}

func sdexMakeTrade(e, a string, v *map[string]interface{}, t *Umf) *Umf { // {{{1
	if t != nil { // TODO do not ignore previous trades
		return t
	}
	this := &Umf{
		Asset:    a,
		Exchange: "sdex",
		Trade:    tradeAsArray(v),
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

func (this *Umf) Skip() bool { // {{{1
	if this.UTC.IsZero() || this.Same(pUmf[AE{this.Asset, this.Exchange}]) {
		return true
	}
	pUmf[AE{this.Asset, this.Exchange}] = this
	return false
}

func (this *Umf) SkipTrade() bool { // {{{1
	if sameTrade(this, pUmf[AE{this.Asset, this.Exchange}]) {
		return true
	}
	pUmf[AE{this.Asset, this.Exchange}] = this
	return false
}

func sameTrade(this, mf *Umf) bool { // {{{1
	if mf == nil {
		return false
	}
	for i, e := range mf.Trade {
		if this.Trade[i] != e {
			return false
		}
	}
	return true
}

func (this *Umf) Same(mf *Umf) bool { // {{{1
	if !sameTrade(this, mf) {
		return false
	}
	if this.Asset != mf.Asset || len(this.Ob.Asks) != len(mf.Ob.Asks) ||
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
