//@authors Hrusanov Aleksandar, Lanzrein Johan, Rinaldi Vincent

//Main file for the gossiper most handles for receiving/sending messages are here.
package gossiper

import (
	"bufio"
	"bytes"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/JohanLanzrein/Peerster/ies"
	"go.dedis.ch/onet/log"
	"go.dedis.ch/protobuf"
)

//NewGossiper Create a new gossiper given the parameter.
//This will not start the gossiper just create the objet. to run it call g.Run()
func NewGossiper(Name string, UIPort string, gossipAddr string, gossipers string, simple bool, antiEntropy int, GUIPort string, rtimer int, N int, stubbornTimeout uint64, ackAll bool, hopLimit uint32, hw3ex3 bool) (Gossiper, error) {

	log.Lvl3("GUIPort : ", GUIPort)
	hosts, err := ParseKnownHosts(gossipers)

	if err != nil {
		return Gossiper{}, err
	}
	if rtimer < 0 {
		rtimer = 0
	}
	if antiEntropy < 0 {
		antiEntropy = 0
	}

	//setting up the local listener..
	udpAddr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:"+UIPort)
	if err != nil {
		return Gossiper{}, err
	}
	udpConnClient, err := net.ListenUDP("udp4", udpAddr)

	//setting up gossip listener
	udpAddr, err = net.ResolveUDPAddr("udp4", gossipAddr)
	if err != nil {
		return Gossiper{}, err
	}
	udpConnRemote, err := net.ListenUDP("udp4", udpAddr)
	if err != nil {
		return Gossiper{}, err
	}

	//the atomic structures to set up
	kg := KnownGossiper{mu: sync.Mutex{}, values: &hosts}
	pl := PeerLog{mu: sync.Mutex{}, logmap: make(map[string][]GossipPacket)}
	routing := RoutingTable{mu: sync.Mutex{}, table: make(map[string]struct {
		string
		int
	})}
	metalock := MetaLock{mu: sync.Mutex{}, data: []*MetaData{}}

	buf := bytes.Buffer{}
	Map1 := CurrentlyDownloading{mu: sync.Mutex{}, values: make(map[chan DataReply][]byte)}
	wl := WaitingList{mu: sync.Mutex{}, list: make(map[string]RumorMessage)}
	//Init stuff for hw3
	RecentlySeen := RecentlySeenRequest{
		Mutex:  sync.Mutex{},
		values: []SearchRequest{},
	}
	CurrentlySearching := CurrentlySearching{
		Mutex:  sync.Mutex{},
		values: []FoundFiles{},
	}
	SearchMatches := make(map[string]uint64)
	Chunks := make(map[string][]byte)
	TLCMessages := make(map[string]*TLCMessage)

	//for project...
	keys := make(map[string]ies.PublicKey)
	call := GossiperCallStatus{
		InCall:            false,
		ExpectingResponse: false,
	}

	gossiper := Gossiper{
		Name:                 Name,
		UIPort:               UIPort,
		connClient:           *udpConnClient,
		connNode:             *udpConnRemote,
		gossipAddr:           udpAddr,
		KnownGossipers:       kg,
		antiEntropy:          time.Duration(antiEntropy) * time.Second,
		SimpleMode:           simple,
		wl:                   wl,
		pl:                   pl,
		counter:              1,
		buffer:               &buf,
		GUIPort:              GUIPort,
		routing:              routing,
		rtimer:               uint(rtimer),
		metalock:             metalock,
		CurrentlyDownloading: Map1,
		//hw3
		CurrentlySearching: CurrentlySearching,
		RecentlySeen:       RecentlySeen,
		SearchMatches:      SearchMatches,
		Chunks:             Chunks,

		N:                N,
		stubbornTimeout:  stubbornTimeout,
		ackAll:           ackAll,
		TLCMessages:      TLCMessages,
		Acknowledgements: make(map[uint32][]string),

		hw3ex3:              hw3ex3,
		HopLimit:            hopLimit,
		ConfirmedMessages:   ConcTLCMessages{sync.Mutex{}, make(map[uint32][]*TLCMessage)},
		FutureMessages:      ConcTLCMessages{sync.Mutex{}, make(map[uint32][]*TLCMessage)},
		RunningConfirmation: false,
		TimeMapping:         TimeMapping{sync.Mutex{}, make(map[string][]uint32)},

		Keys:          keys,
		RolloutTimer:  DEFAULTROLLOUT,
		HearbeatTimer: DEFAULTHEARTBEAT,
		LeaveChan:     make(chan bool),
		CallStatus:    call,

		slice_results:                 make([][]string, 0),
		acks_cases:                    make(map[string][]string),
		correct_results_rcv:           make(map[string][]string),
		reset_requests:                make(map[string][]string),
		members_ready_resend_requests: make(map[string][]string),
		pending_nodes_requests:        make([]string, 0),
		pending_messages_requests:     make([]RequestMessage, 0),
		displayed_requests:            make([]string, 0),
	}

	err = gossiper.GenerateKeys()

	return gossiper, err
}

