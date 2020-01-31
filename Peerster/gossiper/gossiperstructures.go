//@authors Hrusanov Aleksandar, Lanzrein Johan, Rinaldi Vincent


package gossiper

import (
	"bytes"
	"gopkg.in/hraban/opus.v2"

	"github.com/JohanLanzrein/Peerster/clusters"
	"github.com/JohanLanzrein/Peerster/ies"
	"github.com/jfreymuth/pulse"
	"go.dedis.ch/onet/log"
	//	opus "gopkg.in/hraban/opus.v2"

	//"encoding/hex"
	"net"
	"sync"
	"time"
)

//RoutingTable a routing table that is multithread safe
type RoutingTable struct {
	mu    sync.Mutex
	table map[string]struct {
		string
		int
	}
}

//KnownGossiper Atomic structure to add the values of the known gossiper
type KnownGossiper struct {
	//use the mutext to ensure safe read write on the values of known gossiper
	mu     sync.Mutex
	values *[]string
}

//PeerLog Atomic structure to have a log of messages we have seen by each peer.
type PeerLog struct {
	mu     sync.Mutex
	logmap map[string][]GossipPacket
}

//MetaLock holds value for all the metadata
type MetaLock struct {
	mu   sync.Mutex
	data []*MetaData
}

//CurrentlyUploading holds a concurrent safe map of datarequest channels.
type CurrentlyUploading struct {
	mu     sync.Mutex
	values map[chan DataRequest]string
}

//CurrentlyDownloading holds a concurrent safe map of datareplies.
type CurrentlyDownloading struct {
	mu     sync.Mutex
	values map[chan DataReply][]byte
}

//RecentlySeenRequest contains the recently seen search requests.
type RecentlySeenRequest struct {
	sync.Mutex
	values []SearchRequest
}

//WaitingList A list of all the peers we are waiting on.
type WaitingList struct {
	mu   sync.Mutex
	list map[string]RumorMessage
}

//FoundFiles a structure containing all the found files.
type FoundFiles struct {
	Filename     string
	MetaHash     []byte
	OriginChunks map[uint64]string //where the file is
	ChunkCount   uint64
	Done         int
}

//CurrentlySearching a list of all the currently searching files.
type CurrentlySearching struct {
	sync.Mutex
	values []FoundFiles
}

//Gossiper The gossiper structure.
//All the parameters help to handle the data and pass it forward.
type Gossiper struct {
	Name           string
	UIPort         string
	connClient     net.UDPConn
	connNode       net.UDPConn
	gossipAddr     *net.UDPAddr
	KnownGossipers KnownGossiper
	antiEntropy    time.Duration

	SimpleMode bool

	counter uint32

	//keep all the messages in case someone asks for them
	//rumorLogs RumorLogs
	//list of peers we are currently waiting on
	wl WaitingList
	//this is approximately the want list. also contains the packets.
	pl PeerLog
	//a buffer that the peerster writes to to send to the GUI if requested.
	buffer  *bytes.Buffer
	GUIPort string
	routing RoutingTable
	rtimer  uint

	//data structure to keep the meta data with locks
	//Metadata with locks == metalock
	metalock MetaLock
	//structure for sending/receiving chans
	//filesharing FileSharingChannels

	CurrentlyDownloading CurrentlyDownloading

	//HW3 File searching
	RecentlySeen       RecentlySeenRequest
	CurrentlySearching CurrentlySearching
	FoundFiles         []*FoundFiles
	SearchMatches      map[string]uint64 // keywords -> matches
	Chunks             map[string][]byte

	//ackAll blocks
	my_time          uint32
	Acknowledgements map[uint32][]string
	ackAll           bool
	TLCMessages      map[string]*TLCMessage
	N                int
	stubbornTimeout  uint64
	PreviousHash     []byte

	hw3ex3               bool
	HopLimit             uint32
	ConfirmedMessages    ConcTLCMessages //map of TLC messages by their ID
	FutureMessages       ConcTLCMessages //messages from the future that are stored for later uses.
	GiveUp               chan bool
	RunningConfirmation  bool
	FinishedConfirmation chan bool
	StackConfirmation    []*TxPublish

	TimeMapping TimeMapping

	//Stuff for project

	Keypair             *ies.KeyPair
	Cluster             *clusters.Cluster
	LeaveChan           chan bool
	Keys                map[string]ies.PublicKey
	HearbeatTimer       int
	RolloutTimer        int //timer in seconds ~~ usually 300
	CallStatus          GossiperCallStatus
	PulseClient         *pulse.Client
	RecordStream        *pulse.RecordStream
	RecordFrame         []int16
	PlaybackStream      *pulse.PlaybackStream
	PlayBackFrame       []int16
	AudioDataSlice      []byte
	OpusEncoder         *opus.Encoder
	OpusDecoder         *opus.Decoder
	AudioChan           chan struct{}
	is_authority        bool
	nb_authorities      int
	slice_results       [][]string
	acks_cases          map[string][]string
	correct_results_rcv map[string][]string
	reset_requests      map[string][]string
	//=======
	//	Keypair                   *ies.KeyPair
	//	Cluster                   *clusters.Cluster
	//	LeaveChan                 chan bool
	//	Keys                      map[string]ies.PublicKey
	//	HearbeatTimer             int
	//	RolloutTimer              int //timer in seconds ~~ usually 300
	//	CallStatus                GossiperCallStatus
	//	PulseClient               *pulse.Client
	//	RecordStream              *pulse.RecordStream
	//	RecordFrame               []int16
	//	PlaybackStream            *pulse.PlaybackStream
	//	PlayBackFrame             []int16
	//	AudioDataSlice            []byte
	////	OpusEncoder               *opus.Encoder
	////	OpusDecoder               *opus.Decoder
	//	AudioChan                 chan struct{}
	//	is_authority              bool
	//	nb_authorities            int
	//	slice_results             [][]string
	//	acks_cases                map[string][]string
	//	correct_results_rcv       map[string][]string
	//	reset_requests map[string][]string
	//>>>>>>> vincent
	members_ready_resend_requests map[string][]string
	pending_nodes_requests        []string
	pending_messages_requests     []RequestMessage
	displayed_requests            []string
}

