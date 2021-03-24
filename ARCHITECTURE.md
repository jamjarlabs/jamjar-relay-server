# Architecture

This gives an overview of the relay server's architecture, this is not required to use the relay server; it is only
useful for development.

## Architecture v1

The architecture of the relay server is split up into distinct components:

- API routing
- Websocket handler
- Rooms HTTP handler
- Protocol
- Room manager
- Room
- Session

### API routing

The API routing handles routing requests based on the URL path, e.g. `/v1/websocket` routes to the websocket handler,
`/v1/api/rooms` routes to the rooms HTTP handler.

### Websocket handler

The websocket handler is used to manage websocket connections, controlling reading, writing and closing the connection.
This handler is responsible for upgrading the connection to a websocket connection, and also setting up a session
to track the connection, allowing other components to write to the websocket or to close the websocket.

### Rooms HTTP handler

The rooms HTTP handler is used to manage HTTP requests for manipulating rooms. This handler controls reading requests
and writing responses.

### Protocol

A protocol is used to define protocol specific behaviour, for example what a host should be allowed to do, what to
do when a room is closed, message relaying behaviour etc.

### Room manager

A room manager is used to maintain a centralised state of rooms, allowing creation, reading, updating, and deleting
rooms. The room manager also provides some common utility methods for managing rooms, such as generating a combined
summary of all the rooms in the room manager.

### Room

A room is used to track state of a grouping of connected client sessions. This is used to group together clients and
mark certain clients with extra privileges (e.g. host powers).

### Session

A session is used to track a connection, allowing a safe way to write messages from the relay to the user, and a way
to close connections. The sesssion is connection agnostic, allowing any continuous connection to be used. The session
also keeps track of client identifiers, allowing a connection to be identified within a room.

### Flows

These are some example flows showing how components interact.

#### User connects and sends message

This flow shows a user connecting to the relay using websockets, connecting to a room before broadcasting a message.

1. User makes a request to connect to the relay server, the API routing handles this, routing to the websocket handler.
2. The websocket handler then parses the request, upgrading the connection and setting up the following:
    - A goroutine to manage closing the connection.
    - A goroutine to manage writing messages.
    - A goroutine to manage reading messages.
    - A new session to manage the connection, allowing closing and writing to the connection using the goroutines
    defined.
3. The websocket handler will then wait until the user sends a message.
4. The user sends a connect to room message, the websocket handler reads the message and parses it, before routing it
to the protocol to handle the connection request.
5. The protocol handles this connect message, looking up the room using the room manager and validating it (checking
room exists, secret matches etc). The protocol will also handle some protocol specific details, such as determining
if the user should be granted host privledges. The protocol will write any output messages using the session to
send the serialised message bytes for the websocket handler to write out.
6. The websocket handler will then wait for the next message.
7. The user sends a relay broadcast room message, the websocket handler reads the message and parses it, before routing
it to the protocol to handle relaying the request.
8. The protocol handles this relay broadcast, checking some protocol specific details (is the user a host, can they
broadcast) before retrieving a list of connected sessions and writing the broadcasted messages to each session's
write goroutine. The websocket handlers will then handle sending these messages to each user in the room.
9. The websocket handler will then wait for the next message...

#### Server creates a room

This flow shows an HTTP request being made to the relay API, creating a new room.

1. A HTTP request is made to the relay server, the API routing handles this, routing to the rooms HTTP handler.
2. The rooms HTTP handler then parses the request, before routing to the protocol to handle room creation.
3. The protocol then uses the room manager to create a new room.
4. The room manager generates a new room using its room factory, adding it to its internal state.
5. The room is then returned up the chain, until the rooms HTTP handler serialises the room's info and writes it
out as a response.
