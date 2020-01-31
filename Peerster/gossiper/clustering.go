//clustering file for the handling of a cluster by  agossiper
//@authors Hrusanov Aleksandar, Lanzrein Johan, Rinaldi Vincent

package gossiper

import (
	"bytes"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/JohanLanzrein/Peerster/clusters"
	"github.com/JohanLanzrein/Peerster/ies"
	"go.dedis.ch/onet/log"
	"go.dedis.ch/protobuf"
)

//Constant values
const DEFAULTROLLOUT = 300 // WAS 300
const DEFAULTHEARTBEAT = 5

var ClusterUpdated = make(chan (bool), 1)

//InitCounter the current gossiper creates a cluster where he is the sole member
func (g *Gossiper) InitCluster() {
	id := GenerateId()
	seed := rand.Int63()
	members := []string{g.Name}
	publickey := make(map[string]ies.PublicKey)
	publickey[g.Name] = g.Keypair.PublicKey
	masterkey := g.MasterKeyGen()
	authorities := []string{g.Name}
	cluster := clusters.NewCluster(id, members, masterkey, publickey, uint64(seed), authorities)

	g.Cluster = &cluster
	g.PrintInitCluster()
	go g.HeartbeatLoop()
}

//RequestJoining Sends a packet to request joining a cluster from other
func (g *Gossiper) RequestJoining(other string) {
	//send a request packet to the other gossiper
	log.Lvl2("Joining request :", other)
	addr := g.FindPath(other)
	if addr ==""{
		log.Warn("Unexisting path !!")
		return
	}
	publickey := g.Keypair.PublicKey
	req := RequestMessage{
		Origin:    g.Name,
		Recipient: other,
		PublicKey: publickey,
	}

	gp := GossipPacket{JoinRequest: &req}
	go g.SendTo(addr, gp)
	//then the "voting" system starts

}

//RequestExpelling Sends a packet to request expelling other from the cluster
func (g *Gossiper) RequestExpelling(other string) {
	//send a request packet to the whole cluster
	log.Lvl1("Expelling request :" , other)
	publickey := g.Keypair.PublicKey
	req := RequestMessage{
		Origin:    g.Name,
		Recipient: other,
		PublicKey: publickey,
	}
	g.pending_nodes_requests = append(g.pending_nodes_requests, "EXPEL " + other)
	g.pending_messages_requests = append(g.pending_messages_requests, req)
	g.BroadcastExpel(other)
	//then the "voting" system starts

}

//HeartbeatLoop
func (g *Gossiper) HeartbeatLoop() {
	dur := time.Duration(g.RolloutTimer) * time.Second
	timer := time.NewTimer(dur)

	for {

		select {
		case <-time.After(time.Duration(g.HearbeatTimer) * time.Second):
			log.Lvl2(g.Name, "sending heartbeat")
			g.Cluster.HeartBeats[g.Name] = true
			go g.SendBroadcast("", false)

		case <-g.LeaveChan:
			log.Lvl2("Leaving cluster")
			return
		case <-timer.C:
			log.Lvl2("Time for a rolllllllllout")
			idx := g.Cluster.Clock()
			log.Lvl2(g.Name, "My idx is :", idx)
			sort.Strings(g.Cluster.Members)
			g.KeyRollout(g.Cluster.Members[idx])
			timer.Reset(dur)
		}
	}
}

//LeaveCluster stops the heartbeat loop and resets the value of cluster
func (g *Gossiper) LeaveCluster() {
	//Stop the heartbeat loop
	g.PrintLeaveCluster(*g.Cluster.ClusterID)
	g.LeaveChan <- true
	log.Lvl2("Sending leave message..")
	g.RequestLeave()

	g.Cluster = nil
	//Send a message saying we want to leave.
	return
}

//SendBroadcast with thte given text. if the leave flag is set will also be a request to leave.
func (g *Gossiper) SendBroadcast(text string, leave bool) {
	rumor := RumorMessage{
		Origin: g.Name,
		ID:     0,
		Text:   text,
	}
	if text != "" {
		g.PrintBroadcast(rumor)
	}

	data, err := protobuf.Encode(&rumor)
	if err != nil {
		log.Error("Could not encode the packet.. ", err)
		return
	}

	enc := ies.Encrypt(g.Cluster.MasterKey, data)
	bm := BroadcastMessage{
		ClusterID:    *g.Cluster.ClusterID,
		HopLimit:     g.HopLimit,
		Destination:  "",
		Data:         enc,
		LeaveRequest: leave,
	}
	gp := GossipPacket{Broadcast: &bm}
	g.Cluster.HeartBeats[g.Name] = true

	//Send to all member of the cluster.
	//This does not need to be anonymized as an attacker can in any case know who is in a cluster by joining it..
	for _, m := range g.Cluster.Members {
		if m == g.Name {
			continue
		}
		bm.Destination = m
		addr := g.FindPath(m)
		if addr == "" {
			continue
		}
		err := g.SendTo(addr, gp)
		if err != nil {
			log.Error("Error while sending to ", m, " : ", err)
		}
	}
}

