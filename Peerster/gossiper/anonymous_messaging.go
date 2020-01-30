//@authors Hrusanov Aleksandar, Lanzrein Johan, Rinaldi Vincent

package gossiper

import (
	"math/rand"
	"strings"
	"time"

	"github.com/JohanLanzrein/Peerster/clusters"
	"go.dedis.ch/onet/log"
)

//NodeCanSendAnonymousPacket - checks if both nodes are in the same cluster and if
//	sending node knows the receiver's public key
func (g *Gossiper) NodeCanSendAnonymousPacket(destination string) bool {
	// Both sending and receiving node need to be in the same cluster
	// Sending node needs to have information about the destination's public key
	if g.Cluster.ClusterID == nil {
		log.Error("Cannot send anonymous packet - current node does not belong to any cluster.")
		return false
	}
	if !isNodeInCluster(*g.Cluster, destination) {
		log.Error("Cannot send anonymous packet, node ", destination, " is not in the cluster.")
		return false
	}

	if _, ok := g.Cluster.PublicKeys[destination]; !ok {
		log.Error("Cannot send anonymous packet, public key of node ", destination, " is not available.")
		return false
	}

	return true
}

//ClientSendAnonymousMessage - handles anonymous message sending
func (g *Gossiper) ClientSendAnonymousMessage(destination string, text string, relayRate float64, fullAnonimity bool) {

	canSend := g.NodeCanSendAnonymousPacket(destination)

	if canSend {
		anonPrivate := PrivateMessage{Origin: g.Name, Text: text, Destination: destination}
		// anonymize the origin
		if fullAnonimity {
			anonPrivate.Origin = ""
		}
		gp := GossipPacket{Private: &anonPrivate}

		// sending an anonymous private message
		encryptedBytes := g.EncryptPacket(gp, destination)
		log.Lvl2("Encrypting anonymous message...")
		anonMsg := AnonymousMessage{
			EncryptedContent: encryptedBytes,
			Receiver:         destination,
			AnonymityLevel:   relayRate,
			RouteToReceiver:  false,
		}

		go g.ReceiveAnonymousMessage(&anonMsg)
	}
	return
}

// ReceiveAnonymousMessage - handles receiving a gossip packet with an anonymous message
func (g *Gossiper) ReceiveAnonymousMessage(anon *AnonymousMessage) {
	routeBecauseOfCoinFlip := false

	// **NOTE - the 'path' field of an anonymous message is only used for testing
	// anon.Path = anon.Path + g.Name + ", "
	if strings.Compare(anon.Receiver, g.Name) == 0 {
		// anonymous message is for us, decrypt it
		log.Lvl2("Decrypting an anonymous message addressed to us...")
		decryptedPacket, err := g.DecryptBytes(anon.EncryptedContent)
		if err != nil {
			log.Lvl2("Error decrypting an anonymous message")
			return
		}

		if decryptedPacket.Private != nil {
			// we received an anonymous private message
			log.Lvl2("Decrypted an anonymized private message")
			g.PrintAnonymousPrivateMessage(*decryptedPacket.Private)

			// **NOTE - only used for testing
			// g.PrintAnonymousPrivateMessagePath(anon.Path)
		} else if decryptedPacket.CallRequest != nil {
			// we received an anonymous call request
			log.Lvl2("Decrypted an anonymized call request")
			g.ReceiveCallRequest(*decryptedPacket.CallRequest)
		} else if decryptedPacket.CallResponse != nil {
			// we received an anonymous call response
			log.Lvl2("Decrypted an anonymized call response")
			g.ReceiveCallResponse(*decryptedPacket.CallResponse)
		} else if decryptedPacket.HangUpMsg != nil {
			// we received an anonymous hangup message
			log.Lvl2("Decrypted an anonymized hang up message")
			g.ReceiveHangUpMessage(*decryptedPacket.HangUpMsg)
		} else if decryptedPacket.AudioMsg != nil {
			// we received an anonymous audio message
			log.Lvl2("Decrypted an anonymized audio message")
			g.ReceiveAudio(*decryptedPacket.AudioMsg)
		}

	} else if !anon.RouteToReceiver {
		// anonymous message is not for us
		// if we are still relaying the message, flip a weighted coin to relay or send to destination
		log.Lvl2("Flipping a coin to relay or route anonymous message...")
		seed := rand.NewSource(time.Now().UnixNano())
		seededRand := rand.New(seed)
		randFloat := seededRand.Float64()

		packet := GossipPacket{AnonymousMsg: anon}
		if randFloat <= anon.AnonymityLevel {
			// if the random float is less than the desired anonimity level,
			//		pick a random neighbor and relay the message to them
			addr := g.SendToRandom(packet)
			log.Lvl2("Relaying the message to : ", addr)
		} else {
			routeBecauseOfCoinFlip = true
		}

		// if after flip we decided to route to destination or if the packet is already
		//	being routed to the destination (e.g. a node before us flipped a coint to route it)
		if routeBecauseOfCoinFlip || anon.RouteToReceiver {
			addr := g.FindPath(anon.Receiver)
			if addr == "" {
				//we do not know this peer we stop here
				log.Lvl2("No routing information for node ", anon.Receiver)
				return
			}
			log.Lvl2("Routing the anonymous message to it's destination: ", addr)
			err := g.SendTo(addr, packet)
			if err != nil {
				log.Error("Error sending an anonymous packet: ", err)
			}
			return
		}
	}
}

func isNodeInCluster(c clusters.Cluster, node string) bool {

	for _, m := range c.Members {
		if strings.Compare(m, node) == 0 {
			return true
		}
	}

	return false
}
