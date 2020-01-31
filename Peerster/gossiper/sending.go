//@authors Hrusanov Aleksandar, Lanzrein Johan, Rinaldi Vincent

//sending handles all the sending of packets from the gossiper
package gossiper

import (
	"math/rand"
	"net"

	"go.dedis.ch/onet/log"
	"go.dedis.ch/protobuf"

	"time"
)

//SendPrivateMessage sends a private message to the intented receiver
func (g *Gossiper) SendPrivateMessage(message PrivateMessage) error {
	errChan := make(chan error)
	g.PrintClientPrivateMessage(message)
	go g.ReceivePrivateMessage(GossipPacket{Private: &message}, errChan)

	return <-errChan
}

//SendTo send a pckt packet to a host h
func (g *Gossiper) SendTo(h string, pckt GossipPacket) error {
	log.Lvl3("Sending new packet to ", h)
	msg, err := protobuf.Encode(&pckt)
	if err != nil {
		return err
	}
	foreign, err := net.ResolveUDPAddr("udp4", h)
	if err != nil {
		return err
	}

	_, err = g.connNode.WriteToUDP(msg, foreign)

	if err != nil {
		return err
	}
	return nil
}

//SendStatuspacket send a status packet. pretty basic.
func (g *Gossiper) SendStatusPacket(addr string, errChan chan error) {
	sp := StatusPacket{Want: g.PeerLogList()}
	err := g.SendTo(addr, GossipPacket{Status: &sp})
	errChan <- err
}

//SendAll send a packet to all hosts..
//as is the method is not very useful except but might come in handy later..
func (g *Gossiper) SendAll(pckt GossipPacket, avoid []string) error {
	log.Lvl3("Sending message to all peers except : ", avoid)

	g.KnownGossipers.mu.Lock()
	for _, h := range *g.KnownGossipers.values {

		if Contains(avoid, h) {
			continue
		}
		err := g.SendTo(h, pckt)
		if err != nil {
			return err
		}
	}
	g.KnownGossipers.mu.Unlock()
	return nil
}

//SendRumorPacket send a rumor message to the addr.
func (g *Gossiper) SendRumorPacket(message RumorMessage, errChan chan error, addr string) {
	//check if we have peers to send it to..

	pckt := GossipPacket{Rumor: &message}
	var i int
	//if not peer is given select one at random
	if addr == "" {
		if len(*g.KnownGossipers.values) == 0 {
			errChan <- nil
			return
		}
		rand.Seed(int64(time.Now().Nanosecond()))
		i := rand.Intn(len(*g.KnownGossipers.values))
		addr = (*g.KnownGossipers.values)[i]
	}

	log.Lvl3("Sending rumor to : ", addr, " i = ", i)
	//print that we are mongering with the address and send it
	g.MongeringPrint(addr)
	err := g.SendTo(addr, pckt)
	if err != nil {
		log.Error("Error on sending rumor packet : ", err)
		errChan <- err
	}

	//add it to our waiting list-
	g.AddToWaitingList(addr, message)

	//wait for the reply for 10*sec
	go func() {
		log.Lvl3("Waiting on : ", addr)
		<-time.After(10 * time.Second)
		//check if packet has not been acknowledged
		if Contains(g.GetWaitingAddresses(), addr) {
			//remove it from the waiting liste.
			log.Lvl3("Timeout on ", addr)
			_ = g.RemoveFromWaitingList(addr)
			//send it to a new person.
			g.SendRumorPacket(message, errChan, "")
		}
		errChan <- nil

	}()

}

//AcknowledgeRumor :
//This method is called when receiving a *status* packet acknowledging a pending *rumor* message.
//It will decide what needs to be done next according to the protocol.
func (g *Gossiper) AcknowledgeRumor(pckt GossipPacket, addr net.UDPAddr, errChan chan error) {

	sp := *pckt.Status
	//the value of id tells what is the next step
	key, id := g.EvaluateStatus(sp)
	//remove the rumor from the waiting list
	rm := g.RemoveFromWaitingList(addr.String())
	if rm.Origin == "" || rm.ID == 0 {
		//we had an unfortunate case
		//here while we were acking the message the timeout was reached and removed the rumor
		return
	}
	log.Lvl3("id : ", id, " key :", key)
	if id > 0 {
		//the current *g* has some more rumor to send to addr.
		//send a new packet packet to send is RumorLog[decision]
		gp := g.pl.logmap[key][id-1]
		g.SendToRandom(gp)

	} else if id < -1 {
		//the remote address has some unseen packet
		//request a new packet -> send the status
		g.SendStatusPacket(addr.String(), errChan)
	} else if id == 0 {
		//else we are in sync
		g.SyncPrint(addr.String())

		//send a rumor packet again with prob 1/2
		g.KnownGossipers.mu.Lock()
		defer g.KnownGossipers.mu.Unlock()
		rand.Seed(int64(time.Now().Nanosecond()))
		x := rand.Int()
		if x%2 == 0 {
			//send a new rumor packet.

			if len(*g.KnownGossipers.values) <= 0 {
				return //should not happen !!
			}
			//chose random host
			i := rand.Intn(len(*g.KnownGossipers.values))
			newHost := (*g.KnownGossipers.values)[i]

			//send the rumor we got form our address
			g.FlipCoinPrint(newHost)
			go g.SendRumorPacket(rm, errChan, newHost)

		}

	}

}

func (g *Gossiper) SendToRandom(gp GossipPacket) string {
	if len(*g.KnownGossipers.values) == 0 {
		return ""
	}
	rand.Seed(int64(time.Now().Nanosecond()))
	i := rand.Intn(len(*g.KnownGossipers.values))

	addr := (*g.KnownGossipers.values)[i]
	if addr == "" {
		return addr
	}

	err := g.SendTo(addr, gp)
	if err != nil {
		log.Errorf("Could not send packet to : %s, err : %s\n", addr, err)
	}
	return addr
}
