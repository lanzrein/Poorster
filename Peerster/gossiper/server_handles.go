//@authors Hrusanov Aleksandar, Lanzrein Johan, Rinaldi Vincent

package gossiper

import (
	"encoding/hex"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"

	"go.dedis.ch/onet/log"
)

//Some structures..
type DownloadRequest struct {
	Filename    string
	Destination string
	Request     string
}
type FileSearch struct {
	Budget   int
	Keywords string
}
type PrivMessage struct {
	Destination string
	Content     string
	FullAnon    bool
	RelayRate   float64
}

type ClusterMD struct {
	JoinOther string
	ClusterID uint64
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

//PrivateMessageHandle a handler for private messages.
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

//FileDownloadingHandle a handle to download files form GUI
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

//FileSearchHandle a handle to launch a file search
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
		err = g.StartFileSearch(keywords, budget, mulFactor)
		if err != nil {
			log.Error("Could not send message : ", err)
		}
	} else if r.Method == "GET" {
		log.Lvl2("Request for the found files..")
		data, err := json.Marshal(g.FoundFiles)
		if err != nil {
			log.Lvl2("Error when marshalling found files ", err)
			return
		}

		_, _ = w.Write(data)
		return

	}
}

//FoundFileHandle download a previously found file.
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

		g.DownloadFoundFile(bytes, dr.Filename)
		if err != nil {
			log.Error("Could not send message : ", err)
		}
	}
}

func (g *Gossiper) InitClusterHandle(w http.ResponseWriter, r *http.Request) {
	log.Lvl3(g.Name, "Received init request...")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "POST" {
		data := make([]byte, r.ContentLength)
		cnt, err := r.Body.Read(data)
		if cnt != len(data) {
			log.Lvl3("Could not read all data")
		}

		log.Lvl3("Data is : ", string(data))
		if err != nil && err != io.EOF {
			log.Error("Error on reading data : ", err)
		}
		g.InitCluster()

	}
}

func (g *Gossiper) JoinClusterRequest(w http.ResponseWriter, r *http.Request) {
	log.Lvl2(g.Name, "received join request.")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "POST" {
		data := make([]byte, r.ContentLength)
		cnt, err := r.Body.Read(data)
		if cnt != len(data) {
			log.Lvl3("Could not read all data")
		}

		log.Lvl3("Data is : ", string(data))
		if err != nil && err != io.EOF {
			log.Error("Error on reading data : ", err)
		}

		clusterMD := new(ClusterMD)
		err = json.Unmarshal(data, clusterMD)
		if err != nil {
			log.Error("Could not unmarshal message : ", err)
		}
		g.RequestJoining(clusterMD.JoinOther)
	}
}

func (g *Gossiper) LeaveClusterHandle(w http.ResponseWriter, r *http.Request) {
	log.Lvl2(g.Name, "leaving cluster!")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "POST" {
		data := make([]byte, r.ContentLength)
		cnt, err := r.Body.Read(data)
		if cnt != len(data) {
			log.Lvl3("Could not read all data")
		}

		log.Lvl3("Data is : ", string(data))
		if err != nil && err != io.EOF {
			log.Error("Error on reading data : ", err)
		}

		g.LeaveCluster()
	}
}

func (g *Gossiper) BroadcastMessageHandle(w http.ResponseWriter, r *http.Request) {
	log.Lvl2("Broadcast message handle")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "POST" {
		data := make([]byte, r.ContentLength)
		cnt, err := r.Body.Read(data)
		if cnt != len(data) {
			log.Lvl3("Could not read all data")
		}

		log.Lvl3("Data is : ", string(data))
		if err != nil && err != io.EOF {
			log.Error("Error on reading data : ", err)
		}

		message := new(PrivMessage)
		err = json.Unmarshal(data, message)
		if err != nil {
			log.Error("Could not unmarshal message : ", err)
		}
		log.Lvl2("Data : ", message)
		g.SendBroadcast(message.Content, false)

	}

	replyClusterMembers(g, w)
}

func replyClusterMembers(g *Gossiper, w http.ResponseWriter) {
	//Reply the memebrs..
	if g.Cluster == nil {
		return
	}
	origins := g.Cluster.Members
	log.Lvl3("Members : ", origins)
	tosend, _ := json.Marshal(origins)
	log.Lvl3(tosend)
	_, _ = w.Write(tosend)
}

func (g *Gossiper) UpdateClusterMembersHandle(w http.ResponseWriter, r *http.Request) {

}

func (g *Gossiper) AnonymousMessageHandle(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "POST" {
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
		log.Lvl2("Data : ", msg)
		//Todo here you do the anonmessage handling for peerster...
		g.ClientSendAnonymousMessage(msg.Destination, msg.Content, msg.RelayRate, msg.FullAnon)

	}

	//at the end update with the current cluster members.
	replyClusterMembers(g, w)
}

