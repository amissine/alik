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
// - https://play.golang.org/p/ftgPNp31_j4

func main() { // {{{1
	asset := os.Args[1]
	log.Println(os.Getpid(), "- command feed started for asset", asset)
	for {
		ToUMF(asset)
	}
}

func ToUMF(asset string) { // {{{1
	var b bytes.Buffer
	tee := io.TeeReader(bufio.NewReaderSize(os.Stdin, 16384), &b)
	v := decode(tee, &b)
	switch v.(type) {
	case []interface{}: // {{{2
		array := v.([]interface{})
		switch feed_array(&array) {
		case "bitfinex_t":
			aj.BitfinexTradesToUMF(asset, &array)
		case "coinbase_t":
			aj.CoinbaseTradesToUMF(asset, &array)
		default:
			log.Println(os.Getpid(), "- array", array)
		}
	case map[string]interface{}: // {{{2
		object := v.(map[string]interface{})
		switch feed_object(&object) {
		case "sdex_ob":
			aj.SdexOrderbookToUMF(asset, &object)
		case "sdex_t":
			aj.SdexTradeToUMF(asset, &object)
		case "kraken_t":
			aj.KrakenTradesToUMF(asset, &object)
		default:
			log.Println(os.Getpid(), "- object", object)
		}
	default: // {{{2
		if v != nil {
			log.Fatal(os.Getpid(), " - FIXME v ", v)
		}
	} // }}}2
}

func feed_array(pa *[]interface{}) string { // {{{1
	a := *pa
	m, yes := a[0].(map[string]interface{})
	if yes {
		if _, ok := m["price"]; ok {
			return "coinbase_t"
		} else {
			return "unknown"
		}
	}
	a4, si := a[0].([]interface{})
	if si && len(a4) == 4 {
		return "bitfinex_t"
	}
	return "unknown"
}

func feed_object(po *map[string]interface{}) string { // {{{1
	o := *po
	if asks, ok := o["asks"]; ok { // sdex_ob {{{2
		_, yes := asks.([]interface{})
		if yes {
			return "sdex_ob"
		} else {
			return "unknown"
		}
	}
	if links, ok := o["_links"]; ok { // sdex_t {{{2
		_, yes := links.(map[string]interface{})
		if yes {
			return "sdex_t"
		} else {
			return "unknown"
		}
	}
	if result, ok := o["result"]; ok { // kraken_t {{{2
		_, yes := result.(map[string]interface{})
		if yes {
			return "kraken_t"
		} else {
			return "unknown"
		}
	} // }}}2
	return "unknown"
}

func decode(r io.Reader, pb *bytes.Buffer) interface{} { // {{{1
	var v interface{}
	if err := json.NewDecoder(r).Decode(&v); err != nil {
		if err.Error() == "EOF" {
			log.Println(os.Getpid(), "- exiting on EOF")
			os.Exit(0)
		}
		b, err2 := ioutil.ReadAll(pb)
		if err2 != nil {
			log.Fatal(err2)
		}
		log.Printf("%d ===\n%s\n===\n%s\n", os.Getpid(), b, err)
		return nil
	}
	return v
}
