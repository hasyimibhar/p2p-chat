# P2P Chat

This is a naive implementation of a secure p2p chat app for educational purposes. It uses a simplified version of the Chord protocol, with the following modifications:

- each node has no ID, so when a node joins the network via peer P, it immediately sets P as its successor
- no finger table, so broadcasting is done by routing around the entire overlay network (ring)
- it's not used as DHT

## Usage

```sh
# Run a network with 5 peers
$ go run . -port=8000
$ go run . -port=8001 -peer=localhost:8000
$ go run . -port=8002 -peer=localhost:8000
$ go run . -port=8003 -peer=localhost:8000
$ go run . -port=8004 -peer=localhost:8000
```

To send a public chat message, just type anything and press enter.

To send a private chat message, first you need to initialize it with another peer in the network by typing:

```
start_privatechat <base64 public key of peer>
```

Once done, to send a private chat message to that peer:

```
privatechat <base64 public key of peer> <message>
```

## Tests

```sh
$ go test ./...
```

## TODO

- [X] basic chat functionality
- [X] cyptographic handshake using [ECDH](https://en.wikipedia.org/wiki/Elliptic-curve_Diffie%E2%80%93Hellman)
- [X] generate shared key using HKDF
- [X] message passing using Encrypt-then-MAC AEAD
- maintain overlay structure:
   - [X] handle join
   - [X] handle leave & failure
- [X] broadcast chat message
- [X] replicate chat log to new peer
- [X] private message
- [ ] tests
- [ ] lots of edge cases bugs

## References:
- [Chord: A Scalable Peer-to-peer Lookup Service for Internet
Applications](http://nms.csail.mit.edu/papers/chord.pdf)
- [How to Make Chord Correct](https://arxiv.org/pdf/1502.06461.pdf)
