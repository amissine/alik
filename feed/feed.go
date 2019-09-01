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
	var q *aj.Umf
	for {
		var v map[string]interface{}
		if e := dec.Decode(&v); e != nil {
			log.Println(os.Getpid(), "dec.Decode", e)
			break
		}
		if q = q.Init(&v); q.Skip() {
			log.Println(os.Getpid(), "skipping", *q)
			continue
		}
		if e := enc.Encode(&v); e != nil {
			log.Println(os.Getpid(), "enc.Encode", e)
			break
		}
		w.Flush()
	}
	log.Println(os.Getpid(), "feed exiting...")
}