//ReveiceBroadcast handles a broadcast message and displays it if possible otherwise sends it further
func (g *Gossiper) ReceiveBroadcast(message BroadcastMessage) {
	if message.Destination != g.Name {
		//Send it further.
		message.HopLimit--
		log.Lvl2(g.Name, "forwarding to ", message.Destination)
		if message.HopLimit > 0 {
			addr := g.FindPath(message.Destination)
			if addr == "" {
				log.Error(g.Name, "Could not find path to ", message.Destination)
				return
			}
			gp := GossipPacket{Broadcast: &message}
			g.SendToRandom(gp)
		}
		return
	}
	if g.Cluster != nil && message.ClusterID == *g.Cluster.ClusterID {
		log.Lvl2("Got broadcast for my cluster")

		if message.Rollout {
			//Update for a rollout.
			log.Lvl1(g.Name, " received message for rollout")
			cluster := clusters.Cluster{}
			data := ies.Decrypt(g.Cluster.MasterKey, message.Data)
			err := protobuf.Decode(data, &cluster)
			if err != nil {
				log.Error("Could not decode rollout info ", err)
			}
			ClusterUpdated <- true
			g.UpdateFromRollout(cluster)

		} else if message.Reset {
			log.Lvl1(g.Name, " received message for reset")
			cluster := clusters.Cluster{}
			data := ies.Decrypt(g.Cluster.MasterKey, message.Data)
			err := protobuf.Decode(data, &cluster)
			if err != nil {
				log.Error("Could not decode reset info ", err)
			}

			g.UpdateFromReset(cluster)
			g.BroadcastAck(message.CaseRequest)

		} else if message.LeaveRequest {
			log.Lvl1("Got leave request")
			decrypted := ies.Decrypt(g.Cluster.MasterKey, message.Data)
			var rumor RumorMessage
			err := protobuf.Decode(decrypted, &rumor)
			if err != nil {
				log.Error(g.Name, "Error decoding packet : ", err, "This may be due to an ongoing rollout.")
				return
			}

			g.Cluster.HeartBeats[rumor.Origin] = false
			delete(g.Cluster.PublicKeys, rumor.Origin)
			g.Cluster.Members = RemoveFromList(g.Cluster.Members, rumor.Origin)

		} else if message.JoinRequest {
			log.Lvl1(g.Name, " received message for JOIN e-voting request")
			decrypted := ies.Decrypt(g.Cluster.MasterKey, message.Data)
			var rumor RumorMessage
			err := protobuf.Decode(decrypted, &rumor)
			if err != nil {
				log.Error(g.Name, "Error decoding packet : ", err, "This may be due to an ongoing rollout.")
				return
			}
			if rumor.Text != "" {
				//print the message
				g.PrintEvotingJoinStep(rumor.Text)
				g.displayed_requests = append(g.displayed_requests, "JOIN " + rumor.Text)
				if g.Cluster.IsAnAuthority(g.Name) {
					slice_pending := make([]string, 0)
					tag := "JOIN " + rumor.Text
					slice_pending = append(slice_pending, tag)
					g.slice_results = append(g.slice_results, slice_pending)
				}
			}
			//in any case add it to the map..
			if g.Name != rumor.Origin {
				g.Cluster.HeartBeats[rumor.Origin] = true
			}

		} else if message.ExpelRequest {
			log.Lvl1(g.Name, " received message for EXPEL e-voting request")
			decrypted := ies.Decrypt(g.Cluster.MasterKey, message.Data)
			var rumor RumorMessage
			err := protobuf.Decode(decrypted, &rumor)
			if err != nil {
				log.Error(g.Name, "Error decoding packet : ", err, "This may be due to an ongoing rollout.")
				return
			}
			if rumor.Text != "" && rumor.Text != g.Name {
				//print the message
				g.PrintEvotingExpelStep(rumor.Text)
				g.displayed_requests = append(g.displayed_requests, "EXPEL " + rumor.Text)
				if g.Cluster.IsAnAuthority(g.Name) {
					slice_pending := make([]string, 0)
					tag := "EXPEL " + rumor.Text
					slice_pending = append(slice_pending, tag)
					g.slice_results = append(g.slice_results, slice_pending)
				}
			}
			//in any case add it to the map..
			if g.Name != rumor.Origin {
				g.Cluster.HeartBeats[rumor.Origin] = true
			}

		} else if message.AcceptProposition {
			log.Lvl1(g.Name, " received ACCEPT for e-voting case")
			decrypted := ies.Decrypt(g.Cluster.MasterKey, message.Data)
			var rumor RumorMessage
			err := protobuf.Decode(decrypted, &rumor)
			if err != nil {
				log.Error(g.Name, "Error decoding packet : ", err, "This may be due to an ongoing rollout.")
				return
			}
			log.Lvl1(g.Cluster.Authorities)
			if g.Cluster.IsAnAuthority(g.Name) {
				log.Lvl1("rumor txtd", rumor.Text)
				if rumor.Text != "" {
					//print the message
					g.PrintEvotingPropositionStep(1, rumor.Text, rumor.Origin)
					correct_tag_join := "JOIN " + rumor.Text
					correct_tag_expel := "EXPEL " + rumor.Text
					log.Lvl1(g.slice_results)
					for i := 0 ; i < len(g.slice_results) ; i++ {
						if string((g.slice_results[i])[0]) == correct_tag_join {
							is_existing := false
							for j := 1 ; j < len(g.slice_results[i]) ; j++ {
								if string((g.slice_results[i])[j]) == "1 : " + rumor.Origin || string((g.slice_results[i])[j]) == "0 : " + rumor.Origin {
									is_existing = true

									break
								}
							}
							if is_existing == false {
								g.slice_results[i] = append(g.slice_results[i], "1 : " + rumor.Origin)

								if len(g.Cluster.Members) == len(g.slice_results[i]) - 1 {
									g.BroadcastCollected(g.slice_results[i][0])
								}
							}
							break
						} else if string((g.slice_results[i])[0]) == correct_tag_expel && rumor.Text != rumor.Origin && rumor.Text != g.Name {
							is_existing := false
							for j := 1 ; j < len(g.slice_results[i]) ; j++ {
								if string((g.slice_results[i])[j]) == "1 : " + rumor.Origin || string((g.slice_results[i])[j]) == "0 : " + rumor.Origin {
									is_existing = true

									log.Lvl1("Already existing", g.slice_results)

									break
								}
							}
							if is_existing == false {
								g.slice_results[i] = append(g.slice_results[i], "1 : " + rumor.Origin)

								log.Lvl1(g.slice_results)

								if len(g.Cluster.Members) - 1 == len(g.slice_results[i]) - 1 {
									g.BroadcastCollected(g.slice_results[i][0])
								}
							}
							break
						}
					}
				}
			}
			//in any case add it to the map..
			if g.Name != rumor.Origin {
				g.Cluster.HeartBeats[rumor.Origin] = true
			}

		} else if message.DenyProposition {
			log.Lvl1(g.Name, " received DENY for e-voting case")
			decrypted := ies.Decrypt(g.Cluster.MasterKey, message.Data)
			var rumor RumorMessage
			err := protobuf.Decode(decrypted, &rumor)
			if err != nil {
				log.Error(g.Name, "Error decoding packet : ", err, "This may be due to an ongoing rollout.")
				return
			}
			if g.Cluster.IsAnAuthority(g.Name) {
				if rumor.Text != "" {
					//print the message
					g.PrintEvotingPropositionStep(0, rumor.Text, rumor.Origin)
					correct_tag_join := "JOIN " + rumor.Text
					correct_tag_expel := "EXPEL " + rumor.Text
					for i := 0 ; i < len(g.slice_results) ; i++ {
						if string((g.slice_results[i])[0]) == correct_tag_join {
							is_existing := false
							for j := 0 ; j < len(g.slice_results[i]) ; j++ {
								if string((g.slice_results[i])[j]) == "1 : " + rumor.Origin || string((g.slice_results[i])[j]) == "0 : " + rumor.Origin {
									is_existing = true

									break
								}
							}
							if is_existing == false {
								g.slice_results[i] = append(g.slice_results[i], "0 : " + rumor.Origin)

								if len(g.Cluster.Members) == len(g.slice_results[i]) - 1 {
									g.BroadcastCollected(g.slice_results[i][0])
								}
							}
							break
						} else if string((g.slice_results[i])[0]) == correct_tag_expel && rumor.Text != rumor.Origin && rumor.Text != g.Name {
							is_existing := false
							for j := 1 ; j < len(g.slice_results[i]) ; j++ {
								if string((g.slice_results[i])[j]) == "1 : " + rumor.Origin || string((g.slice_results[i])[j]) == "0 : " + rumor.Origin {
									is_existing = true

									//fmt.Println("Already existing", g.slice_results)

									break
								}
							}
							if is_existing == false {
								g.slice_results[i] = append(g.slice_results[i], "0 : " + rumor.Origin)

								//fmt.Println(g.slice_results)

								if len(g.Cluster.Members) - 1 == len(g.slice_results[i]) - 1 {
									g.BroadcastCollected(g.slice_results[i][0])
								}
							}
							break
						}
					}
				}
			}
			//in any case add it to the map..
			if g.Name != rumor.Origin {
				g.Cluster.HeartBeats[rumor.Origin] = true
			}

		} else if message.CaseCompare {
			log.Lvl1(g.Name, " received e-voting identifier to start comparison process")
			decrypted := ies.Decrypt(g.Cluster.MasterKey, message.Data)
			var rumor RumorMessage
			err := protobuf.Decode(decrypted, &rumor)
			if err != nil {
				log.Error(g.Name, "Error decoding packet : ", err, "This may be due to an ongoing rollout.")
				return
			}
			if g.Cluster.IsAnAuthority(g.Name) {
				if rumor.Text != "" {
					//print the message
					g.PrintEvotingCaseStep(rumor.Text, rumor.Origin)

					if string(rumor.Text[0:4]) == "JOIN" {
						_, ok := g.acks_cases[rumor.Text]
						if ok == true {
							is_existing := false
							for i := 0 ; i < len(g.acks_cases[rumor.Text]) ; i++ {
								if g.acks_cases[rumor.Text][i] == rumor.Origin {
									is_existing = true
									break
								}

							}
							if is_existing == false {
								if g.Cluster.IsAnAuthority(rumor.Origin) {
									g.acks_cases[rumor.Text] = append(g.acks_cases[rumor.Text], rumor.Origin)
								}
							}
						} else { // ok == false
							if g.Cluster.IsAnAuthority(rumor.Origin) {
								g.acks_cases[rumor.Text] = []string{rumor.Origin}
							}
						}

						if g.Cluster.AmountAuthorities() == len(g.acks_cases[rumor.Text]) {
							for j := 0 ; j < len(g.slice_results) ; j++ {
								if string((g.slice_results[j])[0]) == rumor.Text {
									if (g.Cluster.AmountAuthorities() > 1) {
										g.BroadcastResults(g.slice_results[j])
									} else { // g.Cluster.AmountAuthorities() == 1
										accept_counts := 0
										deny_counts := 0
										for k := 1 ; k < len(g.slice_results[j]) ; k++ {
											str := string(((g.slice_results[j])[k])[0])
											if str == "1" {
												accept_counts++
											} else if str == "0" {
												deny_counts++
											}

										}
										if deny_counts == accept_counts {
											accept_counts = 0
											deny_counts = 0
											for l := 1 ; l < len(g.slice_results[j]) ; l++ {
												str := (g.slice_results[j])[l]
												list_authorities := g.acks_cases[rumor.Text]
												for m := 0 ; m < len(list_authorities) ; m++ {
													if string(str[4:]) == list_authorities[m] {
														str = string(str[0])
														if str == "1" {
															accept_counts++
														} else if str == "0" {
															deny_counts++
														}
														break
													}
												}
											}
										}
										if accept_counts > deny_counts {
											answer := "ACCEPT " + rumor.Text
											g.BroadcastDecision(answer)
										} else { // accept_counts < deny_counts
											answer := "DENY " + rumor.Text
											g.BroadcastDecision(answer)
										}
									}
									break
								}
							}
						}
					} else if string(rumor.Text[0:5]) == "EXPEL" && string(rumor.Text[6:]) != g.Name {
						_, ok := g.acks_cases[rumor.Text]
						if ok == true {
							is_existing := false
							for i := 0 ; i < len(g.acks_cases[rumor.Text]) ; i++ {
								if g.acks_cases[rumor.Text][i] == rumor.Origin {
									is_existing = true
									break
								}
							}
							if is_existing == false {
								if g.Cluster.IsAnAuthority(rumor.Origin) {
									g.acks_cases[rumor.Text] = append(g.acks_cases[rumor.Text], rumor.Origin)
								}
							}
						} else { // ok == false
							if g.Cluster.IsAnAuthority(rumor.Origin) {
								g.acks_cases[rumor.Text] = []string{rumor.Origin}
							}
						}

						var maxToCollect int
						if g.Cluster.IsAnAuthority(string(rumor.Text[6:])) {
							maxToCollect = g.Cluster.AmountAuthorities() - 1
						} else { // g.Cluster.IsAnAuthority(string(rumor.Text[6:])) == false
							maxToCollect = g.Cluster.AmountAuthorities()
						}

						if maxToCollect == len(g.acks_cases[rumor.Text]) {
							for j := 0 ; j < len(g.slice_results) ; j++ {
								if string((g.slice_results[j])[0]) == rumor.Text {
									g.BroadcastResults(g.slice_results[j])
									break
								}
							}
						}
					}
				}
			}
			//in any case add it to the map..
			if g.Name != rumor.Origin {
				g.Cluster.HeartBeats[rumor.Origin] = true
			}

		} else if message.ResultsValidation {
			log.Lvl1(g.Name, " received list of e-voting results for validation")
			decrypted := ies.Decrypt(g.Cluster.MasterKey, message.Data)
			var rumor RumorMessage
			err := protobuf.Decode(decrypted, &rumor)
			if err != nil {
				log.Error(g.Name, "Error decoding packet : ", err, "This may be due to an ongoing rollout.")
				return
			}
			if g.Cluster.IsAnAuthority(g.Name) {
				if rumor.Results != nil {
					//print the message
					g.PrintEvotingValidationStep(rumor.Results, rumor.Origin)

					if string((rumor.Results[0])[0:4]) == "JOIN" {
						is_matched := true
						for i := 0 ; i < len(g.slice_results) ; i++ {
							if string((g.slice_results[i])[0]) == string(rumor.Results[0]) {
								for j := 1 ; j < len(g.slice_results[i]) ; j++ {
									found := false
									for k := 1 ; k < len(rumor.Results) ; k++ {
										if string((g.slice_results[i])[j]) == string(rumor.Results[k]) {
											found = true
											break
										}
									}
									if found == false {
										is_matched = false

										break
									}
								}
								break
							}
						}

						if is_matched == true {
							_, ok := g.correct_results_rcv[rumor.Results[0]]
							if ok == true {
								is_existing := false
								for i := 0 ; i < len(g.correct_results_rcv[rumor.Results[0]]) ; i++ {
									if g.correct_results_rcv[rumor.Results[0]][i] == rumor.Origin {
										is_existing = true
										break
									}

								}
								if is_existing == false {
									if g.Cluster.IsAnAuthority(rumor.Origin) {
										g.correct_results_rcv[rumor.Results[0]] = append(g.correct_results_rcv[rumor.Results[0]], rumor.Origin)
									}
								}
							} else { // ok == false
								if g.Cluster.IsAnAuthority(rumor.Origin) {
									g.correct_results_rcv[rumor.Results[0]] = []string{rumor.Origin}
								}
							}

							if g.Cluster.AmountAuthorities() - 1 == len(g.correct_results_rcv[rumor.Results[0]]) {
								accept_counts := 0
								deny_counts := 0
								for i := 0 ; i < len(g.slice_results) ; i++ {
									if string(g.slice_results[i][0]) == string(rumor.Results[0]) {
										for j := 1 ; j < len(g.slice_results[i]) ; j++ {
											str := string(((g.slice_results[i])[j])[0])
											if str == "1" {
												accept_counts++
											} else if str == "0" {
												deny_counts++
											}
										}
										break
									}
								}

								if deny_counts == accept_counts {
									accept_counts = 0
									deny_counts = 0
									for i := 0 ; i < len(g.slice_results) ; i++ {
										if string((g.slice_results[i])[0]) == string(rumor.Results[0]) {
											for j := 1 ; j < len(g.slice_results[i]) ; j++ {
												str := (g.slice_results[i])[j]
												list_authorities := g.acks_cases[rumor.Results[0]]
												for k := 0 ; k < len(list_authorities) ; k++ {
													if string(str[4:]) == list_authorities[k] {
														str = string(str[0])
														if str == "1" {
															accept_counts++
														} else if str == "0" {
															deny_counts++
														}
														break
													}
												}
											}
											break
										}
									}
								}

								if accept_counts > deny_counts {
									answer := "ACCEPT " + rumor.Results[0]
									g.BroadcastDecision(answer)
								} else { // accept_counts < deny_counts
									answer := "DENY " + rumor.Results[0]
									g.BroadcastDecision(answer)
								}
							}
						} else { // is_matched == false
							g.BroadcastCancel(rumor.Results[0])
						}
					} else if string((rumor.Results[0])[0:5]) == "EXPEL" && string((rumor.Results[0])[6:]) != g.Name {
						is_matched := true
						for i := 0 ; i < len(g.slice_results) ; i++ {
							if string((g.slice_results[i])[0]) == string(rumor.Results[0]) {
								for j := 1 ; j < len(g.slice_results[i]) ; j++ {
									found := false
									for k := 1 ; k < len(rumor.Results) ; k++ {
										if string((g.slice_results[i])[j]) == string(rumor.Results[k]) {
											found = true
											break

										}
									}
									if found == false {
										is_matched = false
										break
									}
								}
								break
							}
						}

						if is_matched == true {
							_, ok := g.correct_results_rcv[rumor.Results[0]]
							if ok == true {
								is_existing := false
								for i := 0 ; i < len(g.correct_results_rcv[rumor.Results[0]]) ; i++ {
									if g.correct_results_rcv[rumor.Results[0]][i] == rumor.Origin {
										is_existing = true
										break
									}
								}
								if is_existing == false {
									if g.Cluster.IsAnAuthority(rumor.Origin) {
										g.correct_results_rcv[rumor.Results[0]] = append(g.correct_results_rcv[rumor.Results[0]], rumor.Origin)
									}
								}
							} else { // ok == false
								if g.Cluster.IsAnAuthority(rumor.Origin) {
									g.correct_results_rcv[rumor.Results[0]] = []string{rumor.Origin}
								}
							}

							var maxToCollect int
							if g.Cluster.IsAnAuthority(string((rumor.Results[0])[6:])) {
								maxToCollect = g.Cluster.AmountAuthorities() - 2
							} else { // g.Cluster.IsAnAuthority(string((rumor.Results[0])[6:])) == false
								maxToCollect = g.Cluster.AmountAuthorities() - 1
							}

							if maxToCollect == len(g.correct_results_rcv[rumor.Results[0]]) {
								accept_counts := 0
								deny_counts := 0

								for i := 0 ; i < len(g.slice_results) ; i++ {
									if string(g.slice_results[i][0]) == string(rumor.Results[0]) {
										for j := 1 ; j < len(g.slice_results[i]) ; j++ {
											str := string(((g.slice_results[i])[j])[0])
											if str == "1" {
												accept_counts++
											} else if str == "0" {
												deny_counts++
											}
										}
										break
									}
								}

								if deny_counts == accept_counts {
									accept_counts = 0
									deny_counts = 0
									for i := 0 ; i < len(g.slice_results) ; i++ {
										if string((g.slice_results[i])[0]) == string(rumor.Results[0]) {
											for j := 1 ; j < len(g.slice_results[i]) ; j++ {
												str := (g.slice_results[i])[j]
												list_authorities := g.acks_cases[rumor.Results[0]]
												for k := 0 ; k < len(list_authorities) ; k++ {
													if string(str[4:]) == list_authorities[k] {
														str = string(str[0])
														if str == "1" {
															accept_counts++
														} else if str == "0" {
															deny_counts++
														}
														break
													}
												}
											}
											break
										}
									}
								}

								//fmt.Println("Accept counts =", accept_counts, "and deny counts =", deny_counts)

								if accept_counts > deny_counts {
									answer := "ACCEPT " + rumor.Results[0]
									g.BroadcastDecision(answer)
								} else { // accept_counts < deny_counts OR accept_counts == deny_counts
									answer := "DENY " + rumor.Results[0]
									g.BroadcastDecision(answer)
								}
							}
						} else { // is_matched == false
							g.BroadcastCancel(rumor.Results[0])
						}

					}
				}
			}
			//in any case add it to the map..
			if g.Name != rumor.Origin {
				g.Cluster.HeartBeats[rumor.Origin] = true
			}

		} else if message.FinalDecision {
			log.Lvl1(g.Name, " received e-voting final decision")
			if g.Cluster != nil {
				decrypted := ies.Decrypt(g.Cluster.MasterKey, message.Data)
				var rumor RumorMessage
				err := protobuf.Decode(decrypted, &rumor)
				if err != nil {
					log.Error(g.Name, "Error decoding packet : ", err, "This may be due to an ongoing rollout.")

				}
				if rumor.Text != "" && strings.Contains(rumor.Text, g.Name) == false {
					//print the message
					g.PrintEvotingDecisionStep(rumor.Text, rumor.Origin)
					if g.Cluster.IsAnAuthority(g.Name) {
						for i := 0 ; i < len(g.slice_results) ; i++ {
							if strings.Contains(rumor.Text, string((g.slice_results[i])[0])) {
								copy(g.slice_results[i:], g.slice_results[i+1:])
								g.slice_results = g.slice_results[:len(g.slice_results) - 1]
								break
							}
						}
						if strings.Contains(rumor.Text, "ACCEPT") {
							str := string(rumor.Text[7:])
							_, ok1 := g.acks_cases[str]
							if ok1 {
								delete(g.acks_cases, str)
							}
							_, ok2 := g.correct_results_rcv[str]
							if ok2 {
								delete(g.correct_results_rcv, str)
							}
						} else { // strings.Contains(rumor.Text, "DENY") == true
							str := string(rumor.Text[5:])
							_, ok1 := g.acks_cases[str]
							if ok1 {
								delete(g.acks_cases, str)
							}
							_, ok2 := g.correct_results_rcv[str]
							if ok2 {
								delete(g.correct_results_rcv, str)
							}
						}
					}

					if strings.Contains(rumor.Text, "ACCEPT EXPEL")  {
						g.UpdateClusterExpel(rumor.Text[13:])
					}

					found_request := false
					idx_request := -1
					for i := 0 ; i < len(g.pending_nodes_requests) ; i++ {
						if strings.Contains(rumor.Text, g.pending_nodes_requests[i]) {
							found_request = true
							idx_request = i
							break
						}
					}
					if found_request == true {
						if strings.Contains(g.pending_nodes_requests[idx_request], "JOIN ") {
							var msg RequestMessage
							for j := 0 ; j < len(g.pending_messages_requests) ; j++ {
								if (g.pending_messages_requests[j]).Origin == string((g.pending_nodes_requests[idx_request])[5:]) {
									msg = g.pending_messages_requests[j]
									break
								}
							}

							for i := 0 ; i < len(g.pending_nodes_requests) ; i++ {
								if string(g.pending_nodes_requests[i]) == "JOIN " + msg.Origin {
									copy(g.pending_nodes_requests[i:], g.pending_nodes_requests[i+1:])
									g.pending_nodes_requests = g.pending_nodes_requests[:len(g.pending_nodes_requests) - 1]
									break
								}
							}


							for i := 0 ; i < len(g.pending_messages_requests) ; i++ {
								if (g.pending_messages_requests[i]).Origin == msg.Origin && (g.pending_messages_requests[i]).Recipient == msg.Recipient && bytes.Compare(g.pending_messages_requests[i].PublicKey, msg.PublicKey ) == 0 {
									copy(g.pending_messages_requests[i:], g.pending_messages_requests[i+1:])
									g.pending_messages_requests = g.pending_messages_requests[:len(g.pending_messages_requests) - 1]
									break
								}
							}

							var final_decision int
							if (strings.Contains(rumor.Text, "ACCEPT ")) {
								final_decision = 1
							} else { // strings.Contains(rumor.Text, "DENY ") == true
								final_decision = 0
							}
							g.ReceiveDecisionJoinRequest(msg, final_decision)

						} else if strings.Contains(g.pending_nodes_requests[idx_request], "EXPEL ") {
							var msg RequestMessage
							for j := 0 ; j < len(g.pending_messages_requests) ; j++ {
								if (g.pending_messages_requests[j]).Recipient == string((g.pending_nodes_requests[idx_request])[6:]) {
									msg = g.pending_messages_requests[j]
									break
								}
							}

							for i := 0 ; i < len(g.pending_nodes_requests) ; i++ {
								if string(g.pending_nodes_requests[i]) == "EXPEL " + msg.Recipient {
									copy(g.pending_nodes_requests[i:], g.pending_nodes_requests[i+1:])
									g.pending_nodes_requests = g.pending_nodes_requests[:len(g.pending_nodes_requests) - 1]
									break
								}
							}

							for i := 0 ; i < len(g.pending_messages_requests) ; i++ {
								if (g.pending_messages_requests[i]).Origin == msg.Origin && (g.pending_messages_requests[i]).Recipient == msg.Recipient && bytes.Compare(g.pending_messages_requests[i].PublicKey, msg.PublicKey) == 0 {
									copy(g.pending_messages_requests[i:], g.pending_messages_requests[i+1:])
									g.pending_messages_requests = g.pending_messages_requests[:len(g.pending_messages_requests) - 1]
									break
								}
							}

							var final_decision int
							if (strings.Contains(rumor.Text, "ACCEPT ")) {
								final_decision = 1
							} else { // strings.Contains(rumor.Text, "DENY ") == true
								final_decision = 0
							}
							g.ReceiveDecisionExpelRequest(msg, final_decision)

						}
					}
				}
				//in any case add it to the map..
				if g.Name != rumor.Origin {
					g.Cluster.HeartBeats[rumor.Origin] = true
				}
			}

		} else if message.CancelRequest {
			log.Lvl1(g.Name, " received e-voting cancellation")
			decrypted := ies.Decrypt(g.Cluster.MasterKey, message.Data)
			var rumor RumorMessage
			err := protobuf.Decode(decrypted, &rumor)
			if err != nil {
				log.Error(g.Name, "Error decoding packet : ", err, "This may be due to an ongoing rollout.")
				return
			}
			if rumor.Text != "" && strings.Contains(rumor.Text, g.Name) == false {
				//print the message
				g.PrintEvotingCancellationStep(rumor.Text, rumor.Origin)
				if g.Cluster.IsAnAuthority(g.Name) {
					for i := 0 ; i < len(g.slice_results) ; i++ {
						if rumor.Text == string((g.slice_results[i])[0]) {
							copy(g.slice_results[i:], g.slice_results[i+1:])
							g.slice_results = g.slice_results[:len(g.slice_results) - 1]
							break
						}
					}
					_, ok1 := g.acks_cases[rumor.Text]
					if ok1 {
						delete(g.acks_cases, rumor.Text)
					}
					_, ok2 := g.correct_results_rcv[rumor.Text]
					if ok2 {
						delete(g.correct_results_rcv, rumor.Text)
					}
				}
				g.BroadcastReset(rumor.Text)
			}
			//in any case add it to the map..
			if g.Name != rumor.Origin {
				g.Cluster.HeartBeats[rumor.Origin] = true
			}

		} else if message.ResetIndication {
			log.Lvl1(g.Name, " received e-voting reset")
			decrypted := ies.Decrypt(g.Cluster.MasterKey, message.Data)
			var rumor RumorMessage
			err := protobuf.Decode(decrypted, &rumor)
			if err != nil {
				log.Error(g.Name, "Error decoding packet : ", err, "This may be due to an ongoing rollout.")
				return
			}
			if rumor.Text != "" {
				//print the message
				g.PrintEvotingResetStep(rumor.Origin, rumor.Text)

				is_existing := false
				for i := 0 ; i < len(g.pending_nodes_requests) ; i++ {
					if rumor.Text == g.pending_nodes_requests[i] {
						is_existing = true
						break
					}
				}

				if is_existing == true {
					_, ok := g.reset_requests[rumor.Text]
					if ok == true {
						found := false
						for j := 0 ; j < len(g.reset_requests[rumor.Text]) ; j++ {
							if g.reset_requests[rumor.Text][j] == rumor.Origin {
								found = true
								break
							}
						}
						if found == false {
							g.reset_requests[rumor.Text] = append(g.reset_requests[rumor.Text], rumor.Origin)
						}
					} else { // ok == false
						g.reset_requests[rumor.Text] = []string{rumor.Origin}
					}

					var maxToCollect int
					if strings.Contains(rumor.Text, "EXPEL") {
						maxToCollect = len(g.Cluster.Members) - 1
					} else { // strings.Contains(rumor.Text, "EXPEL") == false
						maxToCollect = len(g.Cluster.Members)
					}

					if len(g.reset_requests[rumor.Text]) == maxToCollect {

						delete(g.reset_requests, rumor.Text)

						numbers := (len(g.Cluster.Members) + 1 ) / 2
						if numbers % 2 == 0 {
							numbers ++
						}
						idx := rand.Perm(len(g.Cluster.Members))[:numbers]
						auth := make([]string, numbers)
						for i, e := range idx{
							auth[i] = g.Cluster.Members[e]
						}
						g.Cluster.Authorities = auth

						cluster := g.Cluster
						data, err := protobuf.Encode(cluster)
						if err != nil {
							log.Error("Could not encode cluster :", err)
							return
						}
						cipher := ies.Encrypt(g.Cluster.MasterKey, data)

						for _, member := range cluster.Members {
							//send them the new master key using the previous master key
							if member == g.Name {
								continue
							}
							addr := g.FindPath(member)

							bc := BroadcastMessage{
								ClusterID:   *cluster.ClusterID,
								Destination: member,
								HopLimit:    10,
								Reset:       true,
								CaseRequest: rumor.Text,
								Data:        cipher,
							}
							gp := GossipPacket{Broadcast: &bc}
							go g.SendTo(addr, gp)
						}
					}
				}
			}
			//in any case add it to the map..
			if g.Name != rumor.Origin {
				g.Cluster.HeartBeats[rumor.Origin] = true
			}

		} else if message.AckResend {
			log.Lvl1(g.Name, " received e-voting ack for request resending")
			decrypted := ies.Decrypt(g.Cluster.MasterKey, message.Data)
			var rumor RumorMessage
			err := protobuf.Decode(decrypted, &rumor)
			if err != nil {
				log.Error(g.Name, "Error decoding packet : ", err, "This may be due to an ongoing rollout.")
				return
			}
			if rumor.Text != "" {
				//print the message
				g.PrintEvotingResendStep(rumor.Origin, rumor.Text)

				is_existing := false
				for i := 0 ; i < len(g.pending_nodes_requests) ; i++ {
					if rumor.Text == g.pending_nodes_requests[i] {
						is_existing = true
						break
					}
				}

				if is_existing == true {
					_, ok := g.members_ready_resend_requests[rumor.Text]
					if ok == true {
						found := false
						for j := 0 ; j < len(g.members_ready_resend_requests[rumor.Text]) ; j++ {
							if g.members_ready_resend_requests[rumor.Text][j] == rumor.Origin {
								found = true
								break
							}
						}
						if found == false {
							g.members_ready_resend_requests[rumor.Text] = append(g.members_ready_resend_requests[rumor.Text], rumor.Origin)
						}
					} else { // ok == false
						g.members_ready_resend_requests[rumor.Text] = []string{rumor.Origin}
					}

					if len(g.members_ready_resend_requests[rumor.Text]) == len(g.Cluster.Members) - 1 {
						delete(g.members_ready_resend_requests, rumor.Text)

						if strings.Contains(rumor.Text, "JOIN ") {
							nodeRequest := rumor.Text[5:]
							g.BroadcastJoin(nodeRequest)
						}
					}
				}
			}
			//in any case add it to the map..
			if g.Name != rumor.Origin {
				g.Cluster.HeartBeats[rumor.Origin] = true
			}

		} else {
			decrypted := ies.Decrypt(g.Cluster.MasterKey, message.Data)
			var rumor RumorMessage
			err := protobuf.Decode(decrypted, &rumor)
			if err != nil {
				log.Error(g.Name, "Error decoding packet : ", err, "This may be due to an ongoing rollout.")
				return
			}
			log.Lvl2(g.Name, "got a broadcast..from ", rumor.Origin)

			if rumor.Text != "" && rumor.Origin != g.Name {
				//print the message
				g.PrintBroadcast(rumor)

			}
			//in any case add it to the map..
			g.Cluster.HeartBeats[rumor.Origin] = true
			if !Contains(g.Cluster.Members, rumor.Origin) {
				g.Cluster.Members = append(g.Cluster.Members, rumor.Origin)
			}
		}

	}

}

