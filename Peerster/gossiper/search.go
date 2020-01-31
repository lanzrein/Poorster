//@authors Hrusanov Aleksandar, Lanzrein Johan, Rinaldi Vincent

package gossiper

import (
	"bytes"
	"go.dedis.ch/onet/log"
	"path"
	"strings"
)

const MATCHTHRESHOLD = 2
const MAXBUDGET = 32
const INITBUDGET = uint64(2)

//ReceiveSearchReply receives and handles a search reply
func (g *Gossiper) ReceiveSearchReply(reply SearchReply) error {
	if reply.Destination != g.Name {
		if reply.HopLimit == 0 {
			return nil
		}
		reply.HopLimit -= 1
		gp := GossipPacket{SearchReply: &reply}
		addr := g.FindPath(reply.Destination)
		err := g.SendTo(addr, gp)
		if err != nil {
			log.Error("Could not forward packet : ", err)
			return err
		}
	} else {

		log.Lvl2("Got search reply !!")
		g.PrintMatch(reply)

		for _, result := range reply.Results {
			log.Lvl2("result : ", result.FileName)
			//check if this file has already been found
			alreadyFound := false
			for _, found := range g.FoundFiles {
				if found.Filename == result.FileName && bytes.Equal(found.MetaHash, result.MetafileHash) && found.OriginChunks[1] == reply.Origin {
					log.Lvl2("Already found file ignoring")
					alreadyFound = true
					break
				}
			}
			if alreadyFound {
				//already found
				continue
			}
			//check if already in our current searches

			foundMatch := false
			for _, search := range g.CurrentlySearching.values {
				if bytes.Equal(search.MetaHash, result.MetafileHash) && search.Filename == result.FileName {
					log.Lvl2("Already existing search value.")

					foundMatch = true
					for _, b := range result.ChunkMap {
						search.OriginChunks[b] = reply.Origin
					}

					//check if file is totally downloaded
					if uint64(len(search.OriginChunks)) == search.ChunkCount {
						//we are done with this file
						log.Lvl2("Search for this is done ", search.Filename)
						//g.DownloadResult(search)
						search.Done = 1

					}
				}
			}

			if !foundMatch {
				log.Lvl2("New reply we add it and init it.")
				//if no match we initilaize it
				origins := make(map[uint64]string)
				for _, b := range result.ChunkMap {
					origins[b] = reply.Origin
				}
				ff := FoundFiles{
					Filename:     result.FileName,
					MetaHash:     result.MetafileHash,
					OriginChunks: origins,
					ChunkCount:   result.ChunkCount,
				}
				if uint64(len(result.ChunkMap)) == ff.ChunkCount {
					ff.Done = 1
				}

				log.Lvl2("adding it to wait for further blocks..")
				g.CurrentlySearching.Lock()
				g.CurrentlySearching.values = append(g.CurrentlySearching.values, ff)
				g.CurrentlySearching.Unlock()

			}
		}
		g.RemoveFinishedSearches()

	}
	return nil
}

//RemoveFinishedSearches removes the finished searches from the currently downloading structure
func (g *Gossiper) RemoveFinishedSearches() {
	//remove done searches.
	log.Lvl2("removing the finished searches...")
	var dones []FoundFiles
	g.CurrentlySearching.Lock()
	defer g.CurrentlySearching.Unlock()
	for _, elem := range g.CurrentlySearching.values {
		if elem.Done == 1 {
			log.Lvl2("removing because done : ", elem)

			g.UpdateMatches(elem)
			g.FoundFiles = append(g.FoundFiles, &elem)
		} else {
			dones = append(dones, elem)
		}
	}

	g.CurrentlySearching.values = dones
}

//UpdateMatches update the matches of a found file.
func (g *Gossiper) UpdateMatches(file FoundFiles) {
	//check matches
	log.Lvl2("updateing matches : ", file.Filename)
	for k, v := range g.SearchMatches {
		log.Lvl2("Checking for : ", k, v)
		words := strings.Split(k, ",")
		for _, w := range words {

			val, _ := path.Match("*"+w+"*", file.Filename)
			if val {
				log.Lvl2("match updated : ", k, ", ", v)
				g.SearchMatches[k]++
				break
			}
		}
	}
}
