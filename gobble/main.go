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
	aStEMA float64 = 0.005
)

func main() { // {{{1
	log.Println(os.Getpid(), "gobble started")
	dec := json.NewDecoder(bufio.NewReaderSize(os.Stdin, aj.ENCODER_BUFFER_SIZE))
	w := bufio.NewWriterSize(os.Stdout, 65536)
	enc := json.NewEncoder(w)
	lineNumber := 0
	for {
		var v []interface{}
		lineNumber++
		if e := dec.Decode(&v); e != nil {
			log.Println(os.Getpid(), "lineNumber", lineNumber, "dec.Decode", e)
			break
		}
		asset, trades := v[0].(string), v[1].([]interface{})
		for _, trade := range trades {
			if got, signal := signal4(asset, trade.([]interface{})); got {
				if e := enc.Encode(signal); e != nil {
					log.Println(os.Getpid(), "enc.Encode", e)
					os.Exit(1)
				}
				w.Flush()
			}
		}
	}
}

// Exponential Moving Average (EMA): {{{1
//
//  https://en.wikipedia.org/wiki/Moving_average#Exponential_moving_average
//
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
func signal4(asset string, trade []interface{}) (bool, []interface{}) { // {{{1
	if exchange := trade[1].(string); exchange == "sdex" { // fix sdex data {{{2
		if asset == "XLM" {
			amount := trade[2].(float64) // amount of USD traded
			price := trade[3].(float64)  // price in XLM
			amount *= price              // amount of XLM traded
			price = 1.0 / price          // price in USD
			sdexXLMinUSD = price
			trade[2] = amount
			trade[3] = price
		} else {
			if sdexXLMinUSD == 0.0 {
				return false, nil
			}
			trade[3] = trade[3].(float64) * sdexXLMinUSD
		}
	}
	if pTrade4Asset[asset] == nil { // set pTrade4Asset {{{2
		pTrade4Asset[asset] = make([]interface{}, 2)
		pTrade4Asset[asset][0] = trade[3] // long-term EMA init value
		pTrade4Asset[asset][1] = trade[3] // short-term EMA init value
		return false, nil
	}
	pLtEMA := pTrade4Asset[asset][0].(float64) // do the EMA step {{{2
	pStEMA := pTrade4Asset[asset][1].(float64)
	price := trade[3].(float64)
	qLtEMA := aLtEMA*price + (1.0-aLtEMA)*pLtEMA
	qStEMA := aStEMA*price + (1.0-aStEMA)*pStEMA
	signal_returned := signal(pLtEMA, qLtEMA, pStEMA, qStEMA)
	pTrade4Asset[asset][0], pTrade4Asset[asset][1] = qLtEMA, qStEMA
	if signal_returned == UNDEF {
		return false, nil
	} else {
		signal := make([]interface{}, 3)
		signal[0] = trade[0] // time.Time
		signal[1] = asset
		signal[2] = signal_returned
		return true, signal
	} // }}}2
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
