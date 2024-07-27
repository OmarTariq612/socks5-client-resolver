package auth

import (
	"fmt"
	"io"
)

func usernamePasswordHandler(clientConn, serverConn io.ReadWriter) error {
	var buf [512]byte
	if _, err := io.ReadFull(clientConn, buf[:2]); err != nil {
		return err
	}
	usernameLen := int(buf[1])
	if _, err := io.ReadFull(clientConn, buf[2:usernameLen+3]); err != nil {
		return err
	}
	passwordLen := int(buf[2+usernameLen])
	if _, err := io.ReadFull(clientConn, buf[3+usernameLen:3+usernameLen+passwordLen]); err != nil {
		return err
	}
	if _, err := serverConn.Write(buf[:3+usernameLen+passwordLen]); err != nil {
		return err
	}

	if _, err := io.ReadFull(serverConn, buf[:2]); err != nil {
		return err
	}

	if _, err := clientConn.Write(buf[:2]); err != nil {
		return err
	}

	if buf[1] != 0 {
		return fmt.Errorf("username/password authentication failed")
	}

	return nil
}
