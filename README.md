# gochantcpbridge

A library which provides a simple, secure and efficient way to bridge Go channels over a TCP connection, with TLS support. It enables you to send and receive custom messages between a server and a client using Go channels.

## Features

* Simple API for sending and receiving messages over TCP
* TLS support for secure communication
* Efficient message encoding and decoding using gob
* Concurrent handling of connections and messages
* Separation of server and client instances for clear distinction

## Installation

To install the library, use the following command:

```shell
go get -u github.com/ozfive/gochantcpbridge
```

## Usage
First, import the package:

```go
import (
	"github.com/ozfive/gochantcpbridge"
)
```

### Server

To create a server instance, use the NewServer function, providing the listening address, the certificate file, and the key file for TLS:

```go
server, err := gochantcpbridge.NewServer("localhost:8000", "cert.pem", "key.pem")
if err != nil {
	log.Fatal(err)
}
```
### Client
To create a client instance, use the NewClient function, providing the remote server address and the certificate file for TLS:

```go
client, err := gochantcpbridge.NewClient("localhost:8000", "cert.pem")
if err != nil {
	log.Fatal(err)
}
```

### Sending and Receiving Messages

To send messages, use the Send function:

```go
msg := gochantcpbridge.CustomMessage{
	Type:    "example",
	Content: "Hello, World!",
}
client.Send(msg)
```

To receive messages, use the Receive function:

```go
receivedMsg := server.Receive()
fmt.Println("Received message:", receivedMsg)
```

### Closing Connections

To close the connections, use the Close function:

```go
server.Close()
client.Close()
```

## Example
An example application demonstrating the use of this library can be found in the `examples` directory of this repository.

## License

This project is licensed under the MIT License. See the LICENSE file for more information.

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.
