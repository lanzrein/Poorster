//structures : Different structures used by the gossiper

package gossiper

import (
	"github.com/JohanLanzrein/Peerster/clusters"
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
}

/***********DIFFERENT TYPES OF MESSAGES -******************/
type BroadcastMessage struct {
	ClusterID uint64
	Data      []byte
}

type RequestMessage struct {
	Origin    string
	PublicKey ies.PublicKey
}

type RequestReply struct {
	Accepted           bool
	ClusterInformation clusters.Cluster
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
}

//RumorMessage
type RumorMessage struct {
	Origin string
	ID     uint32
	Text   string
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