//Run The main method of the gossiper. Call it to run the gossiper.
func (g *Gossiper) Run() error {
	//the error chan to verify if we get some errors from the different go routines.

	errChan := make(chan error)

	//if it is not in simple mode *try* to load a server - if it fails ( i.e. if a server is already running then it will just return without saying anything ! )
	if !g.SimpleMode {
		go LoadServer(g)
	}

	//read from client connection
	go g.ReadFromPort(errChan, g.connClient, true)
	//read from gossiper connection
	go g.ReadFromPort(errChan, g.connNode, false)

	if g.SimpleMode {
		//just check for errors..
		for {

			err := <-errChan
			if err != nil {
				return err
			}
		}

	} else {
		if g.rtimer > 0 {
			go g.StartupRouting()
			go g.RouteLoop()
		}

		if g.antiEntropy > 0 {

			//setup a loop to check for errors
			go func() {
				for {
					err := <-errChan
					if err != nil {
						log.Error("Encountered error : ", err)
					}
				}

			}()

			//Loop for anti-entropy
			for {

				//Wait for a certain amount of seconds.
				log.Lvl3("Antientropy sleeping")
				<-time.After(g.antiEntropy)

				log.Lvl3("Antientropy wakeup")
				g.KnownGossipers.mu.Lock()
				if len(*g.KnownGossipers.values) == 0 {
					g.KnownGossipers.mu.Unlock()
					continue
				}

				rand.Seed(int64(time.Now().Nanosecond()))
				i := rand.Intn(len(*g.KnownGossipers.values))
				addr := (*g.KnownGossipers.values)[i]
				g.KnownGossipers.mu.Unlock()
				log.Lvl4("Anti entropy sending to : ", addr, " i = ", i)
				go g.SendStatusPacket(addr, errChan)

				//we may get a confirmed tlc in return because of the vector clock !
				xs, _ := g.ConfirmedMessages.values[g.my_time]
				if len(xs) > 0 {
					log.Lvl2("Sending a confirmed tlc message..")

					tlc := xs[0]
					gp := GossipPacket{TLCMessage: tlc}
					go g.SendToRandom(gp)
				}

			}
		} else {
			//setup a loop to check for errors
			func() {

				for {
					err := <-errChan
					if err != nil {
						log.Error("Encountered error : ", err)
					}
				}

			}()

		}

	}

	//unreachable code -
	return nil
}

