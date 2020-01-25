package gossiper

//SendAnonymousMessage - handles anonymous message sending
func (g *Gossiper) SendAnonymousMessage(gp GossipPacket, anonLevel float64) error {

	errChan := make(chan error)
	var receiver string
	// sending an anonymous private message
	if gp.Private != nil {
		receiver = gp.Private.Destination
	}

	encryptedBytes := g.EncryptPacket(gp, receiver)
	anonMsg := AnonymousMessage{
		EncryptedContent: encryptedBytes,
		Receiver:         receiver,
		AnonymityLevel:   anonLevel,
	}
	packetToSend := GossipPacket{AnonymousMsg: &anonMsg}

	go g.ReceiveAnonymousMessage(packetToSend, errChan)
	return <-errChan
}
