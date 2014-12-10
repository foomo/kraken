package main

import (
	"flag"
	"github.com/foomo/kraken"
	"log"
)

var flagAddress = flag.String("address", "127.0.0.1:8888", "where to listen too like 127.0.0.1:8888")

func main() {
	flag.Parse()
	log.Println("I AM KRAKEN at :", *flagAddress)
	k := kraken.NewKraken()
	kraken.NewServer(k).ListenAndServe(*flagAddress)
}
