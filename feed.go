package main // import {{{1

import (
  "encoding/json"
  "log"
  "os"
)

type MarketFeed struct { // {{{1
  Ob struct {
    Asset string; Asks, Bids OrderBook
  }
  Trade []interface{}
  Ts float64
}

type OrderBook []struct{ Amount, Price float64 } // {{{2

func (this *MarketFeed) Same (mf *MarketFeed) bool { // {{{2
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

func (this *MarketFeed) Duplicate (mf *MarketFeed) bool { // {{{2
  if mf == nil {
    return false
  }
  if this.Ts != mf.Ts {
    return false
  }
  return this.Same(mf)
}

func main () { // {{{1
  dec := json.NewDecoder(os.Stdin)
  enc := json.NewEncoder(os.Stdout)
  var p *MarketFeed
  for {
    var q MarketFeed
    if e := dec.Decode(&q); e != nil {
      log.Println(e)
      return
    }
    if q.Same(p) {
      log.Println(q)
      continue
    }
    p = &q
    if e := enc.Encode(&q); e != nil {
      log.Println(e)
      return
    }
  }
}
