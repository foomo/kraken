package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"

	"github.com/foomo/kraken"
)

var flagAddress = flag.String("address", "127.0.0.1:8888", "where to listen too like 127.0.0.1:8888")
var flagInsecure = flag.Bool("insecure", false, "disabled by default - controls whether a client verifies the server's certificate chain and host name")

func main() {
	flag.Parse()
	fmt.Println("I AM KRAKEN at :", *flagAddress)
	if *flagInsecure {
		fmt.Println("!!! not verifying peers !!!")
		http.DefaultClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}
	k := kraken.NewKraken()
	kraken.NewServer(k).ListenAndServe(*flagAddress)
}
