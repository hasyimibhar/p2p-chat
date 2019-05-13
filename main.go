package main

import (
	"bufio"
	"encoding/base64"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
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

	log.Printf("[info] initialized node with public key %s",
		base64.StdEncoding.EncodeToString(node.PublicKey()))

	go node.ListenForConnections()

	if *peer != "" {
		if err := node.JoinPeer(*peer); err != nil {
			log.Printf("[error] failed to join peer %s: %s", *peer, err)
			os.Exit(1)
		}
	}

	go func() {
		for chat := range node.ChatMessages() {
			log.Printf("[%s] %s", base64.StdEncoding.EncodeToString(chat.PublicKey), chat.Text)
		}
	}()

	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			msg, _ := reader.ReadString('\n')

			if strings.HasPrefix(msg, "start_privatechat") {
				tokens := strings.Split(msg, " ")
				if len(tokens) != 2 {
					log.Println("[error] usage: start_privatechat <public key>")
					continue
				}

				pubkeyStr := tokens[1]

				pubkey, err := base64.StdEncoding.DecodeString(pubkeyStr)
				if err != nil {
					log.Println("[error] start_privatechat:", err)
				}

				if err := node.StartPrivateChat(pubkey); err != nil {
					log.Println("[error] failed to start private chat:", err)
				}
			} else if strings.HasPrefix(msg, "privatechat") {
				tokens := strings.Split(msg, " ")
				if len(tokens) < 3 {
					log.Println("[error] usage: privatechat <public key> <text>")
					continue
				}

				pubkeyStr := tokens[1]

				pubkey, err := base64.StdEncoding.DecodeString(pubkeyStr)
				if err != nil {
					log.Println("[error] start_privatechat:", err)
				}

				text := strings.Join(tokens[2:], " ")

				if err := node.PrivateChat(pubkey, text); err != nil {
					log.Println("[error] failed to send private chat:", err)
				}
			} else {
				if err := node.Chat(msg); err != nil {
					log.Printf("[error] failed to send chat message: %s", err)
				}
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
