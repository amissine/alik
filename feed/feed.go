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
	"os/exec"
	"strings"
)

// see also:
// - https://golang.org/doc/code.html

func moreTrades(asset, feeds string, enc *json.Encoder) { // {{{1
	for _, feed := range strings.Fields(feeds) {
		if feed == "kraken" { // TODO implement {{{2
			continue
		}
		a := asset
		if feed == "coinbase" {
			a += "-"
		}
		cmd := exec.Command("./feed.sh", feed, a+"USD")
		pipe, err := cmd.StdoutPipe()
		if err != nil {
			log.Fatal("moreTrades 1 ", err)
		}
		if err := cmd.Start(); err != nil {
			log.Fatal("moreTrades 2 ", err)
		} // }}}2
		var v []interface{}
		var buf bytes.Buffer
		tee := io.TeeReader(bufio.NewReaderSize(pipe, 16384), &buf)
		if err := json.NewDecoder(tee).Decode(&v); err != nil { // {{{2
			log.Println(feed, asset, err)
			v = nil
			b, e := ioutil.ReadAll(&buf)
			if e != nil {
				log.Fatal(e)
			}
			log.Printf("===\n%s\n===\n", b)
		}
		if err := cmd.Wait(); err != nil {
			log.Fatal("moreTrades 4 ", err)
		}
		if v == nil { // err != nil
			continue
		}

		var q *aj.UMF // {{{2
		var skip bool
		for _, trade := range v {
			if skip, q = q.MakeTrade(feed, asset, trade.(map[string]interface{})).Skip(); skip {
				continue
			}
			if e := enc.Encode(q); e != nil {
				log.Println(os.Getpid(), "moreTrades 5 ", e)
			}
		}
	}
}

func trades(q *aj.UMF, asset, feeds, tp string, enc *json.Encoder) { // {{{1
	if q == nil || q.IsTrade() {
		return
	}
	if asset == "USD" {
		asset = "XLM"
	}
	for _, pair := range strings.Fields(tp) {
		if strings.HasPrefix(pair, asset) {
			moreTrades(asset, feeds, enc)
			break
		}
	}
}

func main() { // {{{1
	asset := os.Args[1]
	feeds := os.Getenv("FEEDS")
	tradingPairs := os.Getenv("TRADING_PAIRS")
	log.Println(os.Getpid(), os.Getenv("SDEX_FEED_STARTED"), asset, "; feeds:", feeds, ", tradingPairs:", tradingPairs)
	dec := json.NewDecoder(bufio.NewReaderSize(os.Stdin, 16384))
	w := bufio.NewWriterSize(os.Stdout, 65536)
	enc := json.NewEncoder(w)
	var q *aj.UMF
	var skip bool
	for {
		var v map[string]interface{}
		if e := dec.Decode(&v); e != nil {
			log.Println(os.Getpid(), "dec.Decode", e)
			break
		}
		if skip, q = q.MakeSDEX(asset, &v).Skip(); skip {
			trades(q, asset, feeds, tradingPairs, enc)
			continue
		}
		trades(q, asset, feeds, tradingPairs, enc)
		if e := enc.Encode(q); e != nil {
			log.Println(os.Getpid(), "enc.Encode", e)
			break
		}
		w.Flush()
	}
	log.Println(os.Getpid(), "feed exiting...")
}
