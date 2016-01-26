package stream

import (
	"bufio"
	"io/ioutil"
	"net/http"
	"os"

	"log"

	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()

	if err != nil {
		log.Println("Error:", err)
	}
}

type Options struct {
	WebSocket string
	Http      string
	Stdin     string
}

type Receiver interface {
	Receive([]byte)
}

func Run(opt *Options, r Receiver) error {

	ws := handleWebSocket(opt.WebSocket)
	rest := handleREST(opt.Http)
	stdin := handleStdin(opt.Stdin)

	for {
		select {
		case d := <-stdin:
			r.Receive(d)
		case d := <-ws:
			r.Receive(d)
		case d := <-rest:
			r.Receive(d)
		}
	}
}

func handleStdin(enable string) (c chan []byte) {

	if enable != "yes" {
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
			log.Println("Error:", err)
		}
	}()
	return
}

func handleWebSocket(addr string) (c chan []byte) {
	if "" == addr {
		return
	}
	c = make(chan []byte)
	log.Println("WebSocket listen on", addr)
	return
}
