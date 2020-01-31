//@authors Hrusanov Aleksandar, Lanzrein Johan, Rinaldi Vincent


package gossiper

import (
	"errors"
	"go.dedis.ch/onet/log"
	"math/rand"
	"path"
	"time"
)

//ReceiveSearchRequest handle a search request locally. the specified flag is if the budget was specified.
func (g *Gossiper) ReceiveSearchRequest(sr *SearchRequest) error {
	//check locally only if its not our own file request..

	if sr.Origin != g.Name {
		log.Lvl2("Got search reques from : ", sr.Origin)
		//Check if already seen
		if g.AlreadySeen(sr) {
			log.Lvl2("Already seen recently")
			return nil
		}
		//Add it to the recently seen
		go g.TimeoutSearchRequest(sr)

		//Search locally
		results := g.GetLocalResults(sr)
		//Send the local results.
		if len(results) > 0 {
			log.Lvl2("Sending local results..:")
			go g.SendLocalResults(sr.Origin, results)

		}
	}
	//if sr.Origin != g.Name || !specified{
	sr.Budget -= 1
	//}

	//Send if possible to other peers. :)
	if sr.Budget > 0 {
		log.Lvl2("Distribution with budget : ", sr.Budget)
		//send it to as many neighbours as budget allows.
		g.KnownGossipers.mu.Lock()
		defer g.KnownGossipers.mu.Unlock()
		length := len(*g.KnownGossipers.values)
		amtPerPeer := sr.Budget / uint64(length)
		mod := sr.Budget % uint64(length)
		//shuffle maybe
		rand.Seed(time.Now().Unix())

		//create a random order.
		order := rand.Perm(length)
		for i := range *g.KnownGossipers.values {

			budget := amtPerPeer
			if uint64(i) < mod {
				budget++
			}
			if budget == 0 {
				break
			}

			request := SearchRequest{
				Origin:   sr.Origin,
				Budget:   budget,
				Keywords: sr.Keywords,
			}
			gp := GossipPacket{SearchRequest: &request}
			peer := (*g.KnownGossipers.values)[order[i]]
			log.Lvl2("Sending with budget :", budget, " to : ", peer)

			err := g.SendTo(peer, gp)
			if err != nil {
				log.Error("Error when sending to ", peer, " : ", err)

			}

		}

	}

	return nil

}

//SendLocalResult send to the origin the result of the local search.
func (g *Gossiper) SendLocalResults(origin string, results []*SearchResult) error {
	reply := SearchReply{
		Origin:      g.Name,
		Destination: origin,
		HopLimit:    g.HopLimit,
		Results:     results,
	}
	gp := GossipPacket{SearchReply: &reply}
	addr := g.FindPath(origin)
	if addr == "" {
		log.Error("Could not find path to the node")
		return errors.New("Unknown destination")
	}

	err := g.SendTo(addr, gp)
	if err != nil {
		log.Error("Could not send search reply : ", err)
		return err
	}
	return nil
}

func (g *Gossiper) GetLocalResults(sr *SearchRequest) []*SearchResult {
	g.metalock.mu.Lock()
	defer g.metalock.mu.Unlock()

	var results []*SearchResult
	log.Lvl2("Looking for keywords :", sr.Keywords)
	for _, md := range g.metalock.data {

		if matching(md.Name, sr.Keywords) {
			log.Lvl2("Match on : ", md.Name)
			//append the result

			chunks := len(md.Metafile) / 32
			chunkmap := GenerateSlice(chunks)
			chunkCount := uint64(md.Length/block + 1)
			res := SearchResult{
				FileName:     md.Name,
				MetafileHash: md.MetaHash,
				ChunkMap:     chunkmap,
				ChunkCount:   chunkCount,
			}
			results = append(results, &res)

		}
	}
	return results
}

func matching(s string, strings []string) bool {
	for _, pat := range strings {
		if ok, _ := path.Match("*"+pat+"*", s); ok {
			return true
		}
	}

	return false
}
