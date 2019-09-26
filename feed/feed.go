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

func main() { // {{{1
	exchange := os.Args[1]
	asset := os.Args[2]
	log.Println(os.Getpid(), exchange, asset, "feed started")
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
		if q = q.Make(exchange, asset, &v); q.Skip() {
			//log.Println(os.Getpid(), "skipping", *q)
			if !q.UTC.IsZero() {
				q = nil
			}
			continue
		}
		if e := enc.Encode(q); e != nil {
			log.Println(os.Getpid(), "enc.Encode", e)
			break
		}
		w.Flush()
		q = nil
	}
	log.Println(os.Getpid(), "feed exiting...")
}