//RemoveFromList removes s from strings
func RemoveFromList(strings []string, s string) []string {
	for i, s1 := range strings {
		if s1 == s {
			if i == len(strings)-1 {
				return strings[:i]
			}
			return append(strings[:i], strings[i+1:]...)
		}
	}
	return strings
}

//ReceiveJoinRequest handles a join request will triger the e-voting mechanism if activated.
func (g *Gossiper) ReceiveJoinRequest(message RequestMessage) {
	if message.Recipient != g.Name {
		addr := g.FindPath(message.Recipient)
		if addr == "" {
			log.Error("Could not find a path to ", message.Recipient)
			return
		}
		gp := GossipPacket{JoinRequest: &message}
		err := g.SendTo(addr, gp)
		if err != nil {
			log.Error("Error : ", err)
			return
		}
		return
	}
	log.Lvl2("Got request from ", message.Origin)
	_, ok := g.Cluster.HeartBeats[message.Origin]
	if ok {
		//its an update message.
		log.Lvl2(g.Name, " got an update message")
		g.Cluster.PublicKeys[message.Origin] = message.PublicKey
		return

	} else {
		log.Lvl2(g.Name, "got new request from : ", message.Origin)
	}

	if g.ackAll {
		g.ReceiveDecisionJoinRequest(message, 1)
	} else {
		g.pending_nodes_requests = append(g.pending_nodes_requests, "JOIN "+message.Origin)
		g.pending_messages_requests = append(g.pending_messages_requests, message)
		g.BroadcastJoin(message.Origin)
	}
	//Start e-voting protocol to decide if accept...

}

