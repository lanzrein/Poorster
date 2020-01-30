//main method for the client
//@authors Hrusanov Aleksandar, Lanzrein Johan, Rinaldi Vincent

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
	//Stuff for project...
	broadcast := flag.Bool("broadcast", false, "Broadcast flag")
	initcluster := flag.Bool("initcluster", false, "Initialize a new cluster at this node")
	joinOther := flag.String("joinOther", "", "Name of the peer that is the entry point to the cluster")
	expelOther := flag.String("expelOther", "", "Name of the peer to ban from the cluster")
	leavecluster := flag.Bool("leavecluster", false, "Leave the current cluster")

	// anonymous messaging
	anonymous := flag.Bool("anonymous", false, "Indicates sending an anonymous message")
	relayRate := flag.Float64("relayRate", 0.5, "Indicates probability of a peer relaying the message")
	fullAnonimity := flag.Bool("fullAnonimity", false, "If true, do not add Origin information in the encrypted packet")

	// audio communication
	call := flag.Bool("call", false, "Indicates sending a call request to a destination node")
	hangUp := flag.Bool("hangup", false, "Indicates hanging up the current call")
	startRecording := flag.Bool("startRecording", false, "Indicates beginning of audio recording")
	stopRecording := flag.Bool("stopRecording", false, "Indicates end of audio recording")

	// e-voting propositions
	accept := flag.String("accept", "", "Accept a node request")
	deny := flag.String("deny", "", "Deny a node request")

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
		if *call {
			// sending a call request
			log.Lvl3("Sending a call request")
			client.Call(dest)
		} else {
			//send a private message
			log.Lvl3("Sending private message ")
			client.SendPrivateMsg(*msg, *dest, anonymous, relayRate, fullAnonimity)
		}
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
		log.Lvl3("Init cluster")
		client.InitCluster()
	} else if *joinOther != "" {
		client.JoinCluster(joinOther)
	} else if *expelOther != "" {
		client.ExpelFromCluster(expelOther)
	} else if *leavecluster {
		client.LeaveCluster()
	} else if *hangUp {
		log.Lvl3("Hanging up")
		client.HangUp()
	} else if *startRecording {
		log.Lvl3("Recording and sending audio data")
		client.StartRecording()
	} else if *stopRecording {
		log.Lvl3("Stop recording audio data")
		client.StopRecording()
	} else if *accept != "" {
		client.ProposeAccept(accept)
	} else if *deny != "" {
		client.ProposeDeny(deny)
	} else {
		fmt.Print("ERROR (Bad argument combination)")
		os.Exit(1)
	}

	if err != nil {
		log.Error("Error : ", err, ".\n")
		os.Exit(1)
	}

}
