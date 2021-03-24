# JamJar Relay Server

This is a 'relay server', designed to allow self-hosted games to be run with the server acting as an intermediary
between all of the clients. This allows for a generic networking solution that allows games to be hosted by a client
and all messages are forwarded, or 'relayed' to other clients through the server.

## Development

### Dependencies

- [Golang](https://golang.org/doc/install) `>= 1.14.6`
- Golint
- [Protoc](http://google.github.io/proto-lens/installing-protoc.html) `== 3.15.6`
- [Golang Protobuf Plugin](https://developers.google.com/protocol-buffers/docs/reference/go-generated) `== 1.26.0`

### Commands

- `make run` - Run the server locally on port `8000`.
- `make cli` - Run the test CLI for interacting with the local server.
- `make generate` - Generates all the Go code from the protobuf specs.
