//Client file that handles message from the client to the node.
package main

import (
	"encoding/hex"
	"errors"
	"github.com/JohanLanzrein/Peerster/gossiper"
	"go.dedis.ch/onet/log"
	"go.dedis.ch/protobuf"
	"net"
	"strings"
)

//Client has one address
type Client struct {
	address net.UDPAddr
}

//NewClient resolves the address for the client and returns the structure created
func NewClient(str string) Client {

	log.Lvl3("Starting new client at address : ", str)
	addr, err := net.ResolveUDPAddr("udp4", str)
	if err != nil {
		return Client{}
	}
	c := Client{
		address: *addr,
	}

	return c
}

//SendsMsg sends a txt message to the address of the client c
//If an error arises it is returned by the function
func (c *Client) SendMsg(txt string) error {

	//encode message
	log.Lvl3("Sending message : ", txt)
	msg := gossiper.Message{
		Text: txt,
	}

	packetBytes, err := protobuf.Encode(&msg)
	if err != nil {
		return err

	}
	err = c.SendBytes(packetBytes)
	if err != nil {
		return err
	}

	return nil
}

//SendPrivateMsg sends a private message msg to the destination dest
func (c *Client) SendPrivateMsg(msg string, dest string, anonymous bool, anonLevel float64) {
	toSend := gossiper.Message{
		Text: msg,
		Destination: &dest,
		Anonymous: anonymous,
	}

	if anonymous {
		toSend.AnonimityLevel = anonLevel
	}

	packetBytes, err := protobuf.Encode(&toSend)
	if err != nil {
		log.Error("Error could not encode message : ", err)
	}
	err = c.SendBytes(packetBytes)
	if err != nil {
		log.Error("Error could not send message : ", err)
	}

	return

}

//SendBytes sends the bytes to the connection of the client
func (c *Client) SendBytes(packetBytes []byte) error {
	conn, err := net.DialUDP("udp", nil, &c.address)
	if err != nil {
		return err
	}
	n, err := conn.Write(packetBytes)
	if err != nil {
		return errors.New("Couldn't write to UDP socket ")
	}
	log.Lvl3("Wrote ", n, " bytes to the connection")
	return nil
}

//SendFileToIndex sends the file with name s to be indexed by the gossiper
func (c *Client) SendFileToIndex(s *string) error {
	tosend := &gossiper.Message{File: s}
	msg, err := protobuf.Encode(tosend)
	if err != nil {
		log.Error("Error on encoding : ", err)
	}
	err = c.SendBytes(msg)
	if err != nil {
		log.Error("Could not write to connection : ", err)

	}
	return err

}

//RequestFile request the file with name file from destination. the request is the MetaHash of the file.
func (c *Client) RequestFile(file *string, destination *string, request *string, anonymous bool, anonLevel float64) error {
	bytes, err := hex.DecodeString(*request)
	if err != nil {
		return errors.New("Unable to decode hex string")
	}

	tosend := &gossiper.Message{
		Text:        "",
		Destination: destination,
		File:        file,
		Request:     &bytes,
		Anonymous:	 anonymous,
	}

	if anonymous {
		tosend.AnonimityLevel = anonLevel
	}

	log.Lvl3("Sending to send : ", *tosend)
	msg, err := protobuf.Encode(tosend)
	if err != nil {
		return err
	}
	err = c.SendBytes(msg)
	return err
}

func (c *Client) SearchFile(keywords *string, budget *int) {
	res := strings.Split(*keywords, ",")
	log.Lvl3(res)

	m := &gossiper.Message{

		Keywords: &res,
	}
	log.Lvl2("Budget : ", budget)
	if *budget < 0 {
		m.Budget = nil
	} else {
		m.Budget = new(uint64)
		*m.Budget = uint64(*budget)
	}
	data, err := protobuf.Encode(m)
	if err != nil {
		log.Error("Could not encode message : ", err)
		return
	}

	log.Lvl3("res : ", data)
	log.Lvl3("budget : ", m.Budget)
	log.Lvl3("kw : ", m.Keywords)

	if err = c.SendBytes(data); err != nil {
		log.Error("Could not send bytes : ", err)
		return
	}

}
