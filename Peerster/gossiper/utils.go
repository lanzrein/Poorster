//@authors Hrusanov Aleksandar, Lanzrein Johan, Rinaldi Vincent

//Utilitiy methods that are not related to networking.
//These can include parsing, adding, modifiying parts of the gossiper or packets.
package gossiper

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"go.dedis.ch/onet/log"
	"net"
	"sort"
	"strings"
)

const DEBUGLEVEL = 1

//Contains check if e is in xs
func Contains(xs []string, e string) bool {
	for _, x := range xs {
		if x == e {
			return true
		}
	}

	return false

}

/**************UTILITY TO process PACKET ****************/
//HostsToString Parses the known gossipers to a single comma separated string
func (g *Gossiper) HostsToString() string {
	var res string = ""
	g.KnownGossipers.mu.Lock()
	for k, v := range *g.KnownGossipers.values {
		res += v
		if k+1 != len(*g.KnownGossipers.values) {
			res += ","
		}
	}
	g.KnownGossipers.mu.Unlock()

	return res
}

//VerifyFreshness check if a given gossip packet has already been seen
//will return true if the packet is fresh
func (g *Gossiper) VerifyFreshness(packet GossipPacket) bool {
	rm := packet.Rumor
	g.pl.mu.Lock()
	defer g.pl.mu.Unlock()
	if rm != nil {
		id := uint32(len(g.pl.logmap[rm.Origin]))
		return id < rm.ID //if we are at len = 0 , and its the first message (1) its ok.. if len = 1 and first message not ok.
	} else if packet.TLCMessage != nil {
		origin := packet.TLCMessage.Origin
		id := uint32(len(g.pl.logmap[origin]))
		return id < packet.TLCMessage.ID
	}

	//we have not received this message.
	return false

}

//FindStatus find the status for a given node s
//if the node has never been seen then the vector clock is initialiatized for this host {s,1}
func (g *Gossiper) FindStatus(s string) PeerStatus {
	g.pl.mu.Lock()
	defer g.pl.mu.Unlock()
	ps := PeerStatus{
		Identifier: s,
		NextID:     uint32(len(g.pl.logmap[s])) + 1,
	}

	return ps

}

//UpdatePeerLog update the peer log. This should only be used when receiving new rumor messages as it can
//throw off the gossiper otherwise
// Will update the value of the origin of the rumor message
func (g *Gossiper) UpdatePeerLog(packet GossipPacket) bool {
	//from a new message update the peer log.
	//ONLY USE IT WITH NEW MESSAGES !

	g.pl.mu.Lock()
	defer g.pl.mu.Unlock()
	if packet.Rumor != nil {
		return g.UpdatePeerLogRumor(packet)
	} else if packet.TLCMessage != nil {
		log.Lvl2("Updateing peer log for a tlc message")
		return g.UpdatePeerLogTLC(packet)
	}

	return false

}

func (g *Gossiper) UpdatePeerLogTLC(packet GossipPacket) bool {
	tlc := *packet.TLCMessage
	if tlc.Origin == g.Name {
		//its from the client....

		g.pl.logmap[g.Name] = append(g.pl.logmap[g.Name], packet)
		g.TimeMapping.Lock()
		g.TimeMapping.times[tlc.Origin] = append(g.TimeMapping.times[tlc.Origin], tlc.ID)
		g.TimeMapping.Unlock()
		return true
	}

	currId := len(g.pl.logmap[tlc.Origin])

	//increment the counter for message.Origin == the id -> increment by one. else ignore it as we are still waiting for the next message.
	if uint32(currId)+1 == tlc.ID {
		g.pl.logmap[tlc.Origin] = append(g.pl.logmap[tlc.Origin], packet)
		g.TimeMapping.Lock()
		g.TimeMapping.times[tlc.Origin] = append(g.TimeMapping.times[tlc.Origin], tlc.ID)
		g.TimeMapping.Unlock()
		return true
	}
	return false
}

