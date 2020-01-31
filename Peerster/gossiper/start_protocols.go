//@authors Hrusanov Aleksandar, Lanzrein Johan, Rinaldi Vincent

package gossiper

import (
	"go.dedis.ch/onet/log"
	"strings"
	"time"
)

//StartFileDownload start the file download for a file in the message.
func (g *Gossiper) StartFileDownload(message Message) {
	replies := make(chan DataReply)

	go g.FileSharingReceiveProtocol(*message.Destination, *message.Request, replies, *message.File, nil)
	return

}

//StartFileSearch starts a new file search from this gossiper
func (g *Gossiper) StartFileSearch(keywords []string, budget uint64, mulFactor int) error {
	stringKeywords := strings.Join(keywords, ",")
	//initialize the SearchMatches
	g.SearchMatches[stringKeywords] = 0

	req := SearchRequest{
		Origin:   g.Name,
		Budget:   budget,
		Keywords: keywords,
	}
	log.Lvl2("Filesearch for keywords : ", stringKeywords, "budget : ", budget, " mulfactor : ", mulFactor)

	for budget <= MAXBUDGET || mulFactor == 1 {
		//gp := GossipPacket{SearchRequest:&req}
		log.Lvl3("Sending req with budget : ", budget)
		err := g.ReceiveSearchRequest(&req)
		if err != nil {
			log.Error("Could not start search request : ", err)
			return err
		}

		<-time.After(time.Second)

		log.Lvl2("WE HAVE : matches amt", g.SearchMatches[stringKeywords])
		if g.SearchMatches[stringKeywords] >= MATCHTHRESHOLD {
			//found all files !
			g.PrintSearchFinish(true)
			break
		}
		if budget == 32 || mulFactor < 2 {
			g.PrintSearchFinish(false)
			break
		}

		budget *= uint64(mulFactor)

		req.Budget = budget

	}

	return nil

}
