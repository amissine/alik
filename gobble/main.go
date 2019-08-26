package main

// import {{{1
import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"time"
)

// see also:
// - https://golang.org/doc/code.html
// - https://blog.golang.org/json-and-go
// - https://www.ardanlabs.com/blog/2013/07/understanding-pointers-and-memory.html

type MarketFeed struct { // {{{1
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

func (this *MarketFeed) Init(v *map[string]interface{}) *MarketFeed { // {{{2
	w := *v
	base, _ := w["base"]
	asks, _ := w["asks"]
	bids, _ := w["bids"]
	loc, _ := time.LoadLocation("UTC")
	this = &MarketFeed{
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

func (this *MarketFeed) Same(mf *MarketFeed) bool { // {{{2
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

func (this *MarketFeed) Duplicate(mf *MarketFeed) bool { // {{{2
	if mf == nil {
		return false
	}
	if !this.UTC.Equal(mf.UTC) {
		return false
	}
	return this.Same(mf)
}

func main() { // {{{1
	log.Println(os.Getpid(), "gobble started")
	dec := json.NewDecoder(os.Stdin)
	w := bufio.NewWriter(os.Stdout)
	enc := json.NewEncoder(w)
	var p, q *MarketFeed
	for {
		var v MarketFeed
		if e := dec.Decode(&v); e != nil {
			log.Println("dec.Decode", e)
			break
		}
		if q = &v; q.Same(p) {
			log.Println("skipping", *q)
			continue
		}
		p = q
		if e := enc.Encode(q); e != nil {
			log.Println("enc.Encode", e)
			break
		}
		w.Flush()
	}
	log.Println(os.Getpid(), "gobble exiting..")
}