func (g *Gossiper) UpdatePeerLogRumor(packet GossipPacket) bool {
	message := *packet.Rumor
	if message.Origin == g.Name {
		//its from the client.
		g.counter++
		g.pl.logmap[message.Origin] = append(g.pl.logmap[message.Origin], packet)

		return true

	}

	currId := len(g.pl.logmap[message.Origin])

	//increment the counter for message.Origin == the id -> increment by one. else ignore it as we are still waiting for the next message.
	if uint32(currId)+1 == message.ID {
		g.pl.logmap[message.Origin] = append(g.pl.logmap[message.Origin], packet)
		return true
	}
	return false
}

//EvaluateStatus packet
//The gossiper will verify given a status packet if its more up to date, or less up to date, or in synch with it
//The int value returned explains the state, the string is the identifier of the Status where there is a difference.
func (g *Gossiper) EvaluateStatus(packet StatusPacket) (string, int) {
	//decide what to do with status packet
	//>0 - can send a new packet
	// < 0 - need a new packet
	// == 0 - equal
	//ElemCount := 0

	g.pl.mu.Lock()
	defer g.pl.mu.Unlock()

	for key, list := range g.pl.logmap {
		found := false
		for _, want := range packet.Want {
			//first check if the name is the same
			if want.Identifier == key {
				//same name..
				if want.NextID-1 < uint32(len(list)) {
					log.Lvl3("Based on : ", want.Identifier, want.NextID)
					log.Lvl3("Sending :", key, want.NextID-1)
					return key, int(want.NextID)
				} else if want.NextID-1 > uint32(len(list)) {
					log.Lvl3("We want something from : ", key, "he has : ", want.NextID, "i have ", len(list))
					return key, -1
				}

				found = true
				continue
			}
		}
		if !found && len(list) > 0 {
			//this key is unknown to the sender..
			log.Lvl3("Sending :", key, 1)
			return key, 1
		}
	}

	return "", 0

}

//PeerLogList Returns the peer list -> this allows to use it without having to use locks.
func (g *Gossiper) PeerLogList() []PeerStatus {
	//returns the peer log map as a list.
	g.pl.mu.Lock()
	xs := make([]PeerStatus, len(g.pl.logmap))
	i := 0
	for key, value := range g.pl.logmap {
		val := PeerStatus{
			Identifier: key,
			NextID:     uint32(len(value)) + 1,
		}
		xs[i] = val
		i++

	}
	g.pl.mu.Unlock()

	return xs
}

//SelectGossipPacket a packet message that would update the Wants Peerstatus
//if no message are found returns an empty message.
func (g *Gossiper) SelectGossipPacket(Wants []PeerStatus) (GossipPacket, bool) {
	//select what rumor to send based on the rumor logs
	g.pl.mu.Lock()
	defer g.pl.mu.Unlock()
	OriginsCommon := []string{}
	for _, ps := range Wants {
		packets := g.pl.logmap[ps.Identifier]
		if uint32(len(packets)) >= ps.NextID {
			//there are more packets so tehre is one to send.
			return packets[ps.NextID-1], true
		}
		OriginsCommon = append(OriginsCommon, ps.Identifier)

	}

	//we send a packet from a host that the Wants has no knowledge about
	//Select an element that is in set(g.rumorLogs) - set(OriginsCommon)
	for key, val := range g.pl.logmap {
		if !Contains(OriginsCommon, key) {
			return val[0], true
		}
	}

	//if we get here it means that no message was found..
	//we return an empty message..
	return GossipPacket{}, false

}

//ParseKnownHosts Parses a comma separated string into a list of string.
func ParseKnownHosts(str string) ([]string, error) {

	xs := strings.Split(str, ",")
	if xs[0] == "" {
		return make([]string, 0), nil
	}
	return xs, nil
}

//GetWaitingAddresses get the addresses the gosisper is currently waiting on
func (g *Gossiper) GetWaitingAddresses() []string {
	g.wl.mu.Lock()
	keys := make([]string, len(g.wl.list))

	i := 0
	for k, _ := range g.wl.list {
		keys[i] = k
		i++
	}
	g.wl.mu.Unlock()

	return keys
}

