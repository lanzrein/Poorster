package gossiper

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jfreymuth/pulse"
	"go.dedis.ch/onet/log"
	opus "gopkg.in/hraban/opus.v2"
)

//      CALL REQUEST      //
// =========================
func (g *Gossiper) ClientSendCallRequest(destination string) {
	if !g.CallStatus.ExpectingResponse && !g.CallStatus.InCall {
		canSend := g.NodeCanSendAnonymousPacket(destination)
		if canSend {
			callRequest := CallRequest{Origin: g.Name, Destination: destination}
			g.CallStatus.ExpectingResponse = true
			g.CallStatus.InCall = false
			g.CallStatus.OtherParticipant = destination
			g.ReceiveCallRequest(callRequest)
		}
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

			gp := GossipPacket{CallRequest: &req}
			// sending an anonymous private message
			encryptedBytes := g.EncryptPacket(gp, req.Destination)
			log.Lvl2("Encrypting anonymous message...")
			anonMsg := AnonymousMessage{
				EncryptedContent: encryptedBytes,
				Receiver:         req.Destination,
				AnonymityLevel:   0.5,
				RouteToReceiver:  false,
			}

			go g.ReceiveAnonymousMessage(&anonMsg)
			go func() {
				// wait for 10 seconds for a response, if the other node doesn't pick up, hang up
				time.Sleep(10 * time.Second)
				if !g.CallStatus.InCall && g.CallStatus.ExpectingResponse {
					log.Lvl2("Node ", g.CallStatus.OtherParticipant, " did not pick up. Hanging up")
					g.ClientSendHangUpMessage()
				}
			}()
		}
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
			g.PrintCallAccepted(resp.Origin)
			g.CallStatus.InCall = true
			g.CallStatus.OtherParticipant = resp.Origin
		} else {
			if resp.Status == Decline {
				log.Lvl2("Node ", resp.Origin, " declined our call")
				g.PrintCallDeclined(resp.Origin)
			} else if resp.Status == Busy {
				log.Lvl2("Node ", resp.Origin, " is in another call")
				g.PrintCallBusy(resp.Origin)
			}
			g.CallStatus.InCall = false
			g.CallStatus.OtherParticipant = ""
		}
	} else {

		canSend := g.NodeCanSendAnonymousPacket(resp.Destination)
		if canSend {
			gp := GossipPacket{CallResponse: &resp}
			// sending an anonymous private message
			encryptedBytes := g.EncryptPacket(gp, resp.Destination)
			log.Lvl2("Encrypting anonymous message...")
			anonMsg := AnonymousMessage{
				EncryptedContent: encryptedBytes,
				Receiver:         resp.Destination,
				AnonymityLevel:   0.5,
				RouteToReceiver:  false,
			}

			go g.ReceiveAnonymousMessage(&anonMsg)
		}
	}
	return
}

//      HANG UP MSG       //
// =========================
func (g *Gossiper) ClientSendHangUpMessage() {
	if (g.CallStatus.InCall && strings.Compare(g.CallStatus.OtherParticipant, "") != 0) ||
		(g.CallStatus.ExpectingResponse && strings.Compare(g.CallStatus.OtherParticipant, "") != 0) {

		hangUp := HangUp{Origin: g.Name, Destination: g.CallStatus.OtherParticipant}
		log.Lvl2("Current node is hanging up on node ", g.CallStatus.OtherParticipant)
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
			log.Lvl2("Node ", hangUp.Origin, " hung up on us")
			g.PrintHangUp(hangUp.Origin)
			g.CallStatus.InCall = false
			g.CallStatus.ExpectingResponse = false
			g.CallStatus.OtherParticipant = ""
		}
	} else if strings.Compare(hangUp.Origin, g.Name) == 0 {
		// otherwise, it must be us sending a hangup message
		canSend := g.NodeCanSendAnonymousPacket(hangUp.Destination)
		if canSend {
			gp := GossipPacket{HangUpMsg: &hangUp}
			// sending an anonymous private message
			encryptedBytes := g.EncryptPacket(gp, hangUp.Destination)
			log.Lvl2("Encrypting anonymous message...")
			anonMsg := AnonymousMessage{
				EncryptedContent: encryptedBytes,
				Receiver:         hangUp.Destination,
				AnonymityLevel:   0.5,
				RouteToReceiver:  false,
			}

			go g.ReceiveAnonymousMessage(&anonMsg)
		}
	}
	return
}

//      AUDIO MESSAGES    //
// =========================
const sampleRate = 48000
const bufferFragmentSize = 1920
const bitRate = 32000
const numChanels = 1

func (g *Gossiper) ClientStartRecording() {
	// only process recording and sending audio if we are in a call with someone
	if g.CallStatus.InCall && strings.Compare(g.CallStatus.OtherParticipant, "") != 0 {
		// start recording and sending audio
		fmt.Println("RECORDING AUDIO DATA")

		g.AudioChan = make(chan struct{})
		go func() {
			for {
				audio := AudioMessage{Origin: g.Name, Destination: g.CallStatus.OtherParticipant}
				g.ReceiveAudio(audio)
				time.Sleep(2 * time.Second)
				select {
				case <-g.AudioChan:
					fmt.Println("\n Finish recording.")
					return
				default:
				}
			}
		}()
	} else {
		log.Error("Current node is not in a call - cannot send audio")
	}
	return
}

func (g *Gossiper) ClientStopRecording() {
	if g.AudioChan != nil {
		close(g.AudioChan)
	}
}

func (g *Gossiper) ReceiveAudio(audio AudioMessage) {
	if g.CallStatus.InCall {
		if strings.Compare(audio.Destination, g.Name) == 0 &&
			strings.Compare(audio.Origin, g.CallStatus.OtherParticipant) == 0 {
			// if we are in a call with the sender of this audio and it was intended for us
			//		listen to it
			fmt.Println("LISTENING TO AUDIO DATA from ", audio.Origin)
		} else if strings.Compare(audio.Destination, g.CallStatus.OtherParticipant) == 0 {
			// otherwise, it must be us sending the audio message out
			canSend := g.NodeCanSendAnonymousPacket(audio.Destination)
			if canSend {
				gp := GossipPacket{AudioMsg: &audio}
				// sending an anonymous private message
				encryptedBytes := g.EncryptPacket(gp, audio.Destination)
				log.Lvl2("Encrypting anonymous message...")
				anonMsg := AnonymousMessage{
					EncryptedContent: encryptedBytes,
					Receiver:         audio.Destination,
					AnonymityLevel:   0.5,
					RouteToReceiver:  false,
				}

				go g.ReceiveAnonymousMessage(&anonMsg)
			}
		}
	}
}

//      HELPERS      //
// ====================
func record() {
	// Opus Encoder
	enc, err := opus.NewEncoder(sampleRate, numChanels, opus.AppVoIP)
	if err != nil {
		log.Panic(err)
	}
	if err := enc.SetMaxBandwidth(opus.SuperWideband); err != nil {
		log.Panic(err)
	}
	if err := enc.SetBitrate(bitRate); err != nil {
		log.Panic(err)
	}

	// Pulse Client
	pa, err := pulse.NewClient()
	if err != nil {
		log.Panic(err)
	}
	defer pa.Close()

}

func playAudio() {

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
