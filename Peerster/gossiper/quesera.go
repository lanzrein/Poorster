package gossiper

//Didnt have time to finish or even start :(
import (
	"crypto/sha256"
	"encoding/binary"
	//"math/rand"
)

func (b *BlockPublish) Hash() (out [32]byte) {
	h := sha256.New()
	h.Write(b.PrevHash[:])
	th := b.Transaction.Hash()
	h.Write(th[:])
	copy(out[:], h.Sum(nil))
	return
}

func (t *TxPublish) Hash() (out [32]byte) {
	h := sha256.New()
	binary.Write(h, binary.LittleEndian, uint32(len(t.Name)))
	h.Write([]byte(t.Name))
	h.Write(t.MetafileHash)
	copy(out[:], h.Sum(nil))

	return
}

func QueSeraRound() {
	//pick a random value
	//fitness := rand.Float32()
	//my_block := TxPublish{}
	//tlc := TLCMessage{
	//	Origin:      "",
	//	ID:          0,
	//	Confirmed:   0,
	//	TxBlock:     BlockPublish{},
	//	VectorClock: nil,
	//	Fitness:     0,
	//}
	////first round s

}

func CheckValidity(publish BlockPublish) {
	//check if there is no other commited block wiht same name

	//
}
