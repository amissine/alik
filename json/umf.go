package json

// import {{{1
import (
	"log"
	"strconv"
	"time"
)

type UMF struct { // Unified Market Feed{{{1
	AE   AE
	Feed []interface{} // either [asks, bids] or tradeAsArray or tradeBitfinex...
	UTC  time.Time
}

type AE struct { // {{{1
	Asset, Exchange string
}

type FeedHistory struct { // {{{1
	LT, OT, AB *UMF // Latest Trade (LT), Oldest Trade (OT), Asks and Bids (AB)
}

var fhm = make(map[AE]FeedHistory)

func add2fhm(mf *UMF) (bool, *UMF) {
	var fh FeedHistory
	var present bool
	if fh, present = fhm[mf.AE]; !present { // {{{2
		fhm[mf.AE] = FeedHistory{LT: mf}
		return true, mf
	}
	if mf.IsTrade() { // {{{2
		if mf.TradeBefore(fh.LT) { // Latest Trade comes first
			if fh.OT == nil { // {{{3
				fh.OT = mf
				return true, nil
			} else {
				return false, nil
			} // }}}3
		} else {
			if fh.OT == nil { // {{{3
				panic("trade out of sequence")
			} else {
				return false, nil
			} // }}}3
		}
	} else { // Asks and Bids {{{2
		if fh.AB == nil {
			fh.AB = mf
			return true, mf
		} else {
			return false, nil
		}
	} // }}}2
}

type OrderBook []struct { // {{{1
	Amount, Price string
	Price_r       struct{ D, N float64 }
}

func (this *UMF) MakeSDEX(a string, v *map[string]interface{}) *UMF { // {{{1
	if tradeSDEX(v) {
		return newUMF(a, "sdex", tradeAsArray(v))
	}
	w := *v
	asks, _ := w["asks"]
	bids, _ := w["bids"]
	feed := make([]interface{}, 2)
	feed[0], feed[1] = ob(asks.([]interface{})), ob(bids.([]interface{}))
	return newUMF(a, "sdex", feed)
}

func (this *UMF) MakeTrade(e, a string, v *[]interface{}) *UMF { // {{{1
	var trade []interface{}
	switch e {
	case "bitfinex":
		trade = tradeBitfinex(a, v)
	case "coinbase":
		trade = tradeCoinbase(a, v)
	default:
		panic("TODO implement trade") // TODO implement trade for e
	}
	return newUMF(a, e, trade)
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

func (this *UMF) Skip() (bool, *UMF) { // {{{1
	if added, mf := add2fhm(this); added {
		return mf == nil || mf.IsTrade(), mf
	}
	fh, _ := fhm[this.AE] // present, not added
	if this.IsTrade() {   // {{{2
		if fh.LT.TradeBefore(this) {
			fh.OT = fh.LT
			fh.LT = this
			return true, nil
		} else {
			if !this.SameTrade(fh.OT) {
				fh.LT.Feed = append(fh.LT.Feed, false)
			}
			return false, fh.LT
		}
	} else { // Asks and Bids {{{2
		if this.Same(fh.AB) {
			return true, this
		} else {
			fh.AB = this
			return false, this
		}
	} // }}}2
}

func (this *UMF) TradeBefore(mf *UMF) bool { // {{{1
	if !this.IsTrade() {
		panic("must be a trade")
	}
	loc, _ := time.LoadLocation("UTC")
	t, _ := time.ParseInLocation(time.RFC3339, this.Feed[0].(string), loc)
	f, _ := time.ParseInLocation(time.RFC3339, mf.Feed[0].(string), loc)
	return t.Before(f)
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
	if this.IsTrade() || mf.IsTrade() { // {{{2
		return false
	}
	if this.AE != mf.AE ||
		len(this.Feed[0].(OrderBook)) != len(mf.Feed[0].(OrderBook)) ||
		len(this.Feed[1].(OrderBook)) != len(mf.Feed[1].(OrderBook)) {
		return false
	}
	for i, o := range mf.Feed[0].(OrderBook) { // asks
		if this.Feed[0].(OrderBook)[i].Amount != o.Amount ||
			this.Feed[0].(OrderBook)[i].Price != o.Price {
			return false
		}
	}
	for i, o := range mf.Feed[1].(OrderBook) { // bids
		if this.Feed[1].(OrderBook)[i].Amount != o.Amount ||
			this.Feed[1].(OrderBook)[i].Price != o.Price {
			return false
		}
	}
	return true
}
