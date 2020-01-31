//@authors Hrusanov Aleksandar, Lanzrein Johan, Rinaldi Vincent

//filesharing methods for the filesharing protocol
package gossiper

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"go.dedis.ch/onet/log"
	"io"
	"math"
	"os"
	"time"
)

const pathShared = "./_SharedFiles/"
const pathDownloads = "./_Downloads/"
const block = 8192 //2 ^ 13

//FileSharingReceiveProtocol Start a file sharing protocol on the receiver end with a destination
func (g *Gossiper) FileSharingReceiveProtocol(destination string, MetaHash []byte, datareplies chan DataReply, filename string, reference *FoundFiles) {
	log.Lvl3("Starting file download from destination : ", destination, "file with meta hash :", MetaHash)

	g.PrintDownloadStart(filename, destination)
	//first request the meta data
	datareq := DataRequest{
		Origin:      g.Name,
		Destination: destination,
		HopLimit:    g.HopLimit,
		HashValue:   MetaHash,
	}

	pckt := GossipPacket{DataRequest: &datareq}
	addr := g.FindPath(destination)
	done := false
	reply := DataReply{}
	g.UpdateRequestList(datareplies, MetaHash)

	_ = g.SendTo(addr, pckt)
	for {
		select {
		case reply = <-datareplies:
			done = true
			break
		case <-time.After(5 * time.Second):
			log.Lvl3("Timeout on data request")
			_ = g.SendTo(addr, pckt)
			break
		}
		if done {
			break
		}
	}

	//we should get the Metadata
	Metafile := make([]byte, len(reply.Data))
	copy(Metafile, reply.Data)

	log.Lvl3("Got metafile  : ", Metafile)
	metadata := MetaData{
		Name:     filename,
		Length:   0,
		Metafile: []byte{},
		MetaHash: MetaHash,
	}

	g.AddMetaData(&metadata)

	var writer bytes.Buffer
	ptr := 0
	total_written := int64(0)
	ith := 1
	for {
		log.Lvl3("Requesting block", ith)
		if reference != nil {
			destination = reference.OriginChunks[uint64(ith)]
		}
		g.PrintDownloadChunk(filename, destination, int64(ith))

		dr := g.RequestIthBlock(ith, Metafile, destination)

		//update what the channel expects.
		g.UpdateRequestList(datareplies, dr.HashValue)
		pckt := GossipPacket{DataRequest: &dr}
		_ = g.SendTo(addr, pckt)

		done = false
		for {
			select {
			case reply = <-datareplies:
				done = true
				log.Lvl3("Got a data reply packet on file ", filename)
				break
			case <-time.After(5 * time.Second):
				log.Lvl3("Timeout on data request")
				_ = g.SendTo(addr, pckt)
				break
			}
			if done {
				break
			}
		}
		log.Lvl3("Data length : ", len(reply.Data))

		sha := sha256.Sum256(reply.Data)
		if !CompareShas(reply.HashValue, sha) {
			log.Error("hash value of reply : non matching shas", hex.EncodeToString(reply.HashValue), " expected : ", hex.EncodeToString(sha[:]))
			g.RemoveDataReplyChannel(datareplies, destination)
			return
		}
		if !CompareShas(Metafile[ptr:ptr+32], sha) {
			log.Error("metafile sha and sha : non matching shas")
			g.RemoveDataReplyChannel(datareplies, destination)
			return
		}

		g.Chunks[hex.EncodeToString(sha[:])] = reply.Data

		cnt, err := writer.Write(reply.Data)
		total_written += int64(cnt)
		//update meta data..
		metadata.Length += int64(cnt)
		metadata.Metafile = append(metadata.Metafile, Metafile[ptr:ptr+32]...)

		if err != nil {
			log.Error(err)
			g.RemoveDataReplyChannel(datareplies, destination)
			return
		}
		maxBlocks := len(Metafile) / 32 * 8192
		if cnt != block || metadata.Length == int64(maxBlocks) {
			break
		}
		ith++
		ptr += 32
	}
	log.Lvl3("Got all block reconstructing of file")
	data := writer.Bytes()
	err := g.BuildFromChunks(filename, data, int64(len(data)))
	if err != nil {
		log.Error("Could not write file ", err)
		g.RemoveDataReplyChannel(datareplies, destination)
		return
	}
	g.PrintReconstructFile(filename)
	g.RemoveDataReplyChannel(datareplies, destination)

	if err != nil {
		log.Error("Could not index newly created file ", err)
	}
	return

}

