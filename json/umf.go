package json

// import {{{1
import (
	"log"
	"time"
)

type Umf struct { // {{{1
	Asset, Exchange string
	Ob              Spread
	Trade           []interface{}
	UTC             time.Time
}

type Spread struct { // {{{1
	Asks, Bids OrderBook
}

type OrderBook []struct { // {{{1
	Amount, Price string
	Price_r       struct{ D, N float64 }
}

func (this *Umf) Init(v *map[string]interface{}) *Umf { // {{{1
	if isTrade(v) {
		return trade(v)
	}
	w := *v // Init Spread {{{2
	base, _ := w["base"]
	asks, _ := w["asks"]
	bids, _ := w["bids"]
	loc, _ := time.LoadLocation("UTC")
	this = &Umf{
		Asset:    base.(map[string]interface{})["asset_code"].(string),
		Exchange: "sdex", // TODO other exchanges
		Ob: Spread{
			Asks: ob(asks.([]interface{})),
			Bids: ob(bids.([]interface{})),
		},
		UTC: time.Now().In(loc),
	}
	return this // }}}2
}

func isTrade(v *map[string]interface{}) bool { // {{{1
	log.Println("Umf.isTrade", *v)
	/* An example of trade json {{{2
	   map[
	     _links: map[
	       base: map[
	   	    href: https://horizon.stellar.org/accounts/GDQ76ÃÂÃÂÃÂÃÂ¢ÃÂÃÂÃÂÃÂÃÂÃÂÃÂÃÂ¦XYDUC
	   	  ]
	   		counter: map[
	   		  href: https://horizon.stellar.org/accounts/GB3LQÃÂÃÂÃÂÃÂ¢ÃÂÃÂÃÂÃÂÃÂÃÂÃÂÃÂ¦JJMKO
	   		]
	   		operation: map[
	   		  href: https://horizon.stellar.org/operations/109964094924283905
	   		]
	   		self: map[ href: ]
	   	]
	   	base_account: GDQ76ÃÂÃÂÃÂÃÂ¢ÃÂÃÂÃÂÃÂÃÂÃÂÃÂÃÂ¦XYDUC
	   	base_amount: 1.3481683
	   	base_asset_code: CNY
	   	base_asset_issuer:GAREEÃÂÃÂÃÂÃÂ¢ÃÂÃÂÃÂÃÂÃÂÃÂÃÂÃÂ¦3RFOX
	   	base_asset_type: credit_alphanum4
	   	base_is_seller: false
	   	base_offer_id: 4721650113351671809
	   	counter_account: GB3LQÃÂÃÂÃÂÃÂ¢ÃÂÃÂÃÂÃÂÃÂÃÂÃÂÃÂ¦JJMKO
	   	counter_amount: 2.9959295
	   	counter_asset_type: native
	   	counter_offer_id: 111904381
	   	id: 109964094924283905-2
	   	ledger_close_time: 2019-08-31T18:35:22Z
	   	offer_id: 111904381
	   	paging_token: 109964094924283905-2
	   	price: map[d: 9 n: 20]
	   ]
	*/ // }}}2
	return (*v)["price"] != nil
}

func trade(v *map[string]interface{}) *Umf { // {{{1
	loc, _ := time.LoadLocation("UTC")
	this := &Umf{
		Asset:    (*v)["base_asset_code"].(string),
		Exchange: "sdex", // TODO other exchanges
		Trade:    tradeAsArray(v),
		UTC:      time.Now().In(loc),
	}
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
// When done, there will be only two values in the resulting array: {{{2
//
// [ base_amount, priceInXLM ]
//
// where base_amount will be positive when the trade is a buy, and negative when
// the trade is a sell. And
//
//     priceInXLM = price.n/price.d
//
// It is worth noting that
//
//     priceInXLM ~= counter_amount/base_amount
//
// I am doing this intermediate step to make sure I operate correctly with
// base_is_seller, base_offer_id, counter_offer_id, and offer_id values. Before I
// set up the rules, I would like to take a look at some use cases. }}}2
func tradeAsArray(v *map[string]interface{}) []interface{} {
	return nil
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
	return false
}

func (this *Umf) Same(mf *Umf) bool { // {{{1
	if mf == nil {
		return false
	}
	if this.Trade != nil {
		if mf.Trade == nil {
			return false
		}
		for i, e := range mf.Trade {
			if this.Trade[i] != e {
				return false
			}
		}
		return true
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
