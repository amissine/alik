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
var (
	prev         map[aj.AE][]interface{}  = make(map[aj.AE][]interface{})
	pTrade4Asset map[string][]interface{} = make(map[string][]interface{})
	sdexXMLinUSD float64
)

const (
	UNDEF          = "???"
	aLtEMA float64 = 0.02
	aStEMA float64 = 0.1
)

func main() { // {{{1
	log.Println(os.Getpid(), "gobble started")
	dec := json.NewDecoder(bufio.NewReaderSize(os.Stdin, 65536))
	w := bufio.NewWriterSize(os.Stdout, 65536)
	enc := json.NewEncoder(w)
	for {
		var v aj.Umf
		if e := dec.Decode(&v); e != nil {
			log.Println(os.Getpid(), "dec.Decode", e)
			break
		}
		if noSignal4Asset(&v) {
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
func noSignal4Asset(q *aj.Umf) bool { // {{{1
	if !allTradesInUSD(q) {
		return true
	}
	if pTrade4Asset[q.Asset] == nil {
		pTrade4Asset[q.Asset] = make([]interface{}, 2)
		pTrade4Asset[q.Asset][0] = q.Trade[2] // long-term EMA init value
		pTrade4Asset[q.Asset][1] = q.Trade[2] // short-term EMA init value
		return true
	}
	pLtEMA := pTrade4Asset[q.Asset][0].(float64)
	pStEMA := pTrade4Asset[q.Asset][1].(float64)
	price := q.Trade[2].(float64)
	qLtEMA := aLtEMA*price + (1.0-aLtEMA)*pLtEMA
	qStEMA := aStEMA*price + (1.0-aStEMA)*pStEMA
	signal_returned := signal(pLtEMA, qLtEMA, pStEMA, qStEMA)
	pTrade4Asset[q.Asset][0], pTrade4Asset[q.Asset][1] = qLtEMA, qStEMA
	q.Trade = append(q.Trade, signal_returned)
	return signal_returned == UNDEF
}

// Assets are being traded on sdex and other exchanges. {{{1
// On sdex, assets are being traded in XML. On other exchanges, assets are being
// traded in USD. On sdex, USD is just another asset traded in XML. On other
// exchanges, XML is just another asset traded in USD. To trade all assets
// (including XML) on all exchanges (including sdex) in USD, we introduce the
// following pseudocode:
//
// 1. [Skip other exchanges.] Leave a non-sdex trade data intact, exit this code.
//
// 2. [Asset == USD] If asset == USD, set:
// - asset = XML;
// - amount = amount*price; // Now it is the amount of XML traded for USD
// - price = 1/price;       // Now it is the price of XML in USD
// - sdexXMLinUSD = price.
//
// 3. [Asset != USD] Otherwise, set price = price*sdexXMLinUSD.
//
// For example, 1 USD @ 10 XML becomes 10 XML @ 0.1 USD, and sdexXMLinUSD = 0.1.
// Then, 1 BTC @ 70000 XML becomes 1 BTC @ 7000 USD.
func allTradesInUSD(q *aj.Umf) bool { // {{{1
	if q.Exchange != "sdex" {
		return true
	}
	if q.Asset == "USD" {
		q.Asset = "XML"
		q.Trade[1] = q.Trade[1].(float64) * q.Trade[2].(float64)
		q.Trade[2] = 1.0 / q.Trade[2].(float64)
		sdexXMLinUSD = q.Trade[2].(float64)
	} else {
		if sdexXMLinUSD == 0.0 {
			return false
		}
		q.Trade[2] = q.Trade[2].(float64) * sdexXMLinUSD
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
	return UNDEF
}
