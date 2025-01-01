![Blur](logo.png)


Blur is a communications software for those who value security above all else.

Here are some of its features:
* End-to-end encryption (less of a feature and more of an obligation). Blur
uses AES to encrypt and decrypt all messages between users on the client side.
This means that even the server itself cannot read the messages you send on it.
* Local servers. Blur is built for those that cannot rely on anybody else to
run a secure server. Even though the server cannot read messages, it still can
keep a record of connections. So if have to rely on full anonimity, why not
run the server yourself? And it's easy too. Blur has a built in server mode
that enables it to run as a server rather than a client.
* Portable. Blur is just a single binary once compiled. This means that after
compilation, blur can exist as a single file and leave no other trace on the
system.
* No trace on client computer. Blur exists in memory only. When you leave Blur,
anything that was on your screen is lost. There are no files that store chats.
No vulnerabilites to fear. When you exit the program, your computer will have
no idea what you did.
* TUI interface. There already are so many E2EE softwares out there, but Blur
is different. It runs completely in the terminal. I think this makes it cooler.

## Installation
Installation is pretty simple.
```
$ go install github.com/therekrab/blur
```
Note: `go install` automatically installs binaries to `$GOPATH/bin`. If this is
unset, it will instead go with `$HOME/go/bin`. Wherever it installs, make sure
that it's on `PATH` or running `blur` will fail.

## Server
Blur can run as a server with the `-server` flag.
```
$ blur -server
```
The `-port` flag specifies what port to listen on, if applicable. Default is
`4040`

The server automatically keeps logs of connections and sessions in the
`blur.log` file in the current directory the application was ran in.
Alternatively, if the `BLURDIR` environment variable is set, the log file will
be written to that directory.

## Client
The following flags are important to know as a client.

`-addr` : The full address of the server, including port.
```
$ blur -addr 10.0.0.2:4040
```
Replace `10.0.0.2` with the address of the actual server.

`-new`: This flag, when supplied, directs blur to create a new session rather
than join a prexisting one. This will create a one-time session ID that should
be shared. __Sesssion IDs are NOT permanent between sessions.__ This means that
when you run `blur -new` and get your session ID, you must keep that client
open to continue the session under that ID. To minimize server memory usage,
an empty session is automatically trashed. So keep your sessions open.

## Misc
Blur supports the [NO_COLOR](https://no-color.org) environment variable.
