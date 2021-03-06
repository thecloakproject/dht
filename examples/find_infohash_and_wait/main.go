// Runs a node on UDP port 11221 that attempt to collect 100 peers for an
// infohash, then keeps running as a passive DHT node.
//
// To collect 100 peers, it usually has to contact some 10k nodes. This process
// is not instant and should take a minute or two, depending on your network
// connection.
//
// NOTE: The node has full debugging level enabled so you'll see a lot of output
// on the screen. It also runs a web server that can be used to collect
// debugging stats from http://localhost:8080/debug/vars.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	l4g "code.google.com/p/log4go"
	"github.com/nictuku/dht"
	"net/http"
)

const (
	httpPortTCP = 8080
	dhtPortUDP  = 11221
)

func main() {
	flag.Parse()
	// Change to l4g.DEBUG to see *lots* of debugging information.
	l4g.AddFilter("stdout", l4g.DEBUG, l4g.NewConsoleLogWriter())
	if len(flag.Args()) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: %v <infohash>\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}
	ih, err := dht.DecodeInfoHash(flag.Args()[0])
	if err != nil {
		l4g.Critical("DecodeInfoHash error: %v\n", err)
		os.Exit(1)
	}

	// This is a hint to the DHT of the minimum number of peers it will try to
	// find for the given node. This is not a reliable limit. In the future this
	// might be moved to "PeersRequest()", so the controlling client can have
	// different targets at different moments or for different infohashes.
	targetNumPeers := 5
	d, err := dht.NewDHTNode(dhtPortUDP, targetNumPeers, false)
	if err != nil {
		l4g.Critical("NewDHTNode error: %v", err)
		os.Exit(1)

	}
	go d.DoDHT()
	go drainresults(d)

	// Give the DHT some time to "warm-up" its routing table.
	time.Sleep(5 * time.Second)

	d.PeersRequest(string(ih), false)

	http.ListenAndServe(fmt.Sprintf(":%d", httpPortTCP), nil)
}

// drainresults loops, printing the address of nodes it has found.
func drainresults(n *dht.DHT) {
	fmt.Println("=========================== DHT")
	for r := range n.PeersRequestResults {
		for ih, peers := range r {
			l4g.Warn("Found peer(s) for infohash %x:", ih)
			for _, x := range peers {
				l4g.Warn("==========================> %v", dht.DecodePeerAddress(x))
				l4g.Warn("Note that there are many bad nodes that reply to anything you ask, so don't get too excited.")
				l4g.Warn("==========================")
			}
		}
	}
}
