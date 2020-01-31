//@authors Hrusanov Aleksandar, Lanzrein Johan, Rinaldi Vincent

package gossiper

//All the function for the confirmation protocol
import (
	"encoding/hex"
	"go.dedis.ch/onet/log"
	"math/rand"
	"time"
)

//StartConfirmation start the confirmation for a transaction txp . can be done either for ex2 (ackAll) or ex3
func (g *Gossiper) StartConfirmation(txp TxPublish) {

	log.Lvl2("Starting confirmation for file ", txp.Name)

	g.RunningConfirmation = true

	hash := [32]byte{}
	block := BlockPublish{
		PrevHash:    hash,
		Transaction: txp,
	}
	id := g.counter
	g.counter++
	sp := StatusPacket{g.PeerLogList()}
	message := TLCMessage{
		Origin:      g.Name,
		ID:          id,
		Confirmed:   -1,
		TxBlock:     block,
		VectorClock: &sp,
		Fitness:     0,
	}
	if g.hw3ex3 {
		//set the vector clock
		message.VectorClock = g.GetVectorClock()

	}
	g.Acknowledgements[id] = []string{g.Name}
	fingerprint := g.Name + message.TxBlock.Transaction.Name + hex.EncodeToString(message.TxBlock.Transaction.MetafileHash)

	g.TLCMessages[fingerprint] = &message
	gp := GossipPacket{TLCMessage: &message}
	g.SendToRandom(gp)
	for {
		finish := false

		select {
		case <-time.After(time.Second * time.Duration(int(g.stubbornTimeout))):
			//do the thing
			log.Lvl2("Sending it... we have : ", len(g.Acknowledgements[id]))
			if len(g.Acknowledgements[id]) >= (g.N+1)/2 {
				log.Lvl2("Got enough acks ! ")
				message.Confirmed = 1
				//send to all
				g.PrintRebroadcast(message, g.Acknowledgements[id])
				g.SendAcknowledgedTLC(&message)
				finish = true

			}
			g.SendToRandom(gp)

		case <-g.GiveUp:
			//take one of the confirmed messages of this round and broadcast it
			//not if we are running hw2ex2/ackAll this never gets called so its okay to have it.
			log.Lvl2("Give up now...")
			finish = true

		}
		if finish {
			log.Lvl2("finished with this..")
			g.FinishedConfirmation <- true
			break
		}
	}

	g.RunningConfirmation = false
}

//SendAck send an ack for the msg
func (g *Gossiper) SendAck(msg *TLCMessage) {
	log.Lvl2("Sending ack for : ", msg.Origin, " id ", msg.ID)

	ack := TLCAck{Origin: g.Name, ID: msg.ID, Destination: msg.Origin, HopLimit: g.HopLimit}
	g.PrintSendAck(ack)

	gp := GossipPacket{Ack: &ack}

	addr := g.FindPath(msg.Origin)
	if addr == "" {
		log.Error("Unknown orrigin")
		return
	}
	err := g.SendTo(addr, gp)
	if err != nil {
		log.Error("Could not send ack to : ", msg.Origin, " :  ", err)
	}
}

//ReceiveAck receive an ack and handle it if necessary.
func (g *Gossiper) ReceiveAck(ack *TLCAck) {
	if ack.Destination != g.Name {
		addr := g.FindPath(ack.Destination)
		if addr == "" {
			//unknwon path...
			return
		}
		ack.HopLimit--
		gp := GossipPacket{Ack: ack}
		err := g.SendTo(addr, gp)
		if err != nil {
			log.Error("Could not send packet to : ", addr, ": ", err)
		}
	} else {
		///have a map of ID <-> number of acks
		if g.ackAll {
			log.Lvl2("Got an ack ...")
			g.Acknowledgements[ack.ID] = append(g.Acknowledgements[ack.ID], ack.Origin)
		} else if g.hw3ex3 {
			log.Lvl2("Got an ack..")
			g.Acknowledgements[ack.ID] = append(g.Acknowledgements[ack.ID], ack.Origin)
		}
	}
}

//SendAcknowledgedTLC - send a msg that is confirmed to a random host.
func (g *Gossiper) SendAcknowledgedTLC(msg *TLCMessage) {
	log.Lvl2("Sending acked tlc")
	if msg.Confirmed < 0 {
		log.Error("Trying to send an unacknowledge block. Aborting")
		return
	}
	//forwards the message.
	gp := GossipPacket{TLCMessage: msg}

	g.SendToRandom(gp)

}