//AddNewPeer add a new peer addr to the known gossiper of g
func (g *Gossiper) AddNewPeer(addr net.UDPAddr) {
	//var NewPeer bool = false
	g.KnownGossipers.mu.Lock()
	if addr.String() != g.gossipAddr.String() && !Contains(*g.KnownGossipers.values, addr.String()) {

		tmp := append(*g.KnownGossipers.values, addr.String())

		g.KnownGossipers.values = &tmp
	}
	g.KnownGossipers.mu.Unlock()

}

//RemoveFromWaitingList remove addr from the waiting list
func (g *Gossiper) RemoveFromWaitingList(addr string) RumorMessage {
	g.wl.mu.Lock()
	orig := g.wl.list[addr]
	//make safe copy
	rm := RumorMessage{
		orig.Origin, orig.ID, orig.Text, []string{},
	}
	delete(g.wl.list, addr)
	g.wl.mu.Unlock()
	return rm
}

//AddToWaitingList addr and its rumor message to the waiting list.
func (g *Gossiper) AddToWaitingList(addr string, message RumorMessage) {
	g.wl.mu.Lock()
	g.wl.list[addr] = message
	g.wl.mu.Unlock()
}

//GetPeerLogKeys get the keys of the peer log.
func (g *Gossiper) GetPeerLogKeys() []string {
	g.pl.mu.Lock()
	defer g.pl.mu.Unlock()
	keys := make([]string, len(g.pl.logmap))
	i := 0
	for k := range g.pl.logmap {
		keys[i] = k
	}

	return keys
}

//UpdateRoutingTable update the routing table according to a new rumor received.
func (g *Gossiper) UpdateRoutingTable(rumor RumorMessage, addr string) {
	log.Lvl3("updating routing table for : ", rumor.Origin)
	origin := rumor.Origin
	if origin == g.Name {
		return
	}
	g.routing.mu.Lock()
	defer g.routing.mu.Unlock()
	val := g.routing.table[origin]
	if uint32(val.int) < rumor.ID {
		g.routing.table[origin] = struct {
			string
			int
		}{string: addr, int: int(rumor.ID)}

		if rumor.Text != "" {
			g.DSDVPrint(origin, addr)
		}
	}

}

//Close close the connections of the gossiper..
func (g *Gossiper) Close() error {
	err := g.connClient.Close()
	if err != nil {
		log.Error(err)
	}
	err = g.connNode.Close()
	return err

}

//Find a path to destination - returns "" if none is known.
func (g *Gossiper) FindPath(destination string) string {
	g.routing.mu.Lock()
	defer g.routing.mu.Unlock()
	return g.routing.table[destination].string
}

//GetOrigins get all the origins known to the gossiper
func (g *Gossiper) GetOrigins() []string {
	g.routing.mu.Lock()
	defer g.routing.mu.Unlock()
	xs := make([]string, len(g.routing.table))
	i := 0
	for k, _ := range g.routing.table {
		xs[i] = k
		i++
	}
	sort.Strings(xs)
	return xs
}

//GetChannelReply get the channel approriate for this reply if there is any, false if none exists
func (g *Gossiper) GetChannelReply(reply *DataReply) (chan DataReply, bool) {
	g.CurrentlyDownloading.mu.Lock()
	defer g.CurrentlyDownloading.mu.Unlock()
	//first check if we have it in our list of currently downloaded files.
	for c, k := range g.CurrentlyDownloading.values {
		if bytes.Compare(k, reply.HashValue) == 0 {
			return c, true
		}
	}

	return nil, false

}

func GenerateSlice(max int) []uint64 {
	res := make([]uint64, max)
	i := uint64(0)
	for i < uint64(max) {
		res[i] = i + 1
		i++
	}
	return res
}

func GenerateId() *uint64 {
	array := make([]byte, 8)
	_, err := rand.Read(array)
	if err != nil {
		panic(err)
	}
	id := new(uint64)
	err = binary.Read(bytes.NewBuffer(array), binary.LittleEndian, id)
	if err != nil {
		panic(err)
	}
	return id
}
