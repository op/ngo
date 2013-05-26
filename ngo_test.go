// Go binding for nanomsg

package ngo

import (
	"encoding/gob"
	"net"
	"testing"
)

func TestListenAndDial(t *testing.T) {
	// Dial out to the listener and send 10 integers in increasing size. We use
	// gob to encode the data sent on the wire for simplicity.
	go func() {
		conn, err := Dial("inproc", "req", "listen-serve")
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		e := gob.NewEncoder(conn)
		for i := 0; i < 10; i++ {
			e.Encode(i)
		}
		if err = conn.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	// Launch the listener and for each new message decode it using gob and make
	// sure the data retrieved is the expected.
	l, err := Listen("inproc", "rep", "listen-serve")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	for i := 0; i < 10; i++ {
		c, err := l.Accept()
		if err != nil {
			t.Fatal(err)
		}
		go func(i int, c net.Conn) {
			var j int
			defer c.Close()
			d := gob.NewDecoder(c)
			err := d.Decode(&j)
			if err != nil {
				t.Fatal(err)
			}
			if i != j {
				t.Error("%d != %d", i, j)
			}
		}(i, c)
	}
}
