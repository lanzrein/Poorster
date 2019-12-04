package main

import (
	"flag"
	"fmt"
	"github.com/JohanLanzrein/Peerster/gossiper"
	"go.dedis.ch/onet/log"
	"os"
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
	flag.Parse()
	client := NewClient("127.0.0.1:" + *UIPort)
	var err error
	if *file != "" && *msg == "" {
		if *request != "" && *dest != "" {
			//file request
			log.Lvl3("Downloading a file ")
			err = client.RequestFile(file, dest, request)

		} else if *request == "" && *dest == "" {
			log.Lvl3("Indexing a file to the gossiper")
			err = client.SendFileToIndex(file)
		} else if *request != "" && *dest == ""{
			log.Lvl3("Downloading known file..")
			err = client.RequestFile(file,nil,request)
		}else{
			fmt.Print("ERROR (Bad argument combination)")
			os.Exit(1)
		}
	} else if (*dest != "") && (*request == "" || *file == "") {

		//send a private message
		log.Lvl3("Sending private message ")
		client.SendPrivateMsg(*msg, *dest)
	} else if (*request == "" && *file == "") && *msg != "" {
		//rumor message
		log.Lvl3("rumor")
		err = client.SendMsg(*msg)
	} else if *keywords != ""{
		//Start a file search
		log.Lvl3("search")


		client.SearchFile(keywords, budget)
	} else {
		fmt.Print("ERROR (Bad argument combination)")
		os.Exit(1)
	}

	if err != nil {
		log.Error("Error : ", err, ".\n")
		os.Exit(1)
	}

}
