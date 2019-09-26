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

var previous *Umf

type Spread struct { // {{{1
	Asks, Bids OrderBook
}

type OrderBook []struct { // {{{1
	Amount, Price string
	Price_r       struct{ D, N float64 }
}

func (this *Umf) Make(e, a string, v *map[string]interface{}) *Umf { // {{{1
	if isTrade(e, v) {
		return trade(e, a, v, this)
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

func isTrade(exchange string, v *map[string]interface{}) bool { // {{{1
	//log.Println("Umf.isTrade", *v)
	return exchange == "sdex" && (*v)["price"] != nil // TODO other exchanges
}

func trade(e, a string, v *map[string]interface{}, t *Umf) *Umf { // {{{1
	if t != nil { // TODO do not ignore previous trades
		return t
	}
	this := &Umf{
		Asset:    a,
		Exchange: "sdex", // TODO other exchanges - use e arg value
		Trade:    tradeAsArray(v),
	}
	//log.Println("trade this =", this)
	return this
}

// Presently, tradeAsArray returns: {{{1
//
// [
//   base_amount,
//   base_is_seller,
//   base_offer_id,
//   counter_amount,
//   counter_offer_id,
//   ledger_close_time,
//   offer_id,
//   priceInXLM
// ]
//
// When done, there will be only three values in the resulting array: {{{2
//
// [ ledger_close_time, base_amount, priceInXLM ]
//
// where base_amount will be positive when the trade is a buy, and negative when
// the trade is a sell. Occasionally, the fourth element may be present in this
// array. When its value is false, this means the trade is not sequential. In other
// words, trade loss is possible between this trade and the previous trade in the
// market feed sequence.
//
// By definition,
//
//     priceInXLM = price.n/price.d
//
// It is worth noting that
//
//     priceInXLM ~= counter_amount/base_amount
//
// I am doing this intermediate step to make sure I operate correctly with
// base_is_seller, base_offer_id, counter_offer_id, and offer_id values. Before I
// set up the rules, I would like to take a look at some use cases.
//
// }}}2
// The rule appears to be quite simple: {{{2
//
//    base_is_seller == false  ==>  the trade is a sell
//    base_is_seller == true   ==>  the trade is a buy
//
// }}}2
// }}}1
func tradeAsArray(v *map[string]interface{}) []interface{} { // {{{1
	w := *v
	price := w["price"].(map[string]interface{})
	base_amount, e := strconv.ParseFloat(w["base_amount"].(string), 64)
	if e != nil {
		log.Println("tradeAsArray ERROR:", e)
	}
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
	if this.UTC.IsZero() || this.Same(previous) {
		//log.Println("Skip this =", this)
		return true
	}
	previous = this
	return false
}

func (this *Umf) Same(mf *Umf) bool { // {{{1
	if mf == nil {
		return false
	}
	for i, e := range mf.Trade {
		if this.Trade[i] != e {
			return false
		}
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
