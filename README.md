# P2P Chat

This is a naive implementation of a p2p chat app for educational purposes.

TODO:

- [X] basic chat functionality
- [X] cyptographic handshake using [ECDH](https://en.wikipedia.org/wiki/Elliptic-curve_Diffie%E2%80%93Hellman)
- [X] generate shared key using HKDF
- [X] message passing using Encrypt-then-MAC AEAD
- maintain overlay structure:
   - [X] handle join
   - [ ] handle leave
   - [ ] handle failure
- [X] broadcast chat message
- [ ] private message
- [ ] tests
