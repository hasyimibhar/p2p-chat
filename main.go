package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var port = flag.Int("port", 8888, "Port to listen for peers")
	var peer = flag.String("peer", "", "Peer to connect to")
	flag.Parse()

	node, err := NewNode(*port)
	if err != nil {
		log.Println("[error] failed to start node:", err)
	}

	go node.ListenForConnections()

	if *peer != "" {
		if err := node.JoinPeer(*peer); err != nil {
			log.Printf("[error] failed to join peer %s: %s", *peer, err)
		}
	}

	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			msg, _ := reader.ReadString('\n')
			if err := node.Chat(msg); err != nil {
				log.Printf("[error] failed to send chat message: %s", err)
			}
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigs
	log.Println("[info] received signal:", sig)

	node.Close()

	os.Exit(0)
}
