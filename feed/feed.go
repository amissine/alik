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
		}
		var v []interface{}
		var buf bytes.Buffer
		tee := io.TeeReader(bufio.NewReaderSize(pipe, 16384), &buf)
		if err := json.NewDecoder(tee).Decode(&v); err != nil {
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
		if v == nil {
			continue
		}

		var q *aj.Umf // {{{2
		if q = q.MakeTrade(feed, asset, &v); q.SkipTrade() {
			continue
		}
		if e := enc.Encode(q); e != nil {
			log.Println(os.Getpid(), "moreTrades 5 ", e)
		}
	}
}

func trades(asset, feeds, tradingPairs string, enc *json.Encoder) { // {{{1
	if asset == "USD" {
		asset = "XLM"
	}
	for _, tp := range strings.Fields(tradingPairs) {
		if strings.HasPrefix(tp, asset) {
			moreTrades(asset, feeds, enc)
			break
		}
	}
}

func main() { // {{{1
	feed := os.Args[1]
	asset := os.Args[2]
	if feed != "sdex" {
		log.Println(os.Getpid(), asset, feed, "- must be sdex")
		return
	}
	feeds := os.Getenv("FEEDS")
	tradingPairs := os.Getenv("TRADING_PAIRS")
	log.Println(os.Getpid(), feed, asset, "; feeds:", feeds, ", tradingPairs:", tradingPairs)
	dec := json.NewDecoder(bufio.NewReaderSize(os.Stdin, 16384))
	w := bufio.NewWriterSize(os.Stdout, 65536)
	enc := json.NewEncoder(w)
	var q *aj.Umf
	for {
		var v map[string]interface{}
		if e := dec.Decode(&v); e != nil {
			log.Println(os.Getpid(), "dec.Decode", e)
			break
		}
		if q = q.Make("sdex", asset, &v); q.Skip() {
			if !q.UTC.IsZero() {
				q = nil
			}
			continue
		}
		trades(asset, feeds, tradingPairs, enc)
		if e := enc.Encode(q); e != nil {
			log.Println(os.Getpid(), "enc.Encode", e)
			break
		}
		w.Flush()
		q = nil
	}
	log.Println(os.Getpid(), "feed exiting...")
}
