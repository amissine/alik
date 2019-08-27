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

func main() { // {{{1
	log.Println(os.Getpid(), "gobble started")
	dec := json.NewDecoder(os.Stdin)
	w := bufio.NewWriter(os.Stdout)
	enc := json.NewEncoder(w)
	var p, q *aj.Umf
	for {
		var v aj.Umf
		if e := dec.Decode(&v); e != nil {
			log.Println("dec.Decode", e)
			break
		}
		if q = &v; q.Same(p) {
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
	log.Println(os.Getpid(), "gobble exiting..")
}
