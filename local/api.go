package local
import (
	"os"
	"strings"
	"net/http"
	
	"centi/util"
	"centi/config"
	"centi/network"
	"centi/protocol"
)



func RunCentiApiServer( sc *config.ServerConfiguration,
			logger *util.Logger,
			conn *protocol.ConnManagement,
			queue *network.Queue ) error {

	// general user-related functions
	for uri, page := range sc.Pages {
		http.HandleFunc( uri, func(w http.ResponseWriter, r *http.Request) {
			sendFile( page, sc.NotFoundPage, w )
		})
	}

	// centi network api
	// get list of known public keys
	http.HandleFunc("GET /api/public-keys", func(w http.ResponseWriter, r *http.Request) {
		sendKeys( w, r, conn )
	})

	http.HandleFunc("POST /api/request-public-keys", func(w http.ResponseWriter, r *http.Request) {
		requestKeys( w, r, logger, conn, queue )
	})

	// connect to specified peer
	http.HandleFunc("POST /api/connect", func(w http.ResponseWriter, r *http.Request) {
		connectToPeer(w, r, logger, conn, queue )
	})

	// get the list of peers connected
	http.HandleFunc("GET /api/peers", func(w http.ResponseWriter, r *http.Request) {
		sendPeers(w, r, conn)
	})

	// send encrypted data
	http.HandleFunc("POST /api/request", func(w http.ResponseWriter, r *http.Request) {
		handleRequest( w, r, logger, conn, queue )
	})

	// receive encrypted data
	http.HandleFunc("GET /api/messages", func(w http.ResponseWriter, r *http.Request) {
		handleGetResponse( w, r, logger, conn, queue )
	})

	// i feel like this is fucking useless...
	http.HandleFunc("POST /api/response", func(w http.ResponseWriter, r *http.Request) {
		handleResponse( w, r, logger, conn, queue )
	})

	util.DebugPrintln( util.CyanColor + "Listening and serving at address "+ sc.Address + util.ResetColor )
	return http.ListenAndServe( sc.Address, nil )
}

func sendFile( filename, notFoundPage string, w http.ResponseWriter ) {
	htmlPage, err := os.ReadFile( filename )
	if err != nil {
		htmlPage, err = os.ReadFile( notFoundPage )
		if err != nil {
			w.WriteHeader( 404 )
			w.Write( []byte("Not found") )
		} else {
			w.WriteHeader( 404 )
			w.Write( htmlPage )
			return
		}
	}
	if strings.HasSuffix( filename, ".css" ) {
		w.Header().Set("Content-Type", "text/css")
	}
	w.Write( htmlPage )
}
