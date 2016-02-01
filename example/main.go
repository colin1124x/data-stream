package main

import (
	"fmt"

	"github.com/colin1124x/data-stream"
)

func main() {

	stream.Run(&stream.Options{
		Http:      ":8000",
		WebSocket: ":8080",
	}, func(b []byte) {
		fmt.Println("Receive:", string(b))
	})
}