type EVote struct {
	Vote     string // JOIN ou EXPEL
	Person   string // person
	Decision bool   // true - yes, false no
}

func (g *Gossiper) EvotingHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "POST" {
		data := make([]byte, r.ContentLength)
		cnt, err := r.Body.Read(data)
		if cnt != len(data) {
			log.Lvl3("Could not read all data")
		}

		log.Lvl3("Data is : ", string(data))
		if err != nil && err != io.EOF {
			log.Error("Error on reading data : ", err)
		}

		message := new(EVote)
		err = json.Unmarshal(data, message)
		if err != nil {
			log.Error("Could not unmarshal message : ", err)
		}
		log.Lvl2("Data : ", message)
		if message.Decision {
			log.Lvl2("Message accept proposition.")
			for i := 0; i < len(g.displayed_requests); i++ {
				if strings.Contains(g.displayed_requests[i], message.Person) {
					copy(g.displayed_requests[i:], g.displayed_requests[i+1:])
					g.displayed_requests = g.displayed_requests[:len(g.displayed_requests)-1]
					break
				}
			}
			g.BroadcastAccept(message.Person)
		} else {
			log.Lvl2("Message deny proposition.")
			for i := 0; i < len(g.displayed_requests); i++ {
				if strings.Contains(g.displayed_requests[i], message.Person) {
					copy(g.displayed_requests[i:], g.displayed_requests[i+1:])
					g.displayed_requests = g.displayed_requests[:len(g.displayed_requests)-1]
					break
				}
			}
			g.BroadcastDeny(message.Person)
		}

	}

	votes := g.displayed_requests
	log.Lvl3("Voting ongoing.  : ", votes)

	tosend, _ := json.Marshal(votes)
	log.Lvl3(tosend)
	_, _ = w.Write(tosend)
}

func (g *Gossiper) ExpellMemberHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "POST" {
		data := make([]byte, r.ContentLength)
		cnt, err := r.Body.Read(data)
		if cnt != len(data) {
			log.Lvl3("Could not read all data")
		}

		log.Lvl3("Data is : ", string(data))
		if err != nil && err != io.EOF {
			log.Error("Error on reading data : ", err)
		}

		message := new(PrivMessage)
		err = json.Unmarshal(data, message)
		if err != nil {
			log.Error("Could not unmarshal message : ", err)
		}
		log.Lvl2("Data expell: ", message.Destination)
		//Todo here you do the evoting handling for expelling...

		//str := "EXPEL "+message.Destination
		g.RequestExpelling(message.Destination)

	}
}

type CallData struct {
	Dial    bool
	Accept  bool
	Decline bool
	Hangup  bool
	Member  string
}

func (g *Gossiper) CallHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "POST" {
		data := make([]byte, r.ContentLength)
		cnt, err := r.Body.Read(data)
		if cnt != len(data) {
			log.Lvl3("Could not read all data")
		}

		log.Lvl3("Data is : ", string(data))
		if err != nil && err != io.EOF {
			log.Error("Error on reading data : ", err)
		}

		message := new(CallData)
		err = json.Unmarshal(data, message)
		if err != nil {
			log.Error("Could not unmarshal message : ", err)
		}
		log.Lvl2("Data : ", message)

		//TODO here handle the packet.
		otherParticipant := message.Member
		if len(otherParticipant) > 0 {
			if otherParticipant[0] == '"' {
				otherParticipant = otherParticipant[1:]
			}
			if otherParticipant[len(otherParticipant)-1] == '"' {
				otherParticipant = otherParticipant[:len(otherParticipant)-1]
			}
		}

		if message.Dial {
			// we are calling
			g.ClientSendCallRequest(otherParticipant)
		} else if message.Accept {
			resp := CallResponse{Origin: g.Name, Destination: otherParticipant, Status: Accept}
			// we are accepting an incoming call
			g.SendCallResponse(resp)
		} else if message.Decline {
			// we are declining an incoming call
			resp := CallResponse{Origin: g.Name, Destination: otherParticipant, Status: Decline}
			// we are accepting an incoming call
			g.SendCallResponse(resp)
		} else if message.Hangup {
			// we are hanging up
			g.ClientSendHangUpMessage()
		}

	}

}

func (g *Gossiper) IncomingCallHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	//TODO here handle the packet.
	if g.CallStatus.OtherParticipant != "" {
		caller := g.CallStatus.OtherParticipant
		tosend, _ := json.Marshal(caller)
		log.Lvl3(tosend)
		_, _ = w.Write(tosend)
	}

}
