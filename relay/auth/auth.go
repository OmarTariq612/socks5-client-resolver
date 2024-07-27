package auth

import (
	"fmt"
	"io"
)

const (
	NoAuthMethodRequiredCode byte = 0x00
	GSSAPICode               byte = 0x01
	UsernamePasswordCode     byte = 0x02
	NoAcceptableMethodCode   byte = 0xFF
)

type AuthHandler interface {
	Handle(clientConn, serverConn io.ReadWriter) error
}

type AuthHandlerFunc func(clientConn, serverConn io.ReadWriter) error

func (f AuthHandlerFunc) Handle(clientConn, serverConn io.ReadWriter) error {
	return f(clientConn, serverConn)
}

func GetAuthMethod(authCode int) (AuthHandler, error) {
	switch byte(authCode) {
	case NoAuthMethodRequiredCode:
		return AuthHandlerFunc(noAuthHandler), nil
	case UsernamePasswordCode:
		return AuthHandlerFunc(usernamePasswordHandler), nil
	}

	return nil, fmt.Errorf("unsupported auth method: %d", authCode)
}

func noAuthHandler(clientConn, serverConn io.ReadWriter) error {
	return nil
}
