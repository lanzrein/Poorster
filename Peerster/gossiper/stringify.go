//Stringify the messages according to the spec.
//This part is very self explanatory so no further comments are needed...
package gossiper

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"github.com/JohanLanzrein/Peerster/clusters"
	"strings"
)

const test = false
const hw2 = true

/**********Printing UTILITIES ********************/

//NodePrint Prints a simple messages
func (g *Gossiper) NodePrint(msg SimpleMessage) {
	if hw2 {
		return
	}
	s := fmt.Sprintf("SIMPLE MESSAGE origin %v from %v contents %v\n", msg.OriginalName, msg.RelayPeerAddr, msg.Contents)
	g.WriteToBuffer(s)
	fmt.Print(s)
}

//PeerPrint prints the peer of the gossiper
func (g *Gossiper) PeerPrint() {
	if test {
		g.routing.mu.Lock()
		defer g.routing.mu.Unlock()
		fmt.Println(g.routing.table)
		return
	}
	if hw2 {
		return
	}

	s := fmt.Sprint("PEERS ", g.HostsToString(), "\n")
	fmt.Print(s)
	g.WriteToBuffer(s)

}

//ClientPrint prints a client message
func (g *Gossiper) ClientPrint(message SimpleMessage) {
	s := fmt.Sprintf("CLIENT MESSAGE %s \n", message.Contents)
	fmt.Print(s)
	g.WriteToBuffer(s)

}

//RumorPrint print a rumour from addr
func (g *Gossiper) RumorPrint(message RumorMessage, addr string) {
	if test || message.Text == "" {
		return
	}
	s := fmt.Sprint("RUMOR origin ", message.Origin, " from ", addr, " ID ", message.ID, " contents ", message.Text, "\n")
	g.WriteToBuffer(s)
	fmt.Print(s)
}

//MongeringPrint prints the mongering msg
func (g *Gossiper) MongeringPrint(addr string) {
	if test || hw2 {
		return
	}
	s := fmt.Sprint("MONGERING with ", addr, "\n")
	fmt.Print(s)
	g.WriteToBuffer(s)
}

//StatusPrint prints the status received
func (g *Gossiper) StatusPrint(message StatusPacket, addr string) {
	if test || hw2 {
		return
	}
	s := fmt.Sprint("STATUS from ", addr)

	for _, elem := range message.Want {
		s += fmt.Sprint(" peer ", elem.Identifier, " nextID ", elem.NextID)
		//g.WriteToBuffer(s)
		//fmt.Print(s)
	}
	s += fmt.Sprint("\n")
	fmt.Print(s)
	g.WriteToBuffer(s)

}

//FlipCoinPrint prints if the gossiper flipped to send
func (g *Gossiper) FlipCoinPrint(addr string) {
	if test || hw2 {
		return
	}
	s := fmt.Sprint("FLIPPED COIN sending rumor to ", addr, "\n")
	fmt.Print(s)
	g.WriteToBuffer(s)
}

//SyncPrint print that peers are in sync
func (g *Gossiper) SyncPrint(addr string) {
	if test || hw2 {
		return
	}
	s := fmt.Sprint("IN SYNC WITH ", addr, "\n")
	fmt.Print(s)
	g.WriteToBuffer(s)
}

//WriteToBuffer writes the current string to the buffer.
func (g *Gossiper) WriteToBuffer(content string) {
	writer := bufio.NewWriter(g.buffer)
	_, _ = fmt.Fprintf(writer, content)
	_ = writer.Flush()

}

//DSDVPrint prints a DSDV messages
func (g *Gossiper) DSDVPrint(peer string, addr string) {
	s := fmt.Sprint("DSDV ", peer, " ", addr, "\n")
	fmt.Print(s)
	g.WriteToBuffer(s)
}

//PrintPrivateMessage prints a received private messages
func (g *Gossiper) PrintPrivateMessage(message PrivateMessage) {
	s := fmt.Sprint("PRIVATE origin ", message.Origin, " hop-limit ", message.HopLimit, " contents ", message.Text, "\n")
	fmt.Print(s)
	g.WriteToBuffer(s)
}

//PrintDownloadStart prints that a downloading has started
func (g *Gossiper) PrintDownloadStart(filename string, destination string) {
	s := fmt.Sprint("DOWNLOADING metafile of ", filename, " from ", destination, "\n")
	fmt.Print(s)
	g.WriteToBuffer(s)
}

//PrintDownloadChunk prints the chunk downloaded
func (g *Gossiper) PrintDownloadChunk(filename string, destination string, chunk int64) {
	s := fmt.Sprint("DOWNLOADING ", filename, " chunk ", chunk, " from ", destination, "\n")
	fmt.Print(s)
	g.WriteToBuffer(s)
}

//PrintReconstructFile Prints that the file has been reconstructed
func (g *Gossiper) PrintReconstructFile(filename string) {
	s := fmt.Sprint("RECONSTRUCTED file ", filename, "\n")
	fmt.Print(s)
	g.WriteToBuffer(s)
}