//ReceiveDecisionJoinRequest receives a decision concerning a request.
func (g *Gossiper) ReceiveDecisionJoinRequest(message RequestMessage, decision int) {
	//Once the decision has been taken we have the result..
	log.Lvl2("Decision for ", message.Origin, ", is :", decision)
	var reply RequestReply
	if decision == 1 {
		g.UpdateCluster(message)
		data, err := protobuf.Encode(g.Cluster)
		if err != nil {
			log.Error("Could not encode cluster : ", err)
		}
		//encrypt it ...
		ek := g.Keypair.KeyDerivation(&message.PublicKey)
		enc := ies.Encrypt(ek, data)

		reply = RequestReply{
			Accepted:           true,
			Recipient:          message.Origin,
			ClusterID:          *g.Cluster.ClusterID,
			EphemeralKey:       g.Keypair.PublicKey,
			ClusterInformation: enc,
		}
	} else { // decision == 0
		reply = RequestReply{
			Accepted:           false,
			Recipient:          message.Origin,
			ClusterID:          *g.Cluster.ClusterID,
			EphemeralKey:       nil,
			ClusterInformation: nil,
		}
	}

	//Update the cluster with the new member info.
	gp := GossipPacket{RequestReply: &reply}
	addr := g.FindPath(message.Origin)
	if addr == "" {
		log.Error("Could not find the address")
	}
	err := g.SendTo(addr, gp)
	if err != nil {
		log.Error("Error while sending reply to ", message.Origin, " : ", err)
	}
	if g.ackAll {
		return
	}

}

