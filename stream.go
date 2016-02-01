package stream

import (
	"bufio"
	"io/ioutil"
	"net/http"
	"os"

	"golang.org/x/net/websocket"

	"log"
)

type Options struct {
	WebSocket string
	Http      string
	Stdin     string
}

type Receiver func([]byte)

func Run(opt *Options, r Receiver) error {

	ws := handleWebSocket(opt.WebSocket)
	rest := handleREST(opt.Http)
	stdin := handleStdin(opt.Stdin)

	for {
		select {
		case d := <-stdin:
			r(d)
		case d := <-ws:
			r(d)
		case d := <-rest:
			r(d)
		}
	}
}

func handleStdin(enable string) (c chan []byte) {

	if enable != "enable" {
		return
	}

	stdin := bufio.NewReader(os.Stdin)
	c = make(chan []byte)
	go func() {
		for {
			b, _, e := stdin.ReadLine()
			if e != nil {
				log.Println("Error:", e)
				close(c)
				return
			}

			c <- b
		}
	}()

	return c
}

type RestServer struct {
	c chan []byte
}

func (s *RestServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Error:", err)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write([]byte("ok"))

	s.c <- b
}

func handleREST(addr string) (c chan []byte) {
	if "" == addr {
		return
	}
	c = make(chan []byte)
	go func() {
		log.Println("REST listen on", addr)
		if err := http.ListenAndServe(addr, &RestServer{c}); err != nil {
			log.Println("REST Error:", err)
		}
	}()

	return
}

func handleWebSocket(addr string) (c chan []byte) {
	if "" == addr {
		return
	}
	c = make(chan []byte)

	server := &websocket.Server{
		Handler: func(ws *websocket.Conn) {
			defer ws.Close()

			websocket.Message.Send(ws, "")

			for {
				var buf []byte
				if err := websocket.Message.Receive(ws, &buf); err != nil {
					log.Println("ws closed")
					break
				}
				c <- buf
			}
		},
	}

	go func() {
		log.Println("WS listen on", addr)
		if err := http.ListenAndServe(addr, server); err != nil {
			log.Println("WS Error:", err)
		}
	}()

	return
}
