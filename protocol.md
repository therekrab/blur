# Message protocol
The following is a detailed description of the standard message protocol for
this application. It is currently without a name, but that's in the works.

## A note on the use of TCP
Because the entire protocol uses TCP, the protocol itself can omit parts that
would include acknowledgement of a message (SYN, SYN/ACK, ACK, etc.). This
allows the protocol itself to be a lot simpler and concise.

Also, TCP's verification system means that we're a lot less likely to send or
receive bad data, which can be really problematic.

## Message names
I have chosen to use the following message names in the protocol:

*Note*: a `?` at the end of a type indicates that it is a *request*, whereas
the lack of a `?` indicates that it is a response or message and does not
require a response.

### `JOIN?` (0)
A request sent by the client. The first 2 bytes of the request
data will be the session ID, and it will be followed with the SHA256 hash of
the session key provided by the user.

### `ACC` (1)
The response to a `JOIN?` request that indicates that the provided session ID
does exist and the credentials supplied were valid. The data portion of an
`ACC` response will be blank, so the length will be 0 bytes. 

### `REJ` (2)
The response to a `JOIN?` request that rejects the attempt at joining the
indicated session. This could be due to the session id being wrong, or the
session key provided is incorrect. If the former is true, then the data
portion of the message will contain `0`, and the latter will be indicated by
a `1` in the data portion. This means that the data portion will have a length
of 1 byte. A `REJ` request also signals the end of any further communications
and the connection is closed by the server.

### `NEW?` (3)
This is a request from a client to a server, and signals that the client would
like to start a new session. The data field of the request will be the SHA256
hash of the session key to be set. If the server cannot create a new session,
an `ERR` message will be sent back instead of a `NEW` response.

### `NEW` (4)
This is a response from a server, and signals that a new session has been
successfully created, and the client is connected to it. The hash that the user
provided to the server in the `NEW?` request was set as the authentication hash
for the session. The session id is contained in the data field of the response.

### `IDENT?` (5)
This request can be sent by either a server or a client. If the client is
the sender, then the server should respond with the identifier of each other
user in the session. If the server is the sender, then the client should reply
with the user's identifier. The response to an `IDENT?` request is always an
`IDENT` response. Because the message carries no information, the data portion
is empty.

### `IDENT` (6)
This is always returned as response to an `IDENT?` request, and it will contain
the identifier(s) requested. The data portion is of variable length, as it can
hold an varying amount of information.

### `CHT` (7)
This is the message format used for actual user-to-user communications. After
a server received a `CHT` request, it will pass it on to all other users in the
session of the sender. The same message is sent to each client, where it will
be displayed to the user.

### `CHTE` (8)
This is identical in function to the `CHT` format, except it lets the client
program know that the message is encrypted, and so they should decrypt it.

### `ERR` (9)
This message signals that an error has occured in the system managing the
connection - not a user error. This would occur if a client attempts to run
a request besides `JOIN?` or `NEW?` to start a communication. The data field
for this message will contain a (non-encrypted) string that should provide
debug information to help fix the problems.

## The protocol itself
Upon establishing a connection, the client is responsible for initiating
communication. The client will begin by sending a `JOIN?` or `NEW?` message,
depending on the desired behavior.

### Creating a new session
If a `NEW?` request is sent, the server will respond with either a `NEW`
response (OK) or an `ERR` message (error, connection closed).

### Joining a session
If a `JOIN?` request is sent, the server will reply with either an `ACC`
message (the user's credentials are valid) or a `REJ` message (rejected,
connection closed).

If the connection is still activated - ie an `ACC` or `NEW` response - the
session is now __authenticated__. This means that the server can now send
`IDENT`, `IDENTR`, or `CHT(E)` messages.

### Sending messages
To send a message, the client will send a `CHT(E)` message to the server, which
will broadcast the message to all other users in the session through another
`CHT(E)` message. A client receiving a `CHTE` message should decrypt the message
and display it to the user.

## Message structure

The structure of all messages is the same, regardless of message type:

### `DSIZE` field
The `DSIZE` field is a 2-byte field that stores the length of the `DATA` field.
It is *not* the actual size of the message. The bytes should be stored in big-
endian order, as that is the standard for most networking protocols.

### `MTYPE` field
The `MTYPE` field is a single byte that specifies the message type. See above
for the list of types and their respecive codes.

### `DATA` field
This field, which is `DSIZE` bytes long, contains all of the bytes of the
message. The `DATA` field is plain text, except for `CHTE` messages, where the
data is encrypted with the session key.

__`DATA` field in `IDENT` responses__

The `DATA` field in `IDENT` responses is slightly unique, as it has to provide
a number of identifiers. The data field should instead be a series of alternating
`DSIZE` fields and regular `DATA` fields of that length,
terminated by a `DSIZE` field of length 0.

__`DATA` field in `CHT(E)` responses__

Like in `IDENT` responses, the `DATA` field is slightly unique in `CHT(E)`
messages. When a client sends a `CHT(E)` message, the `DATA` portion is just
the message to be sent. However, when the server forwards the message to all
clients, the `DATA` field will be slightly different. It will be prepended by a
`DSIZE`/`DATA` pair that will specify the identity of the sender, which will
be identified from the client's `IDENT` response earlier.
