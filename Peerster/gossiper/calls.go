package gossiper

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"go.dedis.ch/onet/log"
)

//      CALL REQUEST      //
// =========================
func (g *Gossiper) ClientSendCallRequest(destination string) {
	if !g.CallStatus.ExpectingResponse && !g.CallStatus.InCall {
		callRequest := CallRequest{Origin: g.Name, Destination: destination}
		g.CallStatus.ExpectingResponse = true
		g.CallStatus.InCall = false
		g.CallStatus.OtherParticipant = destination
		g.ReceiveCallRequest(callRequest)
	} else {
		log.Lvl2("Current gossiper is already in a call or expecting a call response")
	}
}

func (g *Gossiper) ReceiveCallRequest(req CallRequest) {
	if req.Destination == g.Name {
		// call request is for us
		log.Lvl2("Received a call request from: ", req.Origin)
		g.PrintCallRequest(req)
		addr := g.FindPath(req.Origin)
		if addr == "" {
			//we do not know this peer we stop here
			log.Error("No routing information available to node ", req.Origin)
			return
		}

		callResp := CallResponse{Origin: g.Name, Destination: req.Origin}
		if g.CallStatus.InCall || g.CallStatus.ExpectingResponse {
			// if gossiper is in another call or waiting for a response, send a BUSY
			callResp.Status = Busy
			// send BUSY response
			log.Lvl2("Sending a BUSY call response for ", req.Origin, " by routing it to : ", addr)
			packet := GossipPacket{CallResponse: &callResp}
			err := g.SendTo(addr, packet)
			if err != nil {
				log.Error(err)
			}
		} else {
			// terminal prompt
			reader := bufio.NewReader(os.Stdin)
			// TODO: Would this work? (e.g. need a loop until user puts Y or N, but
			//		do not want to block, want to still keep processing incoming messages)
			go func() {
				for {
					fmt.Printf("Received a call request from node %s: accept/decline [y/n]?\n", req.Origin)
					text, err := reader.ReadString('\n')
					if err != nil {
						log.Error("Error reading user input: ", err)
					}
					if strings.Compare(strings.TrimSpace(text), "y") == 0 {
						callResp.Status = Accept
						break
					} else if strings.Compare(strings.TrimSpace(text), "n") == 0 {
						callResp.Status = Decline
						break
					}
				}
				// send ACCEPT or DECLINE response
				g.SendCallResponse(callResp)
			}()
		}
	} else {
		if strings.Compare(req.Origin, g.Name) == 0 {
			go func() {
				// wait for 10 seconds for a response, if the other node doesn't pick up, hang up
				time.Sleep(10 * time.Second)
				if !g.CallStatus.InCall && g.CallStatus.ExpectingResponse {
					log.Lvl2("Node ", g.CallStatus.OtherParticipant, " did not pick up. Hanging up")
					g.ClientSendHangUpMessage()
				}
			}()
		}
		packet := GossipPacket{CallRequest: &req}
		routeMessage(g, packet, req.Destination)
	}
	return
}

//      CALL RESPONSE      //
// =========================
func (g *Gossiper) SendCallResponse(resp CallResponse) {
	if resp.Status == Accept {
		// if we accept a call request, update call status as follows
		g.CallStatus.InCall = true
		g.CallStatus.ExpectingResponse = false
		g.CallStatus.OtherParticipant = resp.Destination
		log.Lvl2("Accepting call from ", resp.Destination)
	} else if resp.Status == Decline {
		// if we decline a call request, update call status as follows ( we are NOT in another call)
		g.CallStatus.InCall = false
		g.CallStatus.ExpectingResponse = false
		g.CallStatus.OtherParticipant = resp.Destination
		log.Lvl2("Declining call from ", resp.Destination)
	}
	// the last possibility is if respond with BUSY - meaning we are in another call,
	//    so call status has been updated either when ACCEPTING someone's request, or
	//    having our request ACCEPTED by someone else - e.g. UPDATE NOTHING

	g.ReceiveCallResponse(resp)
}

func (g *Gossiper) ReceiveCallResponse(resp CallResponse) {
	if strings.Compare(resp.Destination, g.Name) == 0 {
		// we received a call response
		// check if we had sent a call request to the sender
		g.CallStatus.ExpectingResponse = false
		if resp.Status == Accept {
			log.Lvl2("Node ", resp.Origin, " picked up")
			g.CallStatus.InCall = true
			g.CallStatus.OtherParticipant = resp.Origin
		} else {
			if resp.Status == Decline {
				log.Lvl2("Node ", resp.Origin, " declined our call")
			} else if resp.Status == Busy {
				log.Lvl2("Node ", resp.Origin, " is in another call")
			}
			g.CallStatus.InCall = false
			g.CallStatus.OtherParticipant = ""
		}
	} else {
		packet := GossipPacket{CallResponse: &resp}
		routeMessage(g, packet, resp.Destination)
	}
	return
}

//      HANG UP MSG       //
// =========================
func (g *Gossiper) ClientSendHangUpMessage() {
	if (g.CallStatus.InCall && strings.Compare(g.CallStatus.OtherParticipant, "") != 0) ||
		(g.CallStatus.ExpectingResponse && strings.Compare(g.CallStatus.OtherParticipant, "") != 0) {

		hangUp := HangUp{Origin: g.Name, Destination: g.CallStatus.OtherParticipant}
		g.CallStatus.InCall = false
		g.CallStatus.ExpectingResponse = false
		g.CallStatus.OtherParticipant = ""
		g.ReceiveHangUpMessage(hangUp)
	}
}

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

//      HELPERS      //
// ====================
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
