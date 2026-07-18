package session

import (
	"PMud/internal/world"
	"log"
	"net"
)

func StartSession(game *world.World) {
	loop := world.NewLoop(game)
	loop.Start()

	listener, err := net.Listen("tcp", ":4000")

	if err != nil {
		log.Fatal(err)
	}
	log.Println("Server listening on :4000")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go func(conn net.Conn) {
			if err := handleConn(conn, loop); err != nil {
				log.Print(err)
			}
		}(conn)
	}

}
