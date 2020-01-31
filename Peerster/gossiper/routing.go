//@authors Hrusanov Aleksandar, Lanzrein Johan, Rinaldi Vincent

package gossiper

import (
	"go.dedis.ch/onet/log"
	"time"
)

func (g *Gossiper) RouteRumor(errChan chan error) {
	rm := RumorMessage{
		Origin: g.Name,
		ID:     g.counter,
		Text:   "",
	}
	pckt := GossipPacket{Rumor: &rm}
	go g.ReceiveRumorMessage(pckt, *g.gossipAddr, errChan)

}

func (g *Gossiper) RouteLoop() {
	errChan := make(chan error)
	for {
		<-time.After(time.Duration(g.rtimer) * time.Second)
		log.Lvl3("PING : ", g.rtimer, g.counter)
		go g.RouteRumor(errChan)
		//err := <-errChan
		//if err != nil {
		//	log.Error("Error on routing rumor : ", err)
		//}

	}

}

func (g *Gossiper) StartupRouting() {
	log.Lvl3(g.Name, "Startup routing")
	errChan := make(chan error)
	go g.RouteRumor(errChan)

}
