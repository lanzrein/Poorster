//file with function to handle server requests.

package gossiper

import (
	"go.dedis.ch/onet/log"
	"net"
	"net/http"
)

//LoadGossiper Try to load a new server. In cases of failure it will return without saying anything
//This is to prevent the cases where there is already a GUI server running on the same machine.
func LoadServer(g *Gossiper) {

	log.Lvl2("Loading server at address : 127.0.0.1:", (g.GUIPort))
	http.HandleFunc("/message", g.GetMessages)
	http.HandleFunc("/node", g.AddNode)
	http.HandleFunc("/id", g.GetId)
	//Hw2 handler
	http.HandleFunc("/privatemsg", g.PrivateMessageHandle)
	http.HandleFunc("/sharefile", g.FileSharingHandle)
	http.HandleFunc("/downloadfile", g.FileDownloadingHandle)
	//Hw3 handlers
	http.HandleFunc("/searchfile", g.FileSearchHandle)
	http.HandleFunc("/downloadfoundfile", g.FoundFileHandle)
	//project handlers
	http.HandleFunc("/initcluster", g.InitClusterHandle)

	http.Handle("/", http.FileServer(http.Dir("./frontend")))
	for {

		err := http.ListenAndServe("127.0.0.1:"+(g.GUIPort), nil)

		if err != nil {
			if e, ok := err.(*net.OpError); ok {
				if e.Op == "listen" {
					log.Lvl2("Error could not start server : ", e)
					log.Lvl2("GUI port (80) is already used. This instance of Peerster will not use the GUI.")
					break

				}
			} else {
				log.Lvl2("Error on server : ", err)
			}

		}
	}

}
