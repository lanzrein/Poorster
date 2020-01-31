//@authors Hrusanov Aleksandar, Lanzrein Johan, Rinaldi Vincent

//reveice handles all the part where the gossiper receives a packet.
package gossiper

import (
	"errors"
	"net"

	"go.dedis.ch/onet/log"
)

//Receive a new gossippacket from addr, and treat it accordingly.
func (g *Gossiper) Receive(pckt GossipPacket, addr net.UDPAddr, errChan chan error, client bool) {
	//packet has already been received.
	if client {
		//if the message comes from the client it is necessarly a new rumor message.
		if g.SimpleMode {

			g.ReceiveSimpleMessage(true, pckt, errChan, addr)

		} else {
			//build a rumor message
			if pckt.Private != nil {
				//private message
				errChan <- g.SendPrivateMessage(*pckt.Private)
			} else {
				rm := RumorMessage{
					Origin: g.Name,
					ID:     g.counter,
					Text:   pckt.Simple.Contents,
				}

				go g.ReceiveRumorMessage(GossipPacket{Rumor: &rm}, *g.gossipAddr, errChan)
				g.ClientPrint(*pckt.Simple)
			}

		}

		g.PeerPrint()
		return
	}

	//add the peer to the list of known gossiper
	g.AddNewPeer(addr)

	if pckt.Simple != nil {
		//receive a simple message
		log.Lvl3("Simple packet")
		g.ReceiveSimpleMessage(client, pckt, errChan, addr)
	} else if pckt.Rumor != nil {
		//do the rumor protocol
		log.Lvl3("Rumor packet")

		g.ReceiveRumorMessage(pckt, addr, errChan)

		//send the status back to acknowledge
		go g.SendStatusPacket(addr.String(), errChan)

	} else if pckt.Status != nil {
		//check the status
		log.Lvl3("Status packet")
		g.StatusPrint(*pckt.Status, addr.String())

		g.ReceiveStatusMessage(pckt, addr, errChan)
	} else if pckt.Private != nil {
		log.Lvl3("Private packet")
		g.ReceivePrivateMessage(pckt, errChan)
	} else if pckt.DataReply != nil {
		log.Lvl3("Received a data reply")

		g.ReceiveDataReply(pckt.DataReply)
	} else if pckt.DataRequest != nil {
		log.Lvl3("Received a data request")
		g.ReceiveDataRequest(pckt.DataRequest)
	} else if pckt.SearchRequest != nil {
		log.Lvl2("Got search request")
		go func() {
			errChan <- g.ReceiveSearchRequest(pckt.SearchRequest)
		}()
	} else if pckt.SearchReply != nil {
		log.Lvl2("Got search reply")
		go func() {
			errChan <- g.ReceiveSearchReply(*pckt.SearchReply)
		}()
	} else if pckt.TLCMessage != nil {
		log.Lvl2("Got a TLCMessage")
		g.ReceiveTLCMessage(pckt.TLCMessage)
	} else if pckt.Ack != nil {
		log.Lvl2("Got an ack ")
		g.ReceiveAck(pckt.Ack)
	} else if pckt.JoinRequest != nil {
		log.Lvl2("Got a join request")
		g.ReceiveJoinRequest(*pckt.JoinRequest)
	} else if pckt.RequestReply != nil {
		log.Lvl2("Got a request reply !")
		g.ReceiveRequestReply(*pckt.RequestReply)
	} else if pckt.Broadcast != nil {
		log.Lvl2("Got a broadcast !")
		g.ReceiveBroadcast(*pckt.Broadcast)
	} else if pckt.AnonymousMsg != nil {
		log.Lvl2("Got an anonymous message !")
		g.ReceiveAnonymousMessage(pckt.AnonymousMsg)
		// } else if pckt.CallRequest != nil {
		// 	log.Lvl2("Got a call request !")
		// 	g.ReceiveCallRequest(*pckt.CallRequest)
		// } else if pckt.CallResponse != nil {
		// 	log.Lvl2("Got a call response !")
		// 	g.ReceiveCallResponse(*pckt.CallResponse)
		// } else if pckt.HangUpMsg != nil {
		// 	log.Lvl2("Got a hang up message !")
		// 	g.ReceiveHangUpMessage(*pckt.HangUpMsg)
	} else {

		//should not happen !
		errChan <- errors.New("Empty packet.")
		return
	}

	//print the peers
	g.PeerPrint()

	errChan <- nil

}

//ReceiveStatusMessage receive a status message. in the pckt.
//Will decide the next steps
func (g *Gossiper) ReceiveStatusMessage(pckt GossipPacket, addr net.UDPAddr, errChan chan error) {
	//antientropy packet sending.
	//check if waiting on a rumor to arrive.
	waiting := Contains(g.GetWaitingAddresses(), addr.String())
	if waiting {
		//we are waiting on a rumor ack from this address.
		//acknowledgement
		g.AcknowledgeRumor(pckt, addr, errChan)

	} else {

		//its from antientropy.
		_, status := g.EvaluateStatus(*pckt.Status)
		if status > 0 {
			//start rumormongering with this address
			//we have some packets for this addr.
			gp, ok := g.SelectGossipPacket(pckt.Status.Want)
			if ok {
				go g.SendTo(addr.String(), gp)
			}
		} else if status < 0 {

			//send a status to request an update
			go g.SendStatusPacket(addr.String(), errChan)
		} else {
			//else we are equal.
			go g.SyncPrint(addr.String())
		}

	}

}

