// Gomatic nanomsg library
//
// This is experimental and might not work at all in practice.

package ngo

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"syscall"
	"time"

	nn "github.com/op/go-nanomsg"
)

type UnknownProtocolError string

func (e UnknownProtocolError) Error() string   { return "unknown protocol " + string(e) }
func (e UnknownProtocolError) Temporary() bool { return false }
func (e UnknownProtocolError) Timeout() bool   { return false }

type Dialer struct {
}

// Dial connects to the address using the given transport and protocol.
//
// Known protocols are "req", "rep", "pair".
func Dial(transport, protocol, address string) (net.Conn, error) {
	var d Dialer
	return d.Dial(transport, protocol, address)
}

func (d *Dialer) Dial(transport, protocol, address string) (net.Conn, error) {
	// TODO support transport to be eg. tcp+srv?
	// TODO rename transport to network and support the same convention as Go?
	socket, err := newSocket(protocol)
	if err != nil {
		return nil, err
	}
	_, err = socket.Connect(fmt.Sprintf("%s://%s", transport, address))
	return &conn{socket: socket, owner: true}, err
}

func newSocket(protocol string) (*nn.Socket, error) {
	var proto nn.Protocol
	switch protocol {
	case "req":
		proto = nn.REQ
	case "rep":
		proto = nn.REP
	case "pair":
		proto = nn.PAIR
	default:
		return nil, UnknownProtocolError(protocol)
	}
	return nn.NewSocket(nn.AF_SP, proto)
}

type conn struct {
	socket *nn.Socket
	owner  bool
	reader io.Reader
}

func (c *conn) Read(b []byte) (int, error) {
	if c.reader == nil {
		data, err := c.socket.Recv(0)
		if err != nil {
			return 0, err
		}
		c.reader = bytes.NewReader(data)
	}

	return c.reader.Read(b)
}
func (c *conn) Write(b []byte) (int, error) {
	// FIXME buffer stuff and expose Flush?
	return c.socket.Send(b, 0)
}

func (c *conn) Close() error {
	var err error
	if c.owner {
		err = c.socket.Close()
	}
	return err
}

func (c *conn) LocalAddr() net.Addr {
	panic("not implemeneted")
	return nil
}

func (c *conn) RemoteAddr() net.Addr {
	panic("not implemented")
	return nil
}

func (c *conn) SetDeadline(t time.Time) error {
	return syscall.ENOTSUP
}

func (c *conn) SetReadDeadline(t time.Time) error {
	return syscall.ENOTSUP
}

func (c *conn) SetWriteDeadline(t time.Time) error {
	return syscall.ENOTSUP
}

type listener struct {
	socket *nn.Socket
}

func Listen(transport, protocol, address string) (net.Listener, error) {
	socket, err := newSocket(protocol)
	if err != nil {
		return nil, err
	}
	_, err = socket.Bind(fmt.Sprintf("%s://%s", transport, address))
	return &listener{socket}, err
}

func (l *listener) Accept() (net.Conn, error) {
	data, err := l.socket.Recv(0)
	if err != nil {
		return nil, err
	}

	c := &conn{
		socket: l.socket,
		reader: bytes.NewReader(data),
	}
	return c, err
}

func (l *listener) Close() error {
	return l.socket.Close()
}

func (l *listener) Addr() net.Addr {
	return nil
}