type GossiperCallStatus struct {
	InCall            bool
	OtherParticipant  string
	ExpectingResponse bool
	IncomingCall      bool
}

//TimeMapping the mapping of the time <-> id for each known gossiper.
type TimeMapping struct {
	sync.Mutex
	times map[string][]uint32 //origin -> time -> id in the peerlog
}

//ConcTLCMessage a map of the seen TLCMessages concurrent safe.
type ConcTLCMessages struct {
	sync.Mutex
	values map[uint32][]*TLCMessage
}

//AlreadySeen true iff the request has been recently seen ( in the last 0.5 sec )
func (g *Gossiper) AlreadySeen(request *SearchRequest) bool {
	g.RecentlySeen.Lock()
	defer g.RecentlySeen.Unlock()
	for _, elem := range g.RecentlySeen.values {
		if elem.Origin == request.Origin && len(elem.Keywords) == len(request.Keywords) {
			equalCnt := 0
			for _, keyword := range request.Keywords {
				for _, keyword2 := range elem.Keywords {
					if keyword == keyword2 {
						equalCnt++
					}
				}
			}
			if equalCnt == len(elem.Keywords) {
				return true
			}

		}
	}

	return false
}

//TimeoutSearchRequest add a request to the timeout - meaning it has been recently seen.
func (g *Gossiper) TimeoutSearchRequest(request *SearchRequest) {
	r := *request
	g.RecentlySeen.Lock()
	g.RecentlySeen.values = append(g.RecentlySeen.values, *request)
	g.RecentlySeen.Unlock()
	<-time.After(500 * time.Millisecond)
	g.RecentlySeen.Lock()
	//find the thing. and remove it
	i := 0
	for _, elem := range g.RecentlySeen.values {
		if elem.Origin == r.Origin && len(r.Keywords) == len(elem.Keywords) {
			equalCnt := 0
			for _, keyword := range r.Keywords {
				for _, keyword2 := range elem.Keywords {
					if keyword == keyword2 {
						equalCnt++
					}
				}
			}
			if equalCnt == len(elem.Keywords) {
				break
			}
		}
		i++
	}

	g.RecentlySeen.values = append(g.RecentlySeen.values[:i], g.RecentlySeen.values[i+1:]...)
	g.RecentlySeen.Unlock()

}

//DownloadFoundFile downloads a file with an implicit destination
func (g *Gossiper) DownloadFoundFile(request []byte, filename string) {
	//check in found files all files that matches and download them.
	log.Lvl2("found files :", g.FoundFiles)
	for _, elem := range g.FoundFiles {
		//dl a found file
		log.Lvl2("elem :", elem.Filename)
		if bytes.Equal(elem.MetaHash, request) {
			log.Lvl2("Found match :", filename)
			g.DownloadResult(*elem, filename)
			break
		}

	}

}

//DownloadResult downloads a result from a file FoundFile undername filename
func (g *Gossiper) DownloadResult(files FoundFiles, filename string) {
	replies := make(chan DataReply)
	log.Lvl2(files.OriginChunks[1])
	go g.FileSharingReceiveProtocol(files.OriginChunks[1], files.MetaHash, replies, filename, &files)

}

func (g *Gossiper) GetConfirmedTLC(i int, identifier string) *TLCMessage {
	g.ConfirmedMessages.Lock()

	xs := g.ConfirmedMessages.values[uint32(i)]
	g.ConfirmedMessages.Unlock()
	for _, val := range xs {
		if val.Origin == identifier {
			return val
		}
	}
	return nil
}

func (g *Gossiper) GetVectorClock() *StatusPacket {
	sp := new(StatusPacket)

	sp.Want = g.PeerLogList()
	return sp
}
