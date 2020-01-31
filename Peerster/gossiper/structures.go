//@authors Hrusanov Aleksandar, Lanzrein Johan, Rinaldi Vincent

//structures : Different structures used by the gossiper

package gossiper

import (
	"github.com/JohanLanzrein/Peerster/ies"
)

//GossipPacket Main packet sent over the network
type GossipPacket struct {
	//HW1
	Simple *SimpleMessage
	Rumor  *RumorMessage
	Status *StatusPacket
	//HW2
	Private     *PrivateMessage
	DataRequest *DataRequest
	DataReply   *DataReply
	//HW3
	SearchRequest *SearchRequest
	SearchReply   *SearchReply

	TLCMessage *TLCMessage
	Ack        *TLCAck

	Broadcast *BroadcastMessage

	JoinRequest  *RequestMessage
	RequestReply *RequestReply

	AnonymousMsg *AnonymousMessage

	CallRequest  *CallRequest
	CallResponse *CallResponse
	HangUpMsg    *HangUp
	AudioMsg     *AudioMessage
}

/***********DIFFERENT TYPES OF MESSAGES -******************/
type BroadcastMessage struct {
	ClusterID         uint64
	HopLimit          uint32
	Destination       string
	Data              []byte
	Rollout           bool
	Reset             bool
	LeaveRequest      bool
	JoinRequest       bool
	AcceptProposition bool
	DenyProposition   bool
	CaseCompare       bool
	ResultsValidation bool

	FinalDecision	  bool
	CancelRequest	  bool
	ResetIndication	  bool
	CaseRequest 	  string
	AckResend		  bool
	ExpelRequest	  bool
}

type RequestMessage struct {
	Origin    string
	Recipient string
	PublicKey ies.PublicKey
}

type RequestReply struct {
	Accepted           bool
	Banned			   bool
	ExpelRequest	   bool
	Recipient          string
	ClusterID          uint64
	EphemeralKey       []byte
	ClusterInformation []byte
}

/*AnonymousMessage
* EncryptedContent - GossipPacket encrypted with the receiver's public key
* Receiver - the name of the destination node
* RelayRate - a number between 0 and 1, indicating the anonimity level of the message
*									 used for flipping a weighted coin by each relaying node
* RouteToReceiver - initially false,
*										true if after coin flip the current node decides NOT to relay anymore and
*										routes the message to it's actual destination
* Path - !!! only used for testing !!!, comma separated values of the nodes which relayed the
*					anonymous message
 */
type AnonymousMessage struct {
	EncryptedContent []byte
	Receiver         string
	AnonymityLevel   float64
	RouteToReceiver  bool
	// Path             string
}

type CallRequest struct {
	Origin      string
	Destination string
}

type CallResponseStatus int

const (
	Accept CallResponseStatus = iota
	Decline
	Busy
)

type CallResponse struct {
	Origin      string
	Destination string
	Status      CallResponseStatus
}

type HangUp struct {
	Origin      string
	Destination string
}

type AudioMessage struct {
	Origin      string
	Destination string
	Content     AudioData
}

type AudioData struct {
	Data       []byte
	EncryptedN int
}

//SimpleMessage
type SimpleMessage struct {
	OriginalName  string
	RelayPeerAddr string
	Contents      string
}

//Message form the client
type Message struct {
	Text        string
	Destination *string
	File        *string
	Request     *[]byte
	//Hw3 things
	Budget   *uint64
	Keywords *[]string

	//Project things...
	Broadcast    *bool
	InitCluster  *bool
	LeaveCluster *bool
	JoinOther    *string
	ExpelOther 	 *string

	//Project - anonimity
	Anonymous     *bool
	RelayRate     *float64
	FullAnonimity *bool

	//Project - calls
	CallRequest    *bool
	HangUp         *bool
	StartRecording *bool
	StopRecording  *bool

	//Project - e-voting
	PropAccept *string
	PropDeny   *string
}

//RumorMessage
type RumorMessage struct {
	Origin  string
	ID      uint32
	Text    string
	Results []string
}

//PeerStatus
type PeerStatus struct {
	Identifier string
	NextID     uint32
}

//StatusPacket
type StatusPacket struct {
	Want []PeerStatus
}

//PrivateMessage
type PrivateMessage struct {
	Origin      string
	ID          uint32
	Text        string
	Destination string
	HopLimit    uint32
}

//DataRequest
type DataRequest struct {
	Origin      string
	Destination string
	HopLimit    uint32
	HashValue   []byte
}

//DataReply
type DataReply struct {
	Origin      string
	Destination string
	HopLimit    uint32
	HashValue   []byte
	Data        []byte
}

//MetaData
type MetaData struct {
	Name     string
	Length   int64
	Metafile []byte
	MetaHash []byte
}

//SearchRequest
type SearchRequest struct {
	Origin   string
	Budget   uint64
	Keywords []string
}

//SearchReply
type SearchReply struct {
	Origin      string
	Destination string
	HopLimit    uint32
	Results     []*SearchResult
}

//SearchResult
type SearchResult struct {
	FileName     string
	MetafileHash []byte
	ChunkMap     []uint64
	ChunkCount   uint64
}

//TLCMessage
type TLCMessage struct {
	Origin      string
	ID          uint32
	Confirmed   int
	TxBlock     BlockPublish
	VectorClock *StatusPacket
	Fitness     float32
}

//TLCAck
type TLCAck PrivateMessage

//TxPublish
type TxPublish struct {
	Name         string
	Size         int64
	MetafileHash []byte
}

//BlockPublish
type BlockPublish struct {
	PrevHash    [32]byte
	Transaction TxPublish
}