//PrintClientPrivateMessage prints that the client sent a private message
func (g *Gossiper) PrintClientPrivateMessage(message PrivateMessage) {
	s := fmt.Sprint("CLIENT MESSAGE ", message.Text, " dest ", message.Destination, "\n")
	fmt.Print(s)
	g.WriteToBuffer(s)
}

//PrintMatch prints if there is a match for the search reply.
func (g *Gossiper) PrintMatch(message SearchReply) {
	for _, elem := range message.Results {
		hexS := hex.EncodeToString(elem.MetafileHash)
		chunks := ""
		for i, x := range elem.ChunkMap {
			chunks += fmt.Sprint(x)
			if i < len(elem.ChunkMap)-1 {
				chunks += fmt.Sprint(",")
			}

		}
		s := fmt.Sprint("FOUND match ", elem.FileName, " at ", message.Origin, " metafile=", hexS, " chunks=", chunks, "\n")
		fmt.Print(s)
		g.WriteToBuffer(s)
	}
}

//PrintSearchFinish print if the search is finished.
func (g *Gossiper) PrintSearchFinish(fullMatch bool) {
	s := ""
	if !fullMatch {
		s += fmt.Sprint("SEARCH ABORTED REACHED MAX BUDGET\n")
	} else {
		s += fmt.Sprint("SEARCH FINISHED\n")
	}
	fmt.Print(s)
	g.WriteToBuffer(s)
}

//PrintUnconfirmedGossip when broadcasting or receiving unconfirmed gossip
func (g *Gossiper) PrintUnconfirmedGossip(msg TLCMessage) {
	hash := hex.EncodeToString(msg.TxBlock.Transaction.MetafileHash)
	s := fmt.Sprint("UNCONFIRMED GOSSIP ORIGIN ", msg.Origin, " ID ", msg.ID, " file name ", msg.TxBlock.Transaction.Name)
	s += fmt.Sprint(" size ", msg.TxBlock.Transaction.Size, " metahash ", hash, "\n")
	fmt.Print(s)
	g.WriteToBuffer(s)
}

//PrintConfirmedGossip when receiving a confirmed gossip
func (g *Gossiper) PrintConfirmedGossip(msg TLCMessage) {
	hash := hex.EncodeToString(msg.TxBlock.Transaction.MetafileHash)
	s := fmt.Sprint("CONFIRMED GOSSIP ORIGIN ", msg.Origin, " ID ", msg.ID, " file name ", msg.TxBlock.Transaction.Name)
	s += fmt.Sprint(" size ", msg.TxBlock.Transaction.Size, " metahash ", hash, "\n")
	fmt.Print(s)
	g.WriteToBuffer(s)
}

//PrintRebroadcast print in case the tlc waslocally finished and there is a rebroadcast.
func (g *Gossiper) PrintRebroadcast(msg TLCMessage, witnesses []string) {
	witnessesString := strings.Join(witnesses, ",")
	s := fmt.Sprint("RE-BROADCAST ID ", msg.ID, " WITNESSES ", witnessesString, "\n")
	fmt.Print(s)
	g.WriteToBuffer(s)
}

//PrintSendAck print when sending an ack
func (g *Gossiper) PrintSendAck(ack TLCAck) {
	s := fmt.Sprint("SENDING ACK ", ack.Destination, " ID ", ack.ID, "\n")
	fmt.Print(s)
	g.WriteToBuffer(s)
}

//PrintAdvanceToNextRound
func (g *Gossiper) PrintAdvanceToNextRound(witnesses []*TLCMessage) {
	s := fmt.Sprint("ADVANCING TO round ", g.my_time, " BASED ON CONFIRMED MESSAGES ")

	for i, w := range witnesses {
		s += fmt.Sprint("origin", i+1, " ", w.Origin, " ID", i+1, " ", w.ID, " ")
	}
	s += fmt.Sprint("\n")
	fmt.Print(s)
	g.WriteToBuffer(s)
}

func (g *Gossiper) PrintDeniedJoining(clusterID uint64) {
	s := fmt.Sprintf("REQUEST TO JOIN %d DENIED\n", clusterID)
	fmt.Print(s)
	g.WriteToBuffer(s)
}

func (g *Gossiper) PrintAcceptJoiningID(cluster clusters.Cluster) {
	s := fmt.Sprintf("REQUEST TO JOIN %D ACCEPTED. CURRENT MEMBERS : ", cluster.ClusterID)
	for i, member := range cluster.Members {
		s += fmt.Sprintf("%s", member)
		if i < len(cluster.Members)-1 {
			s += fmt.Sprint(",")
		}
	}

	s += fmt.Sprint(".\n")
	fmt.Print(s)
	g.WriteToBuffer(s)
}

func (g *Gossiper) PrintBroadcast(message RumorMessage) {
	s := fmt.Sprint("Broadcast origin ", message.Origin, " contents ", message.Text, "\n")
	g.WriteToBuffer(s)
	fmt.Print(s)
}
