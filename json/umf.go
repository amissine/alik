package json

// import {{{1
import (
	"time"
)

type Umf struct { // {{{1
	Ob    Spread
	Trade []interface{}
	UTC   time.Time
}

type Spread struct {
	Asset      string
	Asks, Bids OrderBook
} // {{{2

type OrderBook []struct { // {{{2
	Amount, Price string
	Price_r       struct{ D, N float64 }
}

func (this *Umf) Init(v *map[string]interface{}) *Umf { // {{{2
	w := *v
	base, _ := w["base"]
	asks, _ := w["asks"]
	bids, _ := w["bids"]
	loc, _ := time.LoadLocation("UTC")
	this = &Umf{
		Ob: Spread{
			Asset: base.(map[string]interface{})["asset_code"].(string),
			Asks:  ob(asks.([]interface{})),
			Bids:  ob(bids.([]interface{})),
		},
		UTC: time.Now().In(loc),
	}
	return this
}

func ob(b []interface{}) OrderBook {
	c := make(OrderBook, len(b))
	for i, v := range c {
		v.Amount = b[i].(map[string]interface{})["amount"].(string)
		v.Price = b[i].(map[string]interface{})["price"].(string)
		v.Price_r.D =
			b[i].(map[string]interface{})["price_r"].(map[string]interface{})["d"].(float64)
		v.Price_r.N =
			b[i].(map[string]interface{})["price_r"].(map[string]interface{})["n"].(float64)
		c[i] = v
	}
	return c
}

func (this *Umf) Same(mf *Umf) bool { // {{{2
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
	if this.Ob.Asset != mf.Ob.Asset || len(this.Ob.Asks) != len(mf.Ob.Asks) ||
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

func (this *Umf) Duplicate(mf *Umf) bool { // {{{2
	if mf == nil {
		return false
	}
	if !this.UTC.Equal(mf.UTC) {
		return false
	}
	return this.Same(mf)
}
