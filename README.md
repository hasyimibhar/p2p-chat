# P2P Chat

This is a naive implementation of a p2p chat app for educational purposes.

```sh
# Run a network with 5 peers
$ go run . -port=8000
$ go run . -port=8001 -peer=localhost:8000
$ go run . -port=8002 -peer=localhost:8000
$ go run . -port=8003 -peer=localhost:8000
$ go run . -port=8004 -peer=localhost:8000
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

## References:
- [Chord: A Scalable Peer-to-peer Lookup Service for Internet
Applications](http://nms.csail.mit.edu/papers/chord.pdf)
- [How to Make Chord Correct](https://arxiv.org/pdf/1502.06461.pdf)