//ReceiveDecisionExpelRequest receives a decision concerning a request.
func (g *Gossiper) ReceiveDecisionExpelRequest(message RequestMessage, decision int) {
	//Once the decision has been taken we have the result..
	log.Lvl2("Decision for ", message.Recipient,", is :" , decision)
	var reply RequestReply
	if decision == 1 {
		reply = RequestReply{
			Banned:             true,
			Recipient:          message.Recipient,
			ClusterID:          *g.Cluster.ClusterID,
			ExpelRequest:		true,
		}
	} else { // decision == 0
		reply = RequestReply{
			Banned:             false,
			Recipient:          message.Recipient,
			ClusterID:          *g.Cluster.ClusterID,
			ExpelRequest:		true,
		}
	}

	//Update the cluster with the new member info.
	gp := GossipPacket{RequestReply: &reply}
	addr := g.FindPath(message.Recipient)
	if addr == "" {
		log.Error("Could not find the address")
	}
	err := g.SendTo(addr, gp)
	if err != nil {
		log.Error("Error while sending reply to ", message.Recipient, " : ", err)
	}
	if g.ackAll {
		return
	}
}

//ReceiveRequestReply receive a reply for a request if the member is accepted or banned then initiliaze the cluster.
func (g *Gossiper) ReceiveRequestReply(message RequestReply) {

	if message.Recipient != g.Name {
		addr := g.FindPath(message.Recipient)
		if addr == "" {
			log.Error("Could not find path to ", message.Recipient)
			return
		}
		gp := GossipPacket{RequestReply: &message}
		err := g.SendTo(addr, gp)
		if err != nil {
			log.Error("Could not send packet ", err)
		}
		return
	}

	if message.ExpelRequest == true {
		if !message.Banned {
			g.PrintDeniedExpelling(message.ClusterID)
			return
		}

		g.PrintGotExpelled(message.ClusterID)
		g.LeaveChan<-true
		g.Cluster = nil
		
	} else { // message.ExpelRequest == false
		//reply <- g.ReplyChan
		if !message.Accepted {
			g.PrintDeniedJoining(message.ClusterID)
			return
		}

		var cluster clusters.Cluster
		pk := ies.PublicKey(message.EphemeralKey)
		data := ies.Decrypt(g.Keypair.KeyDerivation(&pk), message.ClusterInformation)
		err := protobuf.Decode(data, &cluster)
		if err != nil {
			log.ErrFatal(err, "Could not decode cluster information ")
		}
		g.PrintAcceptJoiningID(cluster)

		g.Cluster = &cluster
		clusters.InitCounter(g.Cluster)
		log.Lvl2(g.Name, "Cluster initialized. ")
		//Start the heartbeatloop immediately
		go g.HeartbeatLoop()
	}
}

