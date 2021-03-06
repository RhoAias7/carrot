package carrot

import (
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	_ "net/http/pprof"
)

const (
	serverSecret = "37FUqWlvJhRgwPMM1mlHOGyPNwkVna3b"
	port         = 8080
)

//the server maintains the list of clients and
//broadcasts messages to the clients
type Server struct {

	//register requests from the clients
	register chan *Client

	//unregister requests from the clients
	unregister chan *Client

	//access list of existing sessions
	sessions SessionStore

	//keep track of middleware
	Middleware *MiddlewarePipeline

	clients *Clients

	logger *log.Entry
}

func NewServer(sessionStore SessionStore) (*Server, error) {
	clients, err := NewClientList()
	if err != nil {
		return nil, err
	}

	return &Server{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		sessions:   sessionStore,
		Middleware: NewMiddlewarePipeline(),
		clients:    clients,
		logger:     log.WithField("module", "server"),
	}, nil
}

func (svr *Server) Run() {
	go svr.Middleware.Run()
	for {
		select {
		case client := <-svr.register:
			client.softOpen()
			token := SessionToken("")
			//create persistent token for new or invalid sessions
			exists := svr.sessions.Exists(token)
			if (token == "") || !exists {
				var err error
				token, sessionPtr, err := svr.sessions.NewSession()
				if err != nil {
					//handle later
					log.Error(err)
				}

				client.session = sessionPtr

				if svr.sessions.Length() == 1 { //this is the first connected and therefore primary device
					client.session.primaryDevice = true
				}

				svr.clients.Insert(client)

				//TODO: handle way to assign a new primary device if original device disconnects

				uuid, err := svr.sessions.GetPrimaryDeviceToken()
				if err != nil {
					//handle later
					log.Error(err)
				}
				info, err := createInitialDeviceInfo(string(uuid), string(token))
				if err != nil {
					//handle later
					log.Error(err)
				}

				client.sendBeaconInfo <- info
			}

			close(client.start)
		case client := <-svr.unregister:
			if client.Open() {
				svr.logger.WithField("session_token", client.session.Token).Info("client unregistered")
				client.softClose()
				// delete client?
				close(client.send)
				close(client.sendBeaconInfo)
				client = nil
			}
		}
	}
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	//log.Println(r.URL)

	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	http.ServeFile(w, r, "home.html")

}

func (svr *Server) Serve() {
	addr := flag.String("addr", fmt.Sprintf(":%d", port), "http service address")

	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(svr, w, r)
	})

	log.WithFields(log.Fields{
		"port": port,
		"url":  "ws://localhost/",
	}).Infof("Listening...")

	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		fmt.Println(err)
	}
}
