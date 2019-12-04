//file with function to handle server requests.

package gossiper

import (
	"encoding/hex"
	"encoding/json"
	"go.dedis.ch/onet/log"
	"io"
	"net"
	"net/http"
	"strings"
)

//LoadGossiper Try to load a new server. In cases of failure it will return without saying anything
//This is to prevent the cases where there is already a GUI server running on the same machine.
func LoadServer(g *Gossiper) {

	log.Lvl2("Loading server at address : 127.0.0.1:", (g.GUIPort))
	http.HandleFunc("/message", g.GetMessages)
	http.HandleFunc("/node", g.AddNode)
	http.HandleFunc("/id", g.GetId)
	//Hw2 handler
	http.HandleFunc("/privatemsg", g.PrivateMessageHandle)
	http.HandleFunc("/sharefile", g.FileSharingHandle)
	http.HandleFunc("/downloadfile", g.FileDownloadingHandle)
	//Hw3 handlers
	http.HandleFunc("/searchfile", g.FileSearchHandle)
	http.HandleFunc("/downloadfoundfile", g.FoundFileHandle)

	http.Handle("/", http.FileServer(http.Dir("./frontend")))
	for {

		err := http.ListenAndServe("127.0.0.1:"+(g.GUIPort), nil)

		if err != nil {
			if e, ok := err.(*net.OpError); ok {
				if e.Op == "listen" {
					log.Lvl2("Error could not start server : ", e)
					log.Lvl2("GUI port (80) is already used. This instance of Peerster will not use the GUI.")
					break

				}
			} else {
				log.Lvl2("Error on server : ", err)
			}

		}
	}

}

// GetId /id entry point. returns the id of the gossiper.
func (g *Gossiper) GetId(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	id := g.Name
	sending, _ := json.Marshal(id)
	_, _ = w.Write(sending)
}

//GetMessages /message entry point. will take care of the messages.
func (g *Gossiper) GetMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == "POST" {
		//get the data that was posted and convert it to a message

		data := make([]byte, r.ContentLength)
		cnt, err := r.Body.Read(data)
		if cnt != len(data) {
			return
		}

		if err != nil && err != io.EOF {
			log.Error("Error on sending message")
		}

		//data can be treated by the gossiper
		sm := SimpleMessage{
			OriginalName:  "",
			RelayPeerAddr: "",
			Contents:      string(data),
		}
		errChan := make(chan error)
		g.Receive(GossipPacket{Simple: &sm}, net.UDPAddr{}, errChan, true)

		tosend, err := g.ReplyToClient()
		if err != nil {
			return
		}

		_, err = w.Write(tosend)
		//fmt.Print("Printed " , n ," err : " , err, " \n ")

	} else if r.Method == "GET" {
		//get the update of the messages..
		data, err := g.ReplyToClient()
		if err != nil {
			return
		}

		_, err = w.Write(data)

	}
}

//AddNode /node entry point. handles requests for the known gossipers.
func (g *Gossiper) AddNode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	log.Lvl3("Got new request on node")
	if r.Method == "POST" {
		//get the data
		log.Lvl3("Got new post on node")
		data := make([]byte, r.ContentLength)
		cnt, err := r.Body.Read(data)
		if cnt != len(data) {
			return
		}
		log.Lvl3(" data : ", data)
		if err != nil && err != io.EOF {
			log.Error("Error on sending message")
		}

		//add the new peer.
		addr, err := net.ResolveUDPAddr("udp4", string(data))
		if err != nil {
			log.Error("Error : ", err, " \n ")
		}
		log.Lvl3("Adding new node : ", addr.String())
		g.AddNewPeer(*addr)
		kg := strings.Split(g.HostsToString(), ",")
		tosend, err := json.Marshal(kg)
		_, _ = w.Write(tosend)
	} else if r.Method == "GET" {
		kg := strings.Split(g.HostsToString(), ",")
		tosend, _ := json.Marshal(kg)
		_, _ = w.Write(tosend)
	}

}

type PrivMessage struct {
	Destination string
	Content     string
}

