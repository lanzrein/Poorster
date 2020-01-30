//main entry point of peerster.
package main

import (
	"flag"
	"os"

	"github.com/JohanLanzrein/Peerster/gossiper"
	"go.dedis.ch/onet/log"
)

func main() {
	//some parameters.
	log.SetDebugVisible(1)
	UIPort := flag.String("UIPort", "8080", "port for the UI client ")
	gossipAddr := flag.String("gossipAddr", "127.0.0.1:5000", "ip:port for the gossiper")

	name := flag.String("name", "", "name of g")
	peers := flag.String("peers", "", "comma separated list of peers of the form ip:port")
	simple := flag.Bool("simple", false, "run g in simple broadcast mode")
	antiEntropy := flag.Int("antiEntropy", 10, "Use the given timeout in seconds for anti-entropy. (default : 10)")
	GUIPort := flag.String("GUIPort", "8000", "port for the GUI default implementation expects default port")
	rtimer := flag.Int("rtimer", 0, "Timeout in seconds to send route rumors. Default means disable sending route rumors")
	//hw3 flags

	//hw3ex2 := flag.Bool("hw3ex2",false, "run exercice 2 of homework 3")
	N := flag.Int("N", -1, "Number of nodes in the network")
	stubbornTimeout := flag.Uint64("stubbornTimeout", 5, "Resend the transaction every stubbornTimeout seconds")
	//ex3
	hw3ex3 := flag.Bool("hw3ex3", false, "run exercice 3 of homework 3")
	hopLimit := flag.Uint64("hopLimit", 10, "Hop limit of packets")
	ackAll := flag.Bool("ackAll", false, "ack all the packets ( run hw3ex2 ) ")
	flag.Parse()
	//create a new gossiper

	if *name == "" {
		//no name was given - name is default name
		name = gossipAddr
	}

	log.Lvl3("Starting gossiper(", *name, ") at address : ", *gossipAddr, " with peers : ", *peers)
	g, err := gossiper.NewGossiper(*name, *UIPort, *gossipAddr, *peers, *simple, *antiEntropy, *GUIPort, *rtimer, *N, *stubbornTimeout, *ackAll, uint32(*hopLimit), *hw3ex3)
	defer Finish(g)
	if err != nil {
		log.Error("Error could not start gossiper : ", err)
		os.Exit(1)
	}

	err = g.Run()
	if err != nil {
		log.Error("Error while running the gossiper : ", err)
		os.Exit(1)
	}

}

func Finish(g gossiper.Gossiper) {
	log.Lvl3("Closing gossiper connections")
	err := g.Close()
	if err != nil {
		log.Warn("Could not properly close gossiper : ", err)
	}

}