//KeyRollout Initiate the key rollout. Assume that the leader has been elected and he calls this method.
func (g *Gossiper) KeyRollout(leader string) {
	//Send a new key pair to the "leader"

	var err error
	g.Keypair, err = ies.GenerateKeyPair()
	g.Cluster.PublicKeys = make(map[string]ies.PublicKey)

	if err != nil {
		log.Error("Could not generate new keypair : ", err)
	}

	//go func() {
	log.Lvl2(g.Name, "sending a rollout update to ", leader)
	g.Cluster.PublicKeys[g.Name] = g.Keypair.PublicKey

	if leader != g.Name {
		//Request to join
		done := false
		for !done {
			select {
			case <-time.After(time.Second * 5):
				go g.RequestJoining(leader)

				break
			case <-ClusterUpdated:
				done = true
				break
			}
		}

	}

	//}()

	//leader does the rest.
	if leader == g.Name {
		g.Cluster.HeartBeats[g.Name] = true
		//Check who is still in the cluster.
		var nextMembers []string
		for _, m := range g.Cluster.Members {
			flag, ok := g.Cluster.HeartBeats[m]
			if !ok || !flag {
				//he wants to be removed
				log.Lvl1("Removing : ", m, " from cluster")
				delete(g.Cluster.PublicKeys, m)
			} else {
				log.Lvl1("Staying in cluster ", m)
				nextMembers = append(nextMembers, m)
			}

		}
		g.Cluster.Members = nextMembers

		log.Lvl1("New members for this key rollout : ", nextMembers)

		//Check if received all the keys from them
		for {
			<-time.After(5 * time.Second)
			if len(g.Cluster.PublicKeys) == len(nextMembers) {
				//we got all the maps we can generate the master key and return
				log.Lvl2("Got all the members needed")
				err := g.AnnounceNewMasterKey()
				if err != nil {
					log.Error("Could not announce master key : ", err)
				}
				return
			}
			log.Lvl3("Missing some members ( have ", len(g.Cluster.PublicKeys), "need ", len(nextMembers), ")")
			log.Lvl2(g.Cluster.PublicKeys)
		}

	}

}