//ReceieRumorMessage receive a rumor message.
func (g *Gossiper) ReceiveRumorMessage(pckt GossipPacket, addr net.UDPAddr, errChan chan error) {
	//check if packet has already been received.
	fresh := g.VerifyFreshness(pckt)
	go g.UpdateRoutingTable(*pckt.Rumor, addr.String())

	if fresh {
		//only print it if its not from client
		if pckt.Rumor.Origin != g.Name {
			g.RumorPrint(*pckt.Rumor, addr.String())
			//update the routing table since its a fresh packet
			go g.UpdateRoutingTable(*pckt.Rumor, addr.String())

		}

		//send the rumor to someone randomly chosen
		go g.SendRumorPacket(*pckt.Rumor, errChan, "")

		//update the PeerLog.
		g.UpdatePeerLog(pckt)
		////add the pckt to the RumorLog
		//rumor := *pckt.Rumor
		//g.pl.mu.Lock()
		////alreadyStored := ContainedRumor(g.rumorLogs.values,*pckt.Rumor)
		////if !alreadyStored {
		//if cont {
		//	g.pl.logmap[rumor.Origin] = append(g.pl.logmap[rumor.Origin], pckt)
		//}
		//g.pl.mu.Unlock()

	}

}

//ReceiveSimpleMessage receive a simple message.
func (g *Gossiper) ReceiveSimpleMessage(client bool, pckt GossipPacket, errChan chan error, addr net.UDPAddr) {
	if client {
		//send to all known peers
		//change field Originalname and relayPeer
		sending := GossipPacket{Simple: &SimpleMessage{
			OriginalName:  g.Name,
			RelayPeerAddr: g.gossipAddr.String(),
			Contents:      pckt.Simple.Contents,
		}}

		g.ClientPrint(*pckt.Simple)
		err := g.SendAll(sending, []string{})
		errChan <- err

	} else {

		//if it comes from another peer A
		//store the name of A in the host list.

		g.AddNewPeer(addr)
		//change the relay peer field to its own address
		sending := GossipPacket{Simple: &SimpleMessage{
			OriginalName:  pckt.Simple.OriginalName,
			RelayPeerAddr: g.gossipAddr.String(),
			Contents:      pckt.Simple.Contents,
		}}
		//send ot all known peers except A
		avoid := make([]string, 1)
		avoid[0] = addr.String()
		err := g.SendAll(sending, avoid)
		if err != nil {
			errChan <- err
		}
		//print it.
		g.NodePrint(*pckt.Simple)

	}

}

//ReceivePrivateMessage receive a private message, forward it if needed
func (g *Gossiper) ReceivePrivateMessage(packet GossipPacket, errChan chan error) {
	private := *packet.Private
	log.Lvl3("Got priv : ", private)
	if private.Destination == g.Name {
		//its for us so we print it and stop
		g.PrintPrivateMessage(private)
	} else {

		if private.HopLimit > 0 {
			//send the message out
			private.HopLimit = private.HopLimit - 1
			addr := g.FindPath(private.Destination)
			if addr == "" {
				//we do not know this peer we stop here
				errChan <- nil
				return
			}
			log.Lvl2("Forwarding a message to : ", addr)
			packet = GossipPacket{Private: &private}
			errChan <- g.SendTo(addr, packet)
			return
		}
	}
	errChan <- nil
	return

}

//ReceiveDataReply receive a data reply and move it forward if needed
func (g *Gossiper) ReceiveDataReply(reply *DataReply) {
	log.Lvl3(g.Name, " : Got a data reply! ")
	if reply.Destination != g.Name {
		//forward it

		if reply.HopLimit > 0 {
			reply.HopLimit = reply.HopLimit - 1

			//send the message out

			addr := g.FindPath(reply.Destination)
			log.Lvl3("Forwarding a message to : ", addr)
			pckt := GossipPacket{DataReply: reply}

			go g.SendTo(addr, pckt)
			return
		}
	}
	//same idea as ReceiveDataRequest.
	channel, boolean := g.GetChannelReply(reply)
	if boolean {
		log.Lvl3(g.Name, " : Sending to an existing Channel")
		channel <- *reply
	} else {
		log.Lvl3(g.Name, "Could not find a Channel for the data")
		//start a new receiving protocol. which should usually not happen because
		//if we get some data it is because it is for us. and we are already expecting it.

	}

}

//ReceiveDataRequest handles a datarequest and sends it forward if needed.
func (g *Gossiper) ReceiveDataRequest(request *DataRequest) {
	//forwarding.
	if request.Destination != g.Name {

		if request.HopLimit > 0 {
			request.HopLimit = request.HopLimit - 1
			//send the message out
			addr := g.FindPath(request.Destination)
			log.Lvl2("Forwarding a message to : ", addr)
			pckt := GossipPacket{DataRequest: request}
			err := g.SendTo(addr, pckt)
			if err != nil {
				log.Error(err)
			}
			return
		}
	}

	go g.GetChunk(request.Origin, request.HashValue)
	return
	//receive a request for data first check if there is a pending process that handles this
	//channel, boolean := g.GetChannelRequest(request)
	//if boolean {
	//	log.Lvl3(g.Name, " sending on an existing channel ")
	//	//send it to the Channel and it will handle it.
	//	channel <- *request
	//} else {
	//	//its a new file to be downloaded so we need to start a new protocol
	//	log.Lvl3(g.Name, " : starting a new file sending ! ")
	//	datarequest := make(chan DataRequest)
	//	g.AddToDataRequests(datarequest, *request)
	//	go g.FileSharingSendProtocol(request.Origin, request.HashValue, datarequest)
	//}

}
