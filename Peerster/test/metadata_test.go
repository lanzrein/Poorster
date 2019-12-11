package test

import (
	"github.com/JohanLanzrein/Peerster/gossiper"
	"go.dedis.ch/onet/log"
	"go.dedis.ch/protobuf"
	"strings"
	"testing"
)

func TestMetaData(t *testing.T) {
	log.Print("Testing meta data")

	md := gossiper.MetaData{
		Name:     "hello",
		Length:   int64(32),
		Metafile: []byte{5, 6, 6, 3, 4, 8, 9, 10, 11},
		MetaHash: []byte{1, 2, 3, 4, 5},
	}

	data, err := protobuf.Encode(&md)
	if err != nil {
		log.Error(err)
		t.Fail()
	}
	res := new(gossiper.MetaData)
	err = protobuf.Decode(data, res)

	if err != nil {
		log.Error(err)
		t.Fail()
	}

	if res.Name != md.Name || res.Length != res.Length {
		log.Error("non matching name or length")
		t.Fail()
	}

}

func TestProto(t *testing.T) {
	b := uint64(2)
	kw := strings.Split("a,b,c", ",")
	gm := gossiper.Message{
		Budget:   &b,
		Keywords: &kw,
	}

	data, err := protobuf.Encode(&gm)
	if err != nil {
		log.Error("Error on encoding : ", err)
	}

	res := gossiper.Message{}
	if err = protobuf.Decode(data, &res); err != nil {
		log.Error("Error on deconding : ", err)
	}

	log.Lvl2("Result is : ", res.Keywords)
}

//func TestRemoveAny(t *testing.T) {
//	xs := make([]gossiper.MetaChan, 3)
//	Metadata := gossiper.MetaData{
//		Name:     "hi0",
//		Length:   32,
//		Metafile: []byte{1, 3, 4, 5},
//		MetaHash: [32]byte{6, 7, 8, 9, 10},
//	}
//	Metadata1 := gossiper.MetaData{
//		Name:     "hi1",
//		Length:   30,
//		Metafile: []byte{21, 1, 4, 5},
//		MetaHash: [32]byte{6, 37, 8, 9, 10},
//	}
//	Metadata2 := gossiper.MetaData{
//		Name:     "hi2",
//		Length:   30,
//		Metafile: []byte{1, 33, 4, 5},
//		MetaHash: [32]byte{6, 97, 8, 9, 10},
//	}
//	chan1 := make(chan gossiper.DataReply)
//	chan2 := make(chan gossiper.DataReply)
//	chan3 := make(chan gossiper.DataReply)
//	xs[0] = gossiper.MetaChan{Metadata, chan1}
//	xs[1] = gossiper.MetaChan{Metadata1, chan2}
//	xs[2] = gossiper.MetaChan{Metadata2, chan3}
//	log.Print(xs)
//
//	gossiper.RemoveMetafile(xs, Metadata1)
//	log.Print(xs)
//
//}