//Scan a file with file name. will produce the chunks of hashes.
func (g *Gossiper) Scan(file *os.File, length int64) ([]byte, error) {
	//file , err := os.Open(path+filename)
	log.Lvl3("Scanning file : ", file.Name(), " of length : ", length)
	shaCount := int(math.Ceil(float64(length) / block))
	hashes := make([]byte, 32*shaCount)
	ptr := 0

	for {
		data := make([]byte, block)
		cnt, err := file.Read(data)
		if err != nil && err != io.EOF {
			log.Error("Error while reading file : ", err)
			return []byte{}, err
		}
		if cnt == 0 {
			break
		}

		data = data[:cnt]
		sha := sha256.Sum256(data)
		//Dirty hack bc normal copy does not works :@
		CstmCopy(hashes[ptr:ptr+len(sha)], sha)
		//adding the chunk
		log.Lvl2("Adding chunk : ", hex.EncodeToString(sha[:]))
		g.Chunks[hex.EncodeToString(sha[:])] = data
		ptr += len(sha)
		if cnt != block || (err != nil && err == io.EOF) {
			//we are at last iteration
			break
		}
	}

	return hashes, nil
}

//Index a file with filename. Will produce a meta data file for the gossiper.
func (g *Gossiper) Index(filename string, entry string) error {
	path := entry + filename
	log.Lvl3("Opening file : ", path)
	file, err := os.Open(path)
	if err != nil {

		log.Error("Could not open a file : ", err)
		return err
	}
	defer file.Close()
	fi, err := file.Stat()
	if err != nil {
		log.Error("Could not get file information ", err)
		return err
	}
	length := fi.Size()
	mf, err := g.Scan(file, length)
	if err != nil {
		log.Error("Error while scanning the file : ", err)
		return err
	}
	hash := make([]byte, 32)
	CstmCopy(hash, sha256.Sum256(mf))
	meta := MetaData{
		Name:     filename,
		Length:   length,
		Metafile: mf,
		MetaHash: hash,
	}

	g.AddMetaData(&meta)

	//now we ask to have it on the blockchain..
	if g.ConfirmationRunning() {
		go g.AddToStack(meta.Name, meta.MetaHash, meta.Length)

	}

	return nil
}

func (g *Gossiper) ConfirmationRunning() bool {
	return g.ackAll || g.hw3ex3
}

//AddMetaData adds the meta data to our data structure.
func (g *Gossiper) AddMetaData(data *MetaData) {

	g.metalock.mu.Lock()
	defer g.metalock.mu.Unlock()
	sha := hex.EncodeToString(data.MetaHash)

	log.Lvl2("Adding file : ", data.Name, " with metahash : ", sha)

	for _, md := range g.metalock.data {
		if bytes.Compare(md.MetaHash, data.MetaHash) == 0 {
			return
		}
	}

	g.metalock.data = append(g.metalock.data, data)
	return
}

//CstmCopy a custom copy to copy from a fixed size byte array.
func CstmCopy(bytes []byte, bytes2 [32]byte) {
	if len(bytes) != 32 {
		return
	}
	for i, b := range bytes2 {
		bytes[i] = b
	}

}

//BuildFromChunks builds a filename from chunks given.
//Assumes that the value of the sha256 has already been checked and will do not integrity checks.
func (g *Gossiper) BuildFromChunks(filename string, data []byte, expected int64) error {
	file, err := os.Create(pathDownloads + filename)
	if err != nil {
		log.Error("Error while creating file ", err)
		return err
	}
	defer file.Close()
	n, err := file.Write(data)
	if int64(n) != expected {
		err = errors.New("Error could not write all data to file")
		return err
	}

	err = file.Sync()
	return err
}

//CompareShas compare two sha256 true iff they match
func CompareShas(bytes []byte, bytes2 [32]byte) bool {
	if len(bytes) != len(bytes2) {
		return false
	}
	for i, b := range bytes {
		if b != bytes2[i] {
			return false
		}
	}

	return true
}

