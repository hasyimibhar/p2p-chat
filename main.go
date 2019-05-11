package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	var port = flag.Int("port", 8888, "Port to listen for peers")
	var peers = flag.String("peers", "", "Peers to connect to")
	flag.Parse()

	node, err := NewNode(*port)
	if err != nil {
		log.Println("[error] failed to start node:", err)
	}

	go node.ListenForConnections()

	if *peers != "" {
		addresses := strings.Split(*peers, ",")

		for _, addr := range addresses {
			if err := node.ConnectToPeer(addr); err != nil {
				log.Printf("[error] failed to connect to peer %s: %s", addr, err)
			}
		}
	}

	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			msg, _ := reader.ReadString('\n')
			if err := node.SendChatMessage(msg); err != nil {
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
