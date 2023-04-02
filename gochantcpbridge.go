package gochantcpbridge

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/gob"
	"io"
	"log"
	"net"
	"time"
)

type CustomMessage struct {
	Type    string
	Content interface{}
}

type GoChannelTCPBridge struct {
	listenAddr      string
	remoteAddr      string
	sendChan        chan CustomMessage
	receiveChan     chan CustomMessage
	outboundConn    net.Conn
	inboundListener net.Listener
	certFile        string
	keyFile         string
}

func NewServer(listenAddr, certFile, keyFile string) (*GoChannelTCPBridge, error) {
	bridge := &GoChannelTCPBridge{
		listenAddr:  listenAddr,
		certFile:    certFile,
		keyFile:     keyFile,
		sendChan:    make(chan CustomMessage, 100),
		receiveChan: make(chan CustomMessage, 100),
	}

	listener, err := createSecureListener(listenAddr, certFile, keyFile)
	if err != nil {
		return nil, err
	}
	bridge.inboundListener = listener

	go bridge.runMainLoop()
	go bridge.acceptIncomingConnections()

	return bridge, nil
}

func NewClient(remoteAddr, certFile string) (*GoChannelTCPBridge, error) {
	bridge := &GoChannelTCPBridge{
		remoteAddr:  remoteAddr,
		certFile:    certFile,
		sendChan:    make(chan CustomMessage, 100),
		receiveChan: make(chan CustomMessage, 100),
	}

	go bridge.runMainLoop()
	go bridge.connectToRemoteAndRead()

	return bridge, nil
}

func createSecureListener(listenAddr, certFile, keyFile string) (net.Listener, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	config := &tls.Config{Certificates: []tls.Certificate{cert}}
	return tls.Listen("tcp", listenAddr, config)
}

func createSecureDialer(certFile string) (*tls.Config, error) {
	certPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		RootCAs: certPool,
	}, nil
}

func (bridge *GoChannelTCPBridge) runMainLoop() {
	for {
		msg := <-bridge.sendChan
		log.Println("Sending message over the wire: ", msg)
		if err := gob.NewEncoder(bridge.outboundConn).Encode(&msg); err != nil {
			log.Println(err)
		}
	}
}

func (bridge *GoChannelTCPBridge) acceptIncomingConnections() {
	defer bridge.inboundListener.Close()

	for {
		inboundConn, err := bridge.inboundListener.Accept()
		if err != nil {
			log.Println("Accept error:", err)
			return
		}

		log.Printf("Sender connected %s", inboundConn.RemoteAddr())

		go bridge.handleInboundConnection(inboundConn)
	}
}

func (bridge *GoChannelTCPBridge) handleInboundConnection(inboundConn net.Conn) {
	defer inboundConn.Close()
	decoder := gob.NewDecoder(inboundConn)

	for {
		var msg CustomMessage
		err := decoder.Decode(&msg)
		if err != nil {
			if err == io.EOF {
				log.Println("Connection closed by remote")
			} else {
				log.Println(err)
			}
			break
		}
		bridge.receiveChan <- msg
	}
}

func (bridge *GoChannelTCPBridge) connectToRemoteAndRead() {
	config, err := createSecureDialer(bridge.certFile)
	if err != nil {
		log.Fatal(err)
	}

	outboundConn, err := tls.DialWithDialer(&net.Dialer{Timeout: 3 * time.Second}, "tcp", bridge.remoteAddr, config)
	if err != nil {
		log.Printf("Dial error (%s)", err)
		time.Sleep(time.Second * 3)
		bridge.connectToRemoteAndRead()
		return
	}

	bridge.outboundConn = outboundConn
	defer outboundConn.Close()

	decoder := gob.NewDecoder(outboundConn)

	for {
		var msg CustomMessage
		err := decoder.Decode(&msg)
		if err != nil {
			if err == io.EOF {
				log.Println("Connection closed by remote")
			} else {
				log.Println(err)
			}
			time.Sleep(time.Second * 3)
			bridge.connectToRemoteAndRead()
			break
		}
		bridge.sendChan <- msg
	}
}

func (bridge *GoChannelTCPBridge) Send(msg CustomMessage) {
	bridge.sendChan <- msg
}

func (bridge *GoChannelTCPBridge) Receive() CustomMessage {
	return <-bridge.receiveChan
}

func (bridge *GoChannelTCPBridge) Close() {
	if bridge.outboundConn != nil {
		bridge.outboundConn.Close()
	}
	if bridge.inboundListener != nil {
		bridge.inboundListener.Close()
	}
}