//SendSingleChunk sends the chunk of data corresponding to the hash to the destination.
func (g *Gossiper) SendSingleChunk(metadata MetaData, hash []byte, destination string) {
	//ithChunk := g.GetChunkId(hash, metadata)
	//if ithChunk < 0 {
	//	log.Error(g.Name, " : Could not find chunk for hash : ", hash)
	//	return
	//}
	//
	//data := g.GetDataChunk(ithChunk, metadata.Name)
	log.Lvl2("Hash i s: ", hex.EncodeToString(hash))
	data := g.Chunks[hex.EncodeToString(hash)]
	sha := (sha256.Sum256(data))
	log.Lvl2("Sending chunk of : ", len(data), "sha : ", hex.EncodeToString(sha[:]))
	reply := DataReply{
		Origin:      g.Name,
		Destination: destination,
		HopLimit:    g.HopLimit,
		HashValue:   hash,
		Data:        data,
	}
	pckt := GossipPacket{DataReply: &reply}
	addr := g.FindPath(destination)
	err := g.SendTo(addr, pckt)
	if err != nil {
		log.Error(err)

	}

}

//GetDataChunk gets the ithChunk of data from filename
func (g *Gossiper) GetDataChunk(ithChunk int64, filename string) []byte {
	g.metalock.mu.Lock()
	defer g.metalock.mu.Unlock()
	path := pathShared + filename
	file, err := os.Open(path)
	if err != nil {
		//log.Error(err)
		//return []byte{}
		path = pathDownloads + filename
		file, err = os.Open(path)
		if err != nil {
			log.Error(err)
			return []byte{}
		}

	}

	data := make([]byte, block)
	cnt, err := file.ReadAt(data, ithChunk*block)
	if err != nil && err != io.EOF {
		log.Error(err)
		return []byte{}
	}
	data = data[:cnt]
	return data
}

//Get the MetaData for a file that has a MetaHash...
func (g *Gossiper) GetMetadata(MetaHash []byte) (MetaData, error) {
	g.metalock.mu.Lock()
	defer g.metalock.mu.Unlock()
	for _, md := range g.metalock.data {
		if bytes.Compare(MetaHash, md.MetaHash) == 0 {

			return *md, nil
		}
	}

	return MetaData{}, errors.New("ERROR : gossiper does not have the file indexed")
}

//RequestIthBlock request the ith block creates a datarequest for it.
func (g *Gossiper) RequestIthBlock(i int, data []byte, destination string) DataRequest {
	hash := data[(i-1)*32 : (i)*32]
	log.Lvl2("Hash :", hex.EncodeToString(hash))
	dr := DataRequest{
		Origin:      g.Name,
		Destination: destination,
		HopLimit:    g.HopLimit,
		HashValue:   hash,
	}
	return dr

}

//RemoveDataReplyChannel when the file sending is done close the channel and remove it from the table.
func (g *Gossiper) RemoveDataReplyChannel(ch chan DataReply, destination string) {
	g.CurrentlyDownloading.mu.Lock()
	defer g.CurrentlyDownloading.mu.Unlock()
	close(ch)
	delete(g.CurrentlyDownloading.values, ch)

}

//UpdateRequestList updates the hash expected on the channel ch
func (g *Gossiper) UpdateRequestList(ch chan DataReply, hash []byte) {
	g.CurrentlyDownloading.mu.Lock()
	defer g.CurrentlyDownloading.mu.Unlock()
	g.CurrentlyDownloading.values[ch] = hash
}

//GetChunk returns the correct chunk for the destination and hash
func (g *Gossiper) GetChunk(destination string, hash []byte) {
	//check if its a meta hash
	addr := g.FindPath(destination)
	if addr == "" {
		//unknwon host
		return
	}
	MetaData, err := g.GetMetadata(hash)
	log.Lvl2("Sending some file data.")
	if err != nil {
		log.Lvl2("Sending one chunk")
		//it was most likely a chunk check what chunk it is
		//we assume that the gossiper already sends us the correct destination etc.
		go g.SendSingleChunk(MetaData, hash, destination)
	} else {
		log.Lvl2("Sending some metadata.")

		//send the metadata.
		reply := DataReply{
			Origin:      g.Name,
			Destination: destination,
			HopLimit:    g.HopLimit,
			HashValue:   hash,
			Data:        MetaData.Metafile,
		}

		log.Lvl3("Forwarding a message to destination : ", destination, " at addr", addr)
		err = g.SendTo(addr, GossipPacket{DataReply: &reply})
	}

}
