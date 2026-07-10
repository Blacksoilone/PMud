package session

import (
	"log"
	"net"
)

func StartSession() {
	listener, err := net.Listen("tcp", ":4000")
	log.Println("Server listening on :4000")
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go func() {
			if err := handleConn(conn); err != nil {
				log.Print(err)
			}
		}()
	}

}
