//@authors Hrusanov Aleksandar, Lanzrein Johan, Rinaldi Vincent
//Client file that handles message from the client to the node.
package main

import (
	"encoding/hex"
	"errors"
	"net"
	"strings"

	"github.com/JohanLanzrein/Peerster/gossiper"
	"go.dedis.ch/onet/log"
	"go.dedis.ch/protobuf"
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
func (c *Client) SendPrivateMsg(msg string, dest string, anonymous *bool, relayRate *float64, fullAnonimity *bool) {
	toSend := gossiper.Message{
		Text:        msg,
		Destination: &dest,
		Anonymous:   anonymous,
	}

	if *anonymous {
		toSend.RelayRate = relayRate
		toSend.FullAnonimity = fullAnonimity
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
func (c *Client) RequestFile(file *string, destination *string, request *string, anonymous bool, relayRate float64) error {
	bytes, err := hex.DecodeString(*request)
	if err != nil {
		return errors.New("Unable to decode hex string")
	}

	tosend := &gossiper.Message{
		Text:        "",
		Destination: destination,
		File:        file,
		Request:     &bytes,
		Anonymous:   &anonymous,
	}

	if anonymous {
		tosend.RelayRate = &relayRate
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

//Stuff for the project..

func (c *Client) SendBroadcast(content string) {
	msg := gossiper.Message{
		Text:      content,
		Broadcast: new(bool),
	}
	*msg.Broadcast = true
	data, err := protobuf.Encode(&msg)
	if err != nil {
		log.Error("Could not encode packet : ", err)
		return
	}

	err = c.SendBytes(data)
	if err != nil {
		log.Error("Could not send data : ", err)
	}

}

func (c *Client) InitCluster() {
	msg := gossiper.Message{InitCluster: new(bool)}
	*msg.InitCluster = true
	data, err := protobuf.Encode(&msg)
	if err != nil {
		log.Error("Could not encode packet : ", err)
		return
	}

	err = c.SendBytes(data)
	if err != nil {
		log.Error("Could not send data : ", err)
	}
}

func (c *Client) JoinCluster(other *string) {
	msg := gossiper.Message{}
	msg.JoinOther = other

	data, err := protobuf.Encode(&msg)
	if err != nil {
		log.Error("Could not encode packet : ", err)
		return
	}

	err = c.SendBytes(data)
	if err != nil {
		log.Error("Could not send data : ", err)
	}
}

func (c *Client) ExpelFromCluster(other *string) {
	msg := gossiper.Message{}
	msg.ExpelOther = other

	data, err := protobuf.Encode(&msg)
	if err != nil {
		log.Error("Could not encode packet : ", err)
		return
	}

	err = c.SendBytes(data)
	if err != nil {
		log.Error("Could not send data : ", err)
	}
}

func (c *Client) LeaveCluster() {
	msg := gossiper.Message{LeaveCluster: new(bool)}
	*msg.LeaveCluster = true
	data, err := protobuf.Encode(&msg)
	if err != nil {
		log.Error("Could not encode packet : ", err)
		return
	}

	err = c.SendBytes(data)
	if err != nil {
		log.Error("Could not send data : ", err)
	}
}

func (c *Client) ProposeAccept(node *string) {
	msg := gossiper.Message{}
	msg.PropAccept = node

	data, err := protobuf.Encode(&msg)
	if err != nil {
		log.Error("Could not encode packet : ", err)
		return
	}

	err = c.SendBytes(data)
	if err != nil {
		log.Error("Could not send data : ", err)
	}
}

func (c *Client) ProposeDeny(node *string) {
	msg := gossiper.Message{}
	msg.PropDeny = node

	data, err := protobuf.Encode(&msg)
	if err != nil {
		log.Error("Could not encode packet : ", err)
		return
	}

	err = c.SendBytes(data)
	if err != nil {
		log.Error("Could not send data : ", err)
	}
}

//		CALLS			//
//================
func (c *Client) Call(callee *string) {
	msg := gossiper.Message{CallRequest: new(bool)}
	msg.Destination = callee
	*msg.CallRequest = true
	c.encodeAndSendBytes(msg)
}

func (c *Client) HangUp() {
	msg := gossiper.Message{HangUp: new(bool)}
	*msg.HangUp = true
	c.encodeAndSendBytes(msg)
}

func (c *Client) StartRecording() {
	msg := gossiper.Message{StartRecording: new(bool)}
	*msg.StartRecording = true
	c.encodeAndSendBytes(msg)
}

func (c *Client) StopRecording() {
	msg := gossiper.Message{StopRecording: new(bool)}
	*msg.StopRecording = true
	c.encodeAndSendBytes(msg)
}

//		HELPER		//
//================
func (c *Client) encodeAndSendBytes(msg gossiper.Message) {
	data, err := protobuf.Encode(&msg)
	if err != nil {
		log.Error("Could not encode packet : ", err)
		return
	}

	err = c.SendBytes(data)
	if err != nil {
		log.Error("Could not send data : ", err)
	}
}
