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
	log.Println(os.Getpid(), os.Args[1], os.Args[2], "feed started")
	dec := json.NewDecoder(os.Stdin)
	w := bufio.NewWriter(os.Stdout)
	enc := json.NewEncoder(w)
	var p, q *aj.Umf
	for {
		var v map[string]interface{}
		if e := dec.Decode(&v); e != nil {
			log.Println(os.Getpid(), "dec.Decode", e)
			break
		}
		if q = q.Init(&v); q.Same(p) {
			log.Println("skipping", *q)
			continue
		}
		p = q
		if e := enc.Encode(q); e != nil {
			log.Println("enc.Encode", e)
			break
		}
		w.Flush()
	}
	log.Println(os.Getpid(), "feed exiting...")
}
