package relay

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

type Relay struct {
	bindAddr   string
	serverAddr string
}

func NewRelay(bindAddr, serverAddr string) *Relay {
	return &Relay{bindAddr: bindAddr, serverAddr: serverAddr}
}

const timeoutDuration = 5 * time.Second

func (r *Relay) Serve() error {
	listener, err := net.Listen("tcp", r.bindAddr)
	if err != nil {
		return err
	}
	defer listener.Close()
	log.Printf("Serving on %v\n", r.bindAddr)
	log.Printf("the provided server address is %v\n", r.serverAddr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("could not accept connections")
		}
		go func() {
			defer conn.Close()
			serverConn, err := net.DialTimeout("tcp", r.serverAddr, timeoutDuration)
			if err != nil {
				log.Printf("could not establish server connection, %v", err)
				return
			}
			defer serverConn.Close()
			err = handleConnection(conn, serverConn)
			if err != nil {
				log.Printf("connection failed, %v", err)
			}
		}()
	}
}

func handleConnection(clientConn, serverConn io.ReadWriter) error {
	err := relayHandshake(clientConn, serverConn)
	if err != nil {
		return err
	}

	err = relayRequestReply(clientConn, serverConn)
	if err != nil {
		return err
	}

	errc := make(chan error, 2)
	go func() {
		_, err := io.Copy(serverConn, clientConn)
		if err != nil {
			err = fmt.Errorf("could not copy from client to server, %v", err)
		}
		errc <- err
	}()
	go func() {
		_, err := io.Copy(clientConn, serverConn)
		if err != nil {
			err = fmt.Errorf("could not copy from server to client, %v", err)
		}
		errc <- err
	}()

	return <-errc
}

func relayHandshake(clientConn, serverConn io.ReadWriter) error {
	var buf [2]byte
	_, err := io.ReadFull(clientConn, buf[:])
	if err != nil {
		return fmt.Errorf("could not read handshake header (socks_version + n_methods) from the client")
	}
	_, err = serverConn.Write(buf[:])
	if err != nil {
		return fmt.Errorf("could not write handshake header (socks_version + n_methods) to the server")
	}
	methods := make([]byte, buf[1])
	_, err = io.ReadFull(clientConn, methods)
	if err != nil {
		return fmt.Errorf("could not read methods from the client")
	}
	_, err = serverConn.Write(methods)
	if err != nil {
		return fmt.Errorf("could not write methods to the server")
	}
	_, err = io.ReadFull(serverConn, buf[:])
	if err != nil {
		return fmt.Errorf("could not read handshake reply from the server")
	}
	_, err = clientConn.Write(buf[:])
	if err != nil {
		return fmt.Errorf("could not write handshake reply to the client")
	}
	return nil
}

type addrType byte

const (
	ipv4   addrType = 1
	domain addrType = 3
	ipv6   addrType = 4
)

func relayRequestReply(clientConn, serverConn io.ReadWriter) error {
	var buf [4]byte
	_, err := io.ReadFull(clientConn, buf[:])
	if err != nil {
		return fmt.Errorf("could not read request header from the client")
	}

	destAddrType := addrType(buf[3])

	if destAddrType == domain {
		_, err = serverConn.Write(buf[:3])
		if err != nil {
			return fmt.Errorf("could not write request header to the server")
		}
		var length [1]byte
		_, err = io.ReadFull(clientConn, length[:])
		if err != nil {
			return fmt.Errorf("could not read the length of the domain name from the client")
		}
		domain := make([]byte, length[0])
		_, err = io.ReadFull(clientConn, domain)
		if err != nil {
			return fmt.Errorf("could not read the domain name from the client")
		}
		ips, err := net.LookupIP(string(domain))
		if err != nil {
			return fmt.Errorf("could not resolve the domain name (%v)", string(domain))
		}
		var ip net.IP
		if ip = ips[0].To4(); ip != nil {
			_, err = serverConn.Write([]byte{byte(ipv4)})
		} else {
			ip = ips[0].To16()
			_, err = serverConn.Write([]byte{byte(ipv6)})
		}
		if err != nil {
			return fmt.Errorf("could not write the resolved address type")
		}
		_, err = serverConn.Write(ip)
		if err != nil {
			return fmt.Errorf("could not write the resolved address")
		}
	} else {
		var destination []byte
		if destAddrType == ipv4 {
			destination = make([]byte, 4)
		} else if destAddrType == ipv6 {
			destination = make([]byte, 16)
		} else {
			return fmt.Errorf("invalid address type -> (%v) <-", destAddrType)
		}

		_, err = serverConn.Write(buf[:])
		if err != nil {
			return fmt.Errorf("could not write request header to the server")
		}
		_, err = io.ReadFull(clientConn, destination[:])
		if err != nil {
			return fmt.Errorf("could not read destination address from the client")
		}
		_, err = serverConn.Write(destination[:])
		if err != nil {
			return fmt.Errorf("could now write destination address to the server")
		}
	}

	var portBuf [2]byte

	_, err = io.ReadFull(clientConn, portBuf[:])
	if err != nil {
		return fmt.Errorf("could not read the destination port from the client")
	}

	_, err = serverConn.Write(portBuf[:])
	if err != nil {
		return fmt.Errorf("could not write the destination port to the server")
	}

	return nil
}
