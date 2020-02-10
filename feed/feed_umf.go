package main

// import {{{1
import (
	"bufio"
	"bytes"
	"encoding/json"
	aj "github.com/amissine/alik/json"
	"io"
	"io/ioutil"
	"log"
	"os"
)

// See also:
// - https://github.com/fatih/vim-go/wiki/Tutorial
// - https://play.golang.org/p/X20RtgpEzqL

func main() { // {{{1
	ToUMF(os.Args[1], os.Args[2])
}

var ob = func() *func(string) string {
	orderbooksToUMF := func(asset string) string {
		return asset
	}
	return &orderbooksToUMF
}

func ToUMF(asset, feed string) { // {{{1
	log.Println(os.Getpid(), os.Getenv("SDEX_FEED_STARTED"), asset, feed)

	var v interface{}
	var b bytes.Buffer
	tee := io.TeeReader(bufio.NewReaderSize(os.Stdin, 16384), &b)
	for {
		if err := json.NewDecoder(tee).Decode(&v); err != nil { // {{{2
			b, err := ioutil.ReadAll(&b)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("===\n%s\n===\n", b)
			continue
		}
		switch v.(type) { // {{{2
		case []interface{}:
			array := v.([]interface{})
			switch feed {
			case "bitfinex_t":
				aj.BitfinexTradesToUMF(asset, &array)
			case "coinbase_t":
				aj.CoinbaseTradesToUMF(asset, &array)
			default:
				panic("FIXME")
			}
		case map[string]interface{}:
			object := v.(map[string]interface{})
			switch feed {
			case "sdex_ob":
				aj.SdexOrderbookToUMF(asset, &object)
			case "sdex_t":
				aj.SdexTradeToUMF(asset, &object)
			case "kraken_t":
				aj.KrakenTradesToUMF(asset, &object)
			default:
				panic("FIXME")
			}
		default:
			panic("FIXME")
		}
	}
}