//ReadFromPort loops infinetly on a connection and reads from it...
func (g *Gossiper) ReadFromPort(errChan chan error, conn net.UDPConn, client bool) {
	for {
		packetBytes := make([]byte, 10000)
		cnt, sendingAddr, err := conn.ReadFromUDP(packetBytes)
		if err != nil {
			errChan <- err
		}
		packetBytes = packetBytes[:cnt]

		packet := GossipPacket{}
		anonymous := false
		audio := false
		//if its a client then decode the message and wrap it in gossip packet
		if client {
			log.Lvl3("Got message from client.")
			msg := Message{}
			err = protobuf.Decode(packetBytes, &msg)
			if err != nil {
				log.Error("error on decoding client message : ", err)
				continue
			}
			if msg.Destination != nil && msg.Request != nil && msg.File != nil {
				//its a request for a file download.
				log.Lvl2("Starting a file dl")
				go g.StartFileDownload(msg)
				continue
			} else if msg.Request != nil && msg.File != nil {
				log.Lvl2("Got a download request for a found file")
				go g.DownloadFoundFile(*msg.Request, *msg.File)
				continue
			} else if msg.File != nil {
				//its a request to index a file.
				go g.Index(*msg.File, pathShared)
				continue
			} else if msg.Destination != nil && (msg.Anonymous == nil || !(*msg.Anonymous) || msg.CallRequest == nil || !(*msg.CallRequest)) {
				if msg.Anonymous != nil && *msg.Anonymous {
					log.Lvl2("Message to send anon messaging...")
					anonymous = true
					g.ClientSendAnonymousMessage(*msg.Destination, msg.Text, *msg.RelayRate, *msg.FullAnonimity)
				} else if msg.CallRequest != nil && *msg.CallRequest {
					log.Lvl2("Message to call someone...")
					audio = true
					g.ClientSendCallRequest(*msg.Destination)
				} else {
					packet = GossipPacket{Private: &PrivateMessage{
						Origin:      g.Name,
						ID:          0,
						Text:        msg.Text,
						Destination: *msg.Destination,
						HopLimit:    g.HopLimit,
					}}
				}
			} else if msg.Keywords != nil {
				mulFactor := 1
				if msg.Budget == nil {
					msg.Budget = new(uint64)
					*msg.Budget = INITBUDGET
					mulFactor = 2
				}
				log.Lvl2("Budgetz is : ", *msg.Budget)

				log.Lvl3("Starting file search !")

				go g.StartFileSearch(*msg.Keywords, *msg.Budget, mulFactor)
				continue
			} else if msg.Broadcast != nil && *msg.Broadcast && msg.Text != "" {
				log.Lvl2("Broadcast message...")
				go g.SendBroadcast(msg.Text, false)
				continue
			} else if msg.InitCluster != nil {
				log.Lvl2("Message to init the cluster..")
				g.InitCluster()
				continue
			} else if msg.JoinOther != nil {
				log.Lvl2("Message joining request.")
				g.RequestJoining(*msg.JoinOther)
				continue
			} else if msg.ExpelOther != nil {
				log.Lvl1("Message expel request.")
				if g.Cluster.AmountAuthorities() > 1 && len(g.Cluster.Members) > 2 {
					g.RequestExpelling(*msg.ExpelOther)
				}
				continue
			} else if msg.LeaveCluster != nil && *msg.LeaveCluster {
				g.LeaveCluster()
				continue
			} else if msg.HangUp != nil && *msg.HangUp {
				log.Lvl2("Message to hang up...")
				audio = true
				g.ClientSendHangUpMessage()
			} else if msg.StartRecording != nil && *msg.StartRecording {
				log.Lvl2("Message to record and send audio ...")
				audio = true
				g.ClientStartRecording()
			} else if msg.StopRecording != nil && *msg.StopRecording {
				log.Lvl2("Message to record and send audio ...")
				audio = true
				g.ClientStopRecording()
			} else if msg.PropAccept != nil {
				log.Lvl2("Message accept proposition.")
				for i := 0; i < len(g.displayed_requests); i++ {
					if strings.Contains(g.displayed_requests[i], *msg.PropAccept) {
						copy(g.displayed_requests[i:], g.displayed_requests[i+1:])
						g.displayed_requests = g.displayed_requests[:len(g.displayed_requests)-1]
						break
					}
				}
				g.BroadcastAccept(*msg.PropAccept)
				continue
			} else if msg.PropDeny != nil {
				log.Lvl2("Message deny proposition.")
				for i := 0; i < len(g.displayed_requests); i++ {
					if strings.Contains(g.displayed_requests[i], *msg.PropDeny) {
						copy(g.displayed_requests[i:], g.displayed_requests[i+1:])
						g.displayed_requests = g.displayed_requests[:len(g.displayed_requests)-1]
						break
					}
				}
				g.BroadcastDeny(*msg.PropDeny)
				continue
			} else {
				packet = GossipPacket{Simple: &SimpleMessage{
					OriginalName:  "",
					RelayPeerAddr: "",
					Contents:      msg.Text,
				}}
			}
		} else {
			log.Lvl3("Got message from ", sendingAddr.String())
			_ = protobuf.Decode(packetBytes, &packet)
		}

		//give it to the receive routine
		if !anonymous && !audio {
			go g.Receive(packet, *sendingAddr, errChan, client)
		}
	}

}

//ReplyToClient this takes all that was written in the buffer and returns it.
func (g *Gossiper) ReplyToClient() ([]byte, error) {
	log.Lvl4("Emptying buffer to send to client ")
	reader := bufio.NewReader(g.buffer)
	tosend := make([]byte, reader.Size())
	n, err := reader.Read(tosend)
	tosend = tosend[0:n]
	if err != nil || n <= 0 {
		return []byte{}, err
	}

	return tosend, nil

}
