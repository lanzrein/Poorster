package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/JohanLanzrein/Peerster/gossiper"
	"go.dedis.ch/onet/log"
)

func main() {
	log.SetDebugVisible(gossiper.DEBUGLEVEL)
	log.Lvl3("Starting client")
	UIPort := flag.String("UIPort", "8080", "port for the UI client (default\"8080\"")
	msg := flag.String("msg", "", "message to be sent")
	dest := flag.String("dest", "", "destination for the private message; can be omitted.")
	file := flag.String("file", "", "file to be indexed by gossiper")
	request := flag.String("request", "", "request a chunk or metafile of this hash")
	keywords := flag.String("keywords", "", "Keywords for a file search")
	budget := flag.Int("budget", -1, "Budget for a file search")
	anonymous := flag.Bool("anonymous", false, "Indicates sending an anonymous message")
	relayRate := flag.Float64("relayRate", 0.5, "Indicates probability of a peer relaying the message")
	fullAnonimity := flag.Bool("fullAnonimity", false, "If true, do not add Origin information in the encrypted packet")
	//Stuff for project...
	broadcast := flag.Bool("broadcast", false, "Broadcast flag")
	initcluster := flag.Bool("initcluster", false, "Initialize a new cluster at this node")
	joinId := flag.Uint64("joinID", 0, "ID of the cluster the node wants to join")
	joinOther := flag.String("joinOther", "", "Name of the peer that is the entry point to the cluster")
	leavecluster := flag.Bool("leavecluster", false, "Leave the current cluster")

	flag.Parse()

	client := NewClient("127.0.0.1:" + *UIPort)
	var err error
	if *file != "" && *msg == "" {
		if *request != "" && *dest != "" {
			//file request
			log.Lvl3("Downloading a file ")
			err = client.RequestFile(file, dest, request, *anonymous, *relayRate)
		} else if *request == "" && *dest == "" {
			log.Lvl3("Indexing a file to the gossiper")
			err = client.SendFileToIndex(file)
		} else if *request != "" && *dest == "" {
			log.Lvl3("Downloading known file..")
			err = client.RequestFile(file, nil, request, *anonymous, *relayRate)
		} else {
			fmt.Print("ERROR (Bad argument combination)")
			os.Exit(1)
		}
	} else if (*dest != "") && (*request == "" || *file == "") {

		//send a private message
		log.Lvl3("Sending private message ")
		client.SendPrivateMsg(*msg, *dest)
	} else if (*request == "" && *file == "") && *msg != "" && !*broadcast {
		//rumor message
		log.Lvl3("rumor")
		err = client.SendMsg(*msg)
	} else if *keywords != "" {
		//Start a file search
		log.Lvl3("search")

		client.SearchFile(keywords, budget)
	} else if *broadcast && *msg != "" {
		log.Lvl3("Brodacst..")
		client.SendBroadcast(*msg)
	} else if *initcluster {
		client.InitCluster()
	} else if *joinId > 0 && *joinOther != "" {
		client.JoinCluster(joinId, joinOther)
	} else if *leavecluster {
		client.LeaveCluster()
	} else {
		fmt.Print("ERROR (Bad argument combination)")
		os.Exit(1)
	}

	if err != nil {
		log.Error("Error : ", err, ".\n")
		os.Exit(1)
	}

}
