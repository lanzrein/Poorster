package gossiper

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"go.dedis.ch/onet/log"
)

//      CALL REQUEST      //
// =========================
func (g *Gossiper) ReceiveCallRequest(req CallRequest) {
	if req.Destination == g.Name {
		// call request is for us
		g.PrintCallRequest(req)

		addr := g.FindPath(req.Origin)
		if addr == "" {
			//we do not know this peer we stop here
			log.Error("No routing information available to node ", req.Origin)
			return
		}
		log.Lvl2("Forwarding a call response to : ", addr)

		callResp := CallResponse{Origin: g.Name, Destination: req.Origin}
		if g.CallStatus.InCall || g.CallStatus.ExpectingResponse {
			// if gossiper is in another call or waiting for a response, send a BUSY
			callResp.Status = Busy
		} else {
			// terminal prompt
			reader := bufio.NewReader(os.Stdin)
			for {
				fmt.Printf("Received a call request from node %s: accept/decline [y/n]?\n", req.Origin)
				text, err := reader.ReadString('\n')
				if err != nil {
					log.Error("Error reading user input: ", err)
				}
				if strings.Compare(text, "y") != 0 {
					callResp.Status = Accept
					g.CallStatus.InCall = true
					g.CallStatus.OtherParticipant = req.Origin
					break
				} else if strings.Compare(text, "n") != 0 {
					callResp.Status = Decline
					break
				}
			}
		}
		// send response
		packet := GossipPacket{CallResponse: &callResp}
		err := g.SendTo(addr, packet)
		if err != nil {
			log.Error(err)
		}
	} else {
		packet := GossipPacket{CallRequest: &req}
		routeMessage(g, packet, req.Destination)
	}
	return
}

//      CALL RESPONSE      //
// =========================
func (g *Gossiper) ReceiveCallResponse(resp CallResponse) {
	if strings.Compare(resp.Destination, g.Name) == 0 {
		// we received a call response
		// check if we had sent a call request to the sender

	} else {
		packet := GossipPacket{CallResponse: &resp}
		routeMessage(g, packet, resp.Destination)
	}
	return
}

func routeMessage(g *Gossiper, packet GossipPacket, dest string) {
	addr := g.FindPath(dest)
	if addr == "" {
		//we do not know this peer we stop here
		log.Error("No routing information available to node ", dest)
		return
	}
	log.Lvl2("Forwarding a packet to : ", addr)
	err := g.SendTo(addr, packet)
	if err != nil {
		log.Error(err)
	}
	return
}

//      HANG UP MSG       //
// =========================
func (g *Gossiper) ReceiveHangUpMessage(hangUp HangUp) {
	if strings.Compare(hangUp.Destination, g.Name) == 0 {
		// the other call participant wants to hangup on us
		if g.CallStatus.InCall || strings.Compare(hangUp.Origin, g.CallStatus.OtherParticipant) == 0 {
			g.CallStatus.InCall = false
			g.CallStatus.ExpectingResponse = false
			g.CallStatus.OtherParticipant = ""
		}
	} else {
		packet := GossipPacket{HangUpMsg: &hangUp}
		routeMessage(g, packet, hangUp.Destination)
	}
	return
}