//MasterKeyGen generate a new master key
func (g *Gossiper) MasterKeyGen() ies.PublicKey {
	kp, err := ies.GenerateKeyPair()
	if err != nil {
		log.Error("Could not generate key pair :", err)
	}

	return kp.PublicKey
}

//AnnounceNewMasterKey announce the new master key and new information for this key rollout.
func (g *Gossiper) AnnounceNewMasterKey() error {
	old := g.Cluster.MasterKey
	g.Cluster.MasterKey = g.MasterKeyGen()
	numbers := (len(g.Cluster.Members) + 1) / 2
	if numbers%2 == 0 {
		numbers++
	}
	idx := rand.Perm(len(g.Cluster.Members))[:numbers]
	auth := make([]string, numbers)
	for i, e := range idx {
		auth[i] = g.Cluster.Members[e]
	}
	g.Cluster.Authorities = auth

	cluster := g.Cluster
	data, err := protobuf.Encode(cluster)
	if err != nil {
		log.Error("Could not encode cluster :", err)
		return err
	}
	cipher := ies.Encrypt(old, data)

	for _, member := range cluster.Members {
		//send them the new master key using the previous master key
		if member == g.Name {
			continue
		}
		addr := g.FindPath(member)

		bc := BroadcastMessage{
			ClusterID:   *cluster.ClusterID,
			Destination: member,
			HopLimit:    10,
			Rollout:     true,
			Data:        cipher,
		}
		gp := GossipPacket{Broadcast: &bc}
		go g.SendTo(addr, gp)
	}

	return nil
}