//ReceiveTLCMessage receive and treat a TLCMessage.
func (g *Gossiper) ReceiveTLCMessage(msg *TLCMessage) {
	fingerprint := msg.TxBlock.Transaction.Name + hex.EncodeToString(msg.TxBlock.Transaction.MetafileHash)

	reference, exists := g.TLCMessages[fingerprint]
	currId := msg.ID

	if msg.Confirmed > 0 {
		//its a confirmed message jsut store it
		gp := GossipPacket{TLCMessage: msg}

		fresh := g.VerifyFreshness(gp)
		if fresh {
			log.Lvl3("Got a new confirmed tlc ")
			g.PrintConfirmedGossip(*msg)
			//this also update the log of mytimes of the other persons...
			cont := g.UpdatePeerLog(gp)
			if cont {

				g.TLCMessages[fingerprint] = msg
				g.PreviousHash = msg.TxBlock.Transaction.MetafileHash[:]
				g.ConfirmedMessages.Lock()
				g.ConfirmedMessages.values[currId] = append(g.ConfirmedMessages.values[currId], msg)
				g.ConfirmedMessages.Unlock()

			}

			msg.Confirmed = int(msg.ID)
			msg.VectorClock = new(StatusPacket)
			msg.VectorClock.Want = g.PeerLogList()
			g.SendToRandom(gp)

			//check if can advance to next round...
			if g.hw3ex3 && len(g.ConfirmedMessages.values[currId]) >= (g.N+1)/2 {
				//advance to next round.
				// check 1)
				if g.HasReceivedConfirmedOwn() || g.RunningConfirmation {
					g.AdvanceToNextRound()
				}

			}
		}

		return
	}

	log.Lvl2("Got unconfirmed tlc..")

	if exists {
		//already seen we dont store it and pass
		//broadcast further because it has not yet been confirmed.
		if reference.Confirmed < 0 {
			rand.Seed(time.Now().Unix())
			if rand.Int()%2 == 0 {
				log.Lvl2("send it again..:")

				gp := GossipPacket{TLCMessage: msg}
				g.SendToRandom(gp)
			}

		}

		return

	}
	g.PrintUnconfirmedGossip(*msg)

	if g.ackAll && !exists {
		//always accept
		log.Lvl2("new one we store it and ack it..")

		g.TLCMessages[fingerprint] = msg
		//send ack
		if msg.Origin != g.Name {
			g.SendAck(msg)

		}
		//broadcast further
		gp := GossipPacket{TLCMessage: msg}
		g.SendToRandom(gp)

	} else if g.hw3ex3 && !exists {
		times := g.TimeMapping.times[msg.Origin]
		timeOfMsg := indexOf(times, msg.ID)
		if len(times) > 0 {
			idOfLast := times[len(times)-1]
			if idOfLast > currId {
				log.Lvl2("Message from past..ignoring it")
				return
			}
		}

		if timeOfMsg == g.my_time {
			log.Lvl2("This msg is at my_time..")

			log.Lvl2("Unseen message.")
			g.TLCMessages[fingerprint] = msg
			if msg.Origin != g.Name {
				g.SendAck(msg)

			}
			//broadcast further
			gp := GossipPacket{TLCMessage: msg}

			g.SendToRandom(gp)
		}

	}

}

func indexOf(xs []uint32, val uint32) uint32 {
	for i, x := range xs {
		if x == val {
			return uint32(i)
		}
	}
	//its not there yet. so we assume its unknown
	return uint32(len(xs))
}

//AdvanceToNextRound tries to advance to the next round.
func (g *Gossiper) AdvanceToNextRound() {
	log.Lvl2("Trying to advance to next round ! ")
	witness := g.confirmationForRound()
	if (len(witness)) >= (g.N+1)/2 {
		log.Lvl2("Can advance")
		g.my_time++
		g.PrintAdvanceToNextRound(witness)

		g.GiveUp <- true
	}

}

func (g *Gossiper) confirmationForRound() []*TLCMessage {
	my_time := g.my_time
	g.TimeMapping.Lock()
	defer g.TimeMapping.Unlock()
	var confirmations []*TLCMessage
	for key, val := range g.TimeMapping.times {
		if uint32(len(val)) == my_time+1 {
			tlc := g.pl.logmap[key][val[my_time]-1].TLCMessage
			confirmations = append(confirmations, tlc)
		}
	}

	return confirmations
}

//AddToStack add a transaction to the stack of pending transaction.
func (g *Gossiper) AddToStack(file string, metahash []byte, size int64) {

	log.Lvl2("adding to stack : ", file)
	txp := TxPublish{
		Name:         file,
		Size:         size,
		MetafileHash: metahash,
	}
	g.StackConfirmation = append(g.StackConfirmation, &txp)
	g.PopStack()

}

//PopStack pop the stack of pending transactions.
func (g *Gossiper) PopStack() {
	log.Lvl2("Popping stack ! ")
	if g.RunningConfirmation {
		log.Lvl2("Have to wait..:( ")
		//wait until the confirmation is finished.
		//note this is okay if many thread wait here as only one thread can "take" from the chan.
		<-g.FinishedConfirmation

	}
	if len(g.StackConfirmation) > 0 {
		txp := g.StackConfirmation[0]
		g.RunningConfirmation = true
		g.StackConfirmation = g.StackConfirmation[1:]
		log.Lvl2("Proceeding from popstack. ")
		go g.StartConfirmation(*txp)
	}
}

//HasReceivedConfirmedOwn has already confirmed his own round
func (g *Gossiper) HasReceivedConfirmedOwn() bool {
	return (len(g.TimeMapping.times[g.Name])) >= (g.N+1)/2
}