func (g *Gossiper) PrivateMessageHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "POST" {
		log.Lvl3("Got a private message")
		data := make([]byte, r.ContentLength)
		cnt, err := r.Body.Read(data)
		if cnt != len(data) {
			log.Lvl3("Could not read all data")
		}
		log.Lvl3("Data is : ", string(data))
		if err != nil && err != io.EOF {
			log.Error("Error on reading data : ", err)
		}
		msg := new(PrivMessage)
		err = json.Unmarshal(data, msg)
		if err != nil {
			log.Error("Could not unmarshal message : ", err)
		}
		log.Lvl3("Message is : ", msg)

		res := PrivateMessage{
			Origin:      g.Name,
			ID:          0,
			Text:        msg.Content,
			Destination: msg.Destination,
			HopLimit:    g.HopLimit,
		}
		log.Lvl3("Sending : ", res)
		err = g.SendPrivateMessage(res)
		if err != nil {
			log.Error("Could not send message : ", err)
		}

	}

	//get the name of all our origins.
	origins := g.GetOrigins()
	log.Lvl3("Origins : ", origins)
	tosend, _ := json.Marshal(origins)
	log.Lvl3(tosend)
	_, _ = w.Write(tosend)

}

//FileSharingHandle a handler for the file sharing. The GUI comes here to make request to upload a file
//The file is assumed to be in the _SharedFiles folder.
func (g *Gossiper) FileSharingHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == "POST" {
		log.Lvl3("Got a file sharing request")
		data := make([]byte, r.ContentLength)
		cnt, err := r.Body.Read(data)
		if cnt != len(data) {
			log.Lvl3("Could not read all data")
			return
		}
		log.Lvl3("Data is : ", string(data))
		if err != nil && err != io.EOF {
			log.Error("Error on reading data : ", err)
			return
		}

		log.Lvl3("File to download is : ", string(data))

		err = g.Index(string(data), pathShared)

		if err != nil {
			log.Error("Could not index file : ", err)
			return
		}
		_, _ = w.Write([]byte("OK"))
	}
}

type DownloadRequest struct {
	Filename    string
	Destination string
	Request     string
}

func (g *Gossiper) FileDownloadingHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "POST" {
		log.Lvl3("Got a downloading request ")
		data := make([]byte, r.ContentLength)
		cnt, err := r.Body.Read(data)
		if cnt != len(data) {
			log.Lvl3("Could not read all data")
		}
		log.Lvl3("Data is : ", string(data))
		if err != nil && err != io.EOF {
			log.Error("Error on reading data : ", err)
		}
		dr := new(DownloadRequest)
		err = json.Unmarshal(data, dr)
		if err != nil {
			log.Error("Could not unmarshal message : ", err)
		}
		bytes, err := hex.DecodeString(dr.Request)

		msg := Message{
			Text:        "",
			Destination: &dr.Destination,
			File:        &dr.Filename,
			Request:     &bytes,
		}
		g.StartFileDownload(msg)
		if err != nil {
			log.Error("Could not send message : ", err)
		}
	}
}

type FileSearch struct {
	Budget   int
	Keywords string
}

func (g *Gossiper) FileSearchHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "POST" {
		log.Lvl3("Got a file search request.")
		data := make([]byte, r.ContentLength)
		cnt, err := r.Body.Read(data)
		if cnt != len(data) {
			log.Lvl3("Could not read all data")
		}

		log.Lvl3("Data is : ", string(data))
		if err != nil && err != io.EOF {
			log.Error("Error on reading data : ", err)
		}
		filesearch := new(FileSearch)
		err = json.Unmarshal(data, filesearch)
		if err != nil {
			log.Error("Could not unmarshal message : ", err)
		}
		budget := uint64(filesearch.Budget)
		keywords := strings.Split(filesearch.Keywords, ",")
		mulFactor := 1
		//Start the file search
		err = g.StartFileSearch(keywords,budget, mulFactor)
		if err != nil {
			log.Error("Could not send message : ", err)
		}
	}else if r.Method == "GET"{
		log.Lvl2("Request for the found files..")
		data, err  := json.Marshal(g.FoundFiles)
		if err != nil{
			log.Lvl2("Error when marshalling found files ", err)
			return
		}

		_, _ = w.Write(data)
		return


	}
}
func (g *Gossiper) FoundFileHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "POST" {
		log.Lvl3("Got a downloading request for found file")
		data := make([]byte, r.ContentLength)
		cnt, err := r.Body.Read(data)
		if cnt != len(data) {
			log.Lvl3("Could not read all data")
		}
		log.Lvl3("Data is : ", string(data))
		if err != nil && err != io.EOF {
			log.Error("Error on reading data : ", err)
		}
		dr := new(DownloadRequest)
		err = json.Unmarshal(data, dr)
		if err != nil {
			log.Error("Could not unmarshal message : ", err)
		}
		bytes, err := hex.DecodeString(dr.Request)


		g.DownloadFoundFile(bytes,dr.Filename)
		if err != nil {
			log.Error("Could not send message : ", err)
		}
	}
}