//UpdateCluster updates the cluster.
func (g *Gossiper) UpdateCluster(message RequestMessage) {
	g.Cluster.Members = append(g.Cluster.Members, message.Origin)
	g.Cluster.PublicKeys[message.Origin] = message.PublicKey
	g.Cluster.HeartBeats[message.Origin] = true
}

//UpdateCluster updates the cluster by removing an expelled node.
func (g *Gossiper) UpdateClusterExpel(expelledNode string) {
	for i := 0 ; i < len(g.Cluster.Members) ; i++ {
		if g.Cluster.Members[i] == expelledNode {
			copy(g.Cluster.Members[i:], g.Cluster.Members[i+1:])
			g.Cluster.Members = g.Cluster.Members[:len(g.Cluster.Members) - 1]
			break
		}
	}
	_, ok1 := g.Cluster.PublicKeys[expelledNode]
	if ok1 == true {
		delete(g.Cluster.PublicKeys, expelledNode)
	}
	_, ok2 := g.Cluster.HeartBeats[expelledNode]
	if ok2 == true {
		delete(g.Cluster.HeartBeats, expelledNode)
	}
}

func (g *Gossiper) UpdateFromReset(cluster clusters.Cluster) {
	log.Lvl2("Update information form reset ")

	g.Cluster = &cluster
	clusters.InitCounter(g.Cluster)
	g.Cluster.HeartBeats = make(map[string]bool)
}

//UpdateFromRollout update information from a rollout.
func (g *Gossiper) UpdateFromRollout(cluster clusters.Cluster) {
	log.Lvl2("Update information form a new cluster :O ")

	g.Cluster = &cluster
	clusters.InitCounter(g.Cluster)
	g.Cluster.HeartBeats = make(map[string]bool)

	g.slice_results = make([][]string, 0)
	g.acks_cases = make(map[string][]string)
	g.correct_results_rcv = make(map[string][]string)
	g.reset_requests = make(map[string][]string)
	g.members_ready_resend_requests = make(map[string][]string)
	g.pending_nodes_requests = make([]string, 0)
	g.pending_messages_requests = make([]RequestMessage, 0)
	g.displayed_requests = make([]string, 0)
}

//RequestLeave sends a broadcast with the leave flag set
func (g *Gossiper) RequestLeave() {
	g.SendBroadcast("", true)
}
