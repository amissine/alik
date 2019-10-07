package main

// import {{{1
import (
	"bufio"
	"encoding/json"
	aj "github.com/amissine/alik/json"
	"log"
	"os"
	"os/exec"
	"strings"
)

// see also:
// - https://golang.org/doc/code.html

func moreTrades(asset, feeds string, enc *json.Encoder) { // {{{1
	for _, feed := range strings.Fields(feeds) {
		if feed == "kraken" { // TODO implement
			continue
		}
		if feed == "coinbase" {
			asset += "-"
		}
		tp := asset + "USD"
		cmd := exec.Command("./feed.sh", feed, tp)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Fatal("ERROR", err)
		}
		if err := cmd.Start(); err != nil {
			log.Fatal(err)
		}
		var v map[string]interface{}
		if err := json.NewDecoder(stdout).Decode(&v); err != nil {
			log.Fatal(err)
		}
		if err := cmd.Wait(); err != nil {
			log.Fatal(err)
		}
		log.Println(os.Getpid(), "moreTrades", v)
		if e := enc.Encode(&v); e != nil {
			log.Println(os.Getpid(), "moreTrades", e)
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
		log.Println(os.Getpid(), asset, feed)
		return
	}
	feeds := os.Args[3]
	tradingPairs := os.Args[4]
	log.Println(os.Getpid(), asset, "sdex feed started")
	dec := json.NewDecoder(os.Stdin)
	w := bufio.NewWriter(os.Stdout)
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
