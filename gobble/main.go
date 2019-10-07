package main

// import {{{1
import (
	"bufio"
	"encoding/json"
	aj "github.com/amissine/alik/json"
	"log"
	"os"
)

// see also:
// - https://golang.org/doc/code.html
// - https://blog.golang.org/json-and-go

// locals {{{1
var trades int = 0

type AE struct {
	Asset, Exchange string
}

var prev map[AE][]interface{} = make(map[AE][]interface{})

const UNDEF = "???"

func main() { // {{{1
	log.Println(os.Getpid(), "gobble started")
	dec := json.NewDecoder(os.Stdin)
	w := bufio.NewWriter(os.Stdout)
	enc := json.NewEncoder(w)
	for {
		var v aj.Umf
		if e := dec.Decode(&v); e != nil {
			log.Println("dec.Decode", e)
			break
		}
		if skip(&v) {
			continue
		}
		if e := enc.Encode(v); e != nil {
			log.Println("enc.Encode", e)
			break
		}
		w.Flush()
	}
	log.Println(os.Getpid(), trades, "trades, gobble exiting...")
}

func skip(q *aj.Umf) bool { // {{{1
	ae := AE{
		Asset:    q.Asset,
		Exchange: q.Exchange,
	}
	if prev[ae] == nil {
		prev[ae] = make([]interface{}, 5)
	}
	pae := prev[ae]
	if q.Trade[0] == pae[0] && q.Trade[1] == pae[1] && q.Trade[2] == pae[2] {
		return true
	}

	trades++
	pae[0] = q.Trade[0]
	pae[1] = q.Trade[1]
	pae[2] = q.Trade[2]
	price := q.Trade[2].(float64)

	longterm_p := 0.0
	if pae[3] != nil {
		longterm_p = pae[3].(float64)
	}
	longterm_q := ema(0.01, price, longterm_p)
	pae[3] = longterm_q

	shortterm_p := 0.0
	if pae[4] != nil {
		shortterm_p = pae[4].(float64)
	}
	shortterm_q := ema(0.99, price, shortterm_p)
	pae[4] = shortterm_q

	signal_returned := signal(longterm_p, longterm_q, shortterm_p, shortterm_q)
	q.Trade = append(q.Trade, pae[3], pae[4], signal_returned)

	//	if trades%100 == 0 {
	//		log.Println(trades, *q)
	//	}
	return signal_returned == UNDEF
}

func signal(ltp, ltq, stp, stq float64) string { // {{{1
	// ltp  stq
	// stp  ltq    buy
	//
	// stp  ltq
	// ltp  stq    sell
	if ltq <= ltp {
		if stq <= stp || stq <= ltq {
			return UNDEF
		}
		return "buy"
	} // else ltq > ltp
	if stq >= stp || stq >= ltq {
		return UNDEF
	}
	return "sell"
}

func ema(a, y, s float64) float64 { // {{{1
	if s == 0 {
		return y
	}
	return a*y + (1-a)*s
}
