package main // see also: {{{1
// - https://golang.org/doc/code.html
// - https://blog.golang.org/json-and-go

// import {{{1
import (
	"bufio"
	"encoding/json"
	aj "github.com/amissine/alik/json"
	"log"
	"os"
)

// locals {{{1
var (
	pTrade4Asset map[string][]interface{} = make(map[string][]interface{})
	sdexXLMinUSD float64
)

const (
	UNDEF          = "???"
	aLtEMA float64 = 0.001
	aStEMA float64 = 0.01
)

func main() { // {{{1
	log.Println(os.Getpid(), "gobble started")
	dec := json.NewDecoder(bufio.NewReaderSize(os.Stdin, 65536))
	w := bufio.NewWriterSize(os.Stdout, 65536)
	enc := json.NewEncoder(w)
	lineNumber := 0
	for {
		var v aj.UMF
		lineNumber++
		if e := dec.Decode(&v); e != nil {
			log.Println(os.Getpid(), "lineNumber", lineNumber, "dec.Decode", e)
			break
		}
		if !v.IsTrade() || noSignal(&v) {
			continue
		}
		if e := enc.Encode(v); e != nil {
			log.Println(os.Getpid(), "enc.Encode", e)
			break
		}
		w.Flush()
	}
}

// Exponential Moving Average (EMA): {{{1
//
//  https://en.wikipedia.org/wiki/Moving_average#Exponential_moving_average
//
func noSignal(q *aj.UMF) bool { // {{{1
	if !allTradesInUSD(q) {
		return true
	}
	if pTrade4Asset[q.AE.Asset] == nil {
		pTrade4Asset[q.AE.Asset] = make([]interface{}, 2)
		pTrade4Asset[q.AE.Asset][0] = q.Feed[2] // long-term EMA init value
		pTrade4Asset[q.AE.Asset][1] = q.Feed[2] // short-term EMA init value
		return true
	}
	pLtEMA := pTrade4Asset[q.AE.Asset][0].(float64)
	pStEMA := pTrade4Asset[q.AE.Asset][1].(float64)
	price := q.Feed[2].(float64)
	qLtEMA := aLtEMA*price + (1.0-aLtEMA)*pLtEMA
	qStEMA := aStEMA*price + (1.0-aStEMA)*pStEMA
	signal_returned := signal(pLtEMA, qLtEMA, pStEMA, qStEMA)
	pTrade4Asset[q.AE.Asset][0], pTrade4Asset[q.AE.Asset][1] = qLtEMA, qStEMA
	q.Feed = append(q.Feed, signal_returned)
	q.AssumeTrade()
	return signal_returned == UNDEF
}

// Assets are being traded on sdex and other exchanges. {{{1
// On sdex, assets are being traded in XLM. On other exchanges, assets are being
// traded in USD. On sdex, USD is just another asset traded in XLM. On other
// exchanges, XLM is just another asset traded in USD. To trade all assets
// (including XLM) on all exchanges (including sdex) in USD, we introduce the
// following pseudocode:
//
// 1. [Skip other exchanges.] Leave a non-sdex trade data intact, exit this code.
//
// 2. [Asset == USD] If asset == USD, set:
// - asset = XLM;
// - amount = amount*price; // Now it is the amount of XLM traded for USD
// - price = 1/price;       // Now it is the price of XLM in USD
// - sdexXLMinUSD = price.
//
// 3. [Asset != USD] Otherwise, set price = price*sdexXLMinUSD.
//
// For example, 1 USD @ 10 XLM becomes 10 XLM @ 0.1 USD, and sdexXLMinUSD = 0.1.
// Then, 1 BTC @ 70000 XLM becomes 1 BTC @ 7000 USD.
func allTradesInUSD(q *aj.UMF) bool { // {{{1
	if q.AE.Exchange != "sdex" {
		return true
	}
	if q.AE.Asset == "USD" {
		q.AE.Asset = "XLM"
		q.Feed[1] = q.Feed[1].(float64) * q.Feed[2].(float64)
		q.Feed[2] = 1.0 / q.Feed[2].(float64)
		sdexXLMinUSD = q.Feed[2].(float64)
	} else {
		if sdexXLMinUSD == 0.0 {
			return false
		}
		q.Feed[2] = q.Feed[2].(float64) * sdexXLMinUSD
	}
	return true
}

func signal(pLt, qLt, pSt, qSt float64) string { // {{{1
	// pLt  qSt
	// pSt  qLt    buy
	//
	// pSt  qLt
	// pLt  qSt    sell
	if pLt >= pSt && qSt > qLt {
		return "buy"
	}
	if pSt >= pLt && qLt > qSt {
		return "sell"
	}
	return UNDEF // TODO "stop", "buy" and "sell" above being start events
}
