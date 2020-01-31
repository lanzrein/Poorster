//@authors Hrusanov Aleksandar, Lanzrein Johan, Rinaldi Vincent

//Stringify the messages according to the spec.
//This part is very self explanatory so no further comments are needed...
package gossiper

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/JohanLanzrein/Peerster/clusters"
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
	str := strconv.FormatUint(clusterID, 10)
	s := fmt.Sprintf("REQUEST TO JOIN %s DENIED\n", str)
	fmt.Print(s)
	g.WriteToBuffer(s)
}

func (g *Gossiper) PrintDeniedExpelling(clusterID uint64) {
	str := strconv.FormatUint(clusterID, 10)
	s := fmt.Sprintf("A REQUEST TO EXPEL YOU FROM CLUSTER %s HAS BEEN DENIED\n", str)
	fmt.Print(s)
	g.WriteToBuffer(s)
}

func (g *Gossiper) PrintAcceptJoiningID(cluster clusters.Cluster) {
	str := strconv.FormatUint(*cluster.ClusterID, 10)

	s := fmt.Sprintf("REQUEST TO JOIN %s ACCEPTED. CURRENT MEMBERS : ", str)

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

func (g *Gossiper) PrintGotExpelled(clusterID uint64) {
	str := strconv.FormatUint(clusterID, 10)
	s := fmt.Sprintf("A REQUEST TO EXPEL YOU FROM CLUSTER %s HAS BEEN ACCEPTED.", str)
	fmt.Print(s)
	g.WriteToBuffer(s)
}

func (g *Gossiper) PrintBroadcast(message RumorMessage) {
	s := fmt.Sprint("Broadcast origin ", message.Origin, " contents ", message.Text, "\n")
	g.WriteToBuffer(s)
	fmt.Print(s)
}

func (g *Gossiper) PrintLeaveCluster(id uint64) {
	s := fmt.Sprintf("LEAVING cluster %s\n", strconv.FormatUint(id, 10))
	g.WriteToBuffer(s)
	fmt.Print(s)
}

func (g *Gossiper) PrintInitCluster() {
	str := strconv.FormatUint(*g.Cluster.ClusterID, 10)
	s := fmt.Sprintf("INIT new cluster %s\n", str)
	g.WriteToBuffer(s)
	fmt.Print(s)
}

func (g *Gossiper) PrintEvotingJoinStep(node string) {
	s := fmt.Sprintf("WAITING FOR CLIENT TO ACCEPT/DENY JOIN REQUEST FROM %s\n", node)
	g.WriteToBuffer(s)
	fmt.Print(s)
}

func (g *Gossiper) PrintEvotingExpelStep(node string) {
	if node != g.Name {
		s := fmt.Sprintf("WAITING FOR CLIENT TO ACCEPT/DENY EXPEL REQUEST FOR %s\n", node)
		g.WriteToBuffer(s)
		fmt.Print(s)
	}
}

func (g *Gossiper) PrintEvotingPropositionStep(rcvProp int, node string, origin string) {
	var s string
	if rcvProp == 1 {
		if node != g.Name {
			s = fmt.Sprintf("RECEIVED ACCEPT FOR CASE CONCERNING NODE %s FROM %s\n", node, origin)
			g.WriteToBuffer(s)
			fmt.Print(s)
		}
	} else { // receivedAnswer == 0
		if node != g.Name {
			s = fmt.Sprintf("RECEIVED DENY FOR CASE CONCERNING NODE %s FROM %s\n", node, origin)
			g.WriteToBuffer(s)
			fmt.Print(s)
		}
	}
}

func (g *Gossiper) PrintEvotingCaseStep(caseStep string, authOrigin string) {
	if strings.Contains(caseStep, g.Name) == false {
		s := fmt.Sprintf("RECEIVED COMPARISON REQUEST FOR CASE %s FROM %s\n", caseStep, authOrigin)
		g.WriteToBuffer(s)
		fmt.Print(s)
	}
}

func (g *Gossiper) PrintEvotingValidationStep(results []string, authOrigin string) {
	if strings.Contains(results[0], "EXPEL") {
		if (results[0])[6:] != g.Name {
			s := fmt.Sprintf("RECEIVED RESULTS LIST %x FOR VALIDATION FROM %s\n", results, authOrigin)
			g.WriteToBuffer(s)
			fmt.Print(s)
		}
	} else { // strings.Contains(results[0], "EXPEL") == false
		s := fmt.Sprintf("RECEIVED RESULTS LIST %x FOR VALIDATION FROM %s\n", results, authOrigin)
		g.WriteToBuffer(s)
		fmt.Print(s)
	}
}

func (g *Gossiper) PrintEvotingDecisionStep(decision string, memberOrigin string) {
	s := fmt.Sprintf("RECEIVED DECISION %s FROM %s\n", decision, memberOrigin)
	g.WriteToBuffer(s)
	fmt.Print(s)
}

//PrintPrivateMessage prints a received private messages
func (g *Gossiper) PrintAnonymousPrivateMessage(message PrivateMessage) {
	var s string
	if message.Origin == "" {
		s = fmt.Sprint("FULLY ANONYMOUS contents ", message.Text, "\n")
	} else {
		s = fmt.Sprint("ANONYMOUS from ", message.Origin, " with contents ", message.Text, "\n")
	}
	fmt.Print(s)
	g.WriteToBuffer(s)
}

// **NOTE - the 'path' field of an anonymous message is only used for testing
//PrintAnonymousPrivateMessagePath prints a received private messages
// func (g *Gossiper) PrintAnonymousPrivateMessagePath(path string) {
// 	s := fmt.Sprint("ANONYMOUS message path ", path, "\n")
// 	fmt.Print(s)
// 	g.WriteToBuffer(s)
// }

//PrintCallRequest prints receiving a call request
func (g *Gossiper) PrintCallRequest(message CallRequest) {
	s := fmt.Sprintf("CALL REQUEST from %s\n", message.Origin)
	fmt.Print(s)
	g.WriteToBuffer(s)
}

func (g *Gossiper) PrintCallAccepted(callee string) {
	s := fmt.Sprintf("CALL ACCEPTED by %s\n", callee)
	fmt.Print(s)
	g.WriteToBuffer(s)
}

func (g *Gossiper) PrintCallDeclined(callee string) {
	s := fmt.Sprintf("CALL DECLINED by %s\n", callee)
	fmt.Print(s)
	g.WriteToBuffer(s)
}

func (g *Gossiper) PrintCallBusy(callee string) {
	s := fmt.Sprintf("LINE BUSY by %s\n", callee)
	fmt.Print(s)
	g.WriteToBuffer(s)
}

func (g *Gossiper) PrintHangUp(callee string) {
	s := fmt.Sprintf("HANG UP by %s\n", callee)
	fmt.Print(s)
	g.WriteToBuffer(s)
}

func (g *Gossiper) PrintEvotingCancellationStep(cancelCase string, authOrigin string) {
	s := fmt.Sprintf("RECEIVED CANCELLATION FOR CASE %s FROM %s\n", cancelCase, authOrigin)
	g.WriteToBuffer(s)
	fmt.Print(s)
}

func (g *Gossiper) PrintEvotingResetStep(memberReset string, caseReset string) {
	if strings.Contains(caseReset, g.Name) == false {
		s := fmt.Sprintf("RECEIVED RESET FROM MEMBER %s FOR CASE %s\n", memberReset, caseReset)
		g.WriteToBuffer(s)
		fmt.Print(s)
	}
}

func (g *Gossiper) PrintEvotingResendStep(memberAck string, request string) {
	s := fmt.Sprintf("RECEIVED RESEND ACK FROM MEMBER %s FOR REQUEST %s\n", memberAck, request)
	g.WriteToBuffer(s)
	fmt.Print(s)
}
