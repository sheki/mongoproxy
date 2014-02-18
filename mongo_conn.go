package mongoproxy

import (
	"errors"
	"io"
	"net"
	"time"
)

// TODO add time outs
// Look at http://docs.mongodb.org/meta-driver/latest/legacy/mongodb-wire-protocol/ for the protocol.

const (
	opReply       = 1
	opMessage     = 1000
	opUpdate      = 2001
	opInsert      = 2002
	reserved      = 2003
	opQuery       = 2004
	opGetMore     = 2005
	opDelete      = 2006
	opKillCursors = 2007
)

// MsgHeader is the mongo MessageHeader
type MsgHeader struct {
	// MessageLength is the total message size, including this header
	MessageLength int32
	// RequestID is the identifier for this miessage
	RequestID int32
	// ResponseTo is the RequestID of the message being responded to. used in DB responses
	ResponseTo int32
	// OpCode is the request type, see consts above.
	OpCode int32
}

// WaitForResponse tells if mongo will reply to this message
func (m *MsgHeader) WaitForResponse() bool {
	return m.OpCode == opInsert || m.OpCode == opUpdate || m.OpCode == opDelete
}

// ToWire converts the MsgHeader to the wire protocol
func (m *MsgHeader) ToWire() []byte {
	b := make([]byte, 16)
	setInt32(b, 0, m.MessageLength)
	setInt32(b, 4, m.RequestID)
	setInt32(b, 8, m.ResponseTo)
	setInt32(b, 12, m.OpCode)
	return b
}

// FromWire reads the wirebytes into this object
func (m *MsgHeader) FromWire(b []byte) {
	m.MessageLength = getInt32(b, 0)
	m.RequestID = getInt32(b, 4)
	m.ResponseTo = getInt32(b, 8)
	m.OpCode = getInt32(b, 12)
}

// MongoConn provides operations on top of a mongo connection.
type MongoConn struct {
	conn net.Conn
	// ReadWriteTimeout  sets the read and write timeouets associated
	// with the connection
	ReadWriteTimeout time.Duration
}

// NewMongoConn creates a mongo connection using the provided connection
func NewMongoConn(c net.Conn) *MongoConn {
	return &MongoConn{conn: c}
}

// ReadHeader reads the MsgHeader from the mongo connection
func (m *MongoConn) ReadHeader() (*MsgHeader, error) {
	m.setDeadline()
	b := make([]byte, 16)
	if err := fill(m.conn, b); err != nil {
		return nil, err
	}
	h := MsgHeader{}
	h.FromWire(b)
	return &h, nil
}

var errWrite = errors.New("incorrect number of bytes written")

// WriteHeader writes a mongo MsgHeader into this connection
func (m *MongoConn) WriteHeader(h *MsgHeader) error {
	m.setDeadline()
	b := h.ToWire()
	n, err := m.conn.Write(b)
	if err != nil {
		return err
	}
	if n != len(b) {
		return errWrite
	}
	return nil
}

// CopyResponse copies one entire response from the given connection.
func (m *MongoConn) CopyResponse(fromConn *MongoConn) error {
	m.setDeadline()
	header, err := fromConn.ReadHeader()
	if err != nil {
		return err
	}
	if err := m.WriteHeader(header); err != nil {
		return err
	}
	return m.CopyN(fromConn, int64(header.MessageLength-16))
}

// CopyN copies n bytes from the fromConn into this connection.
func (m *MongoConn) CopyN(fromConn *MongoConn, n int64) error {
	m.setDeadline()
	written, err := io.CopyN(m.conn, fromConn.conn, n)
	if err != nil {
		return err
	}
	if written != n {
		return errWrite
	}
	return nil
}

func (m *MongoConn) setDeadline() {
	if m.ReadWriteTimeout != 0 {
		m.conn.SetDeadline(time.Now().Add(m.ReadWriteTimeout))
	}
}

// Close closes the connection
func (m *MongoConn) Close() error {
	return m.conn.Close()
}

// all data in the MongoDB wire protocol is little-endian.
// all the read/write functions below are little-endian.
func getInt32(b []byte, pos int) int32 {
	return (int32(b[pos+0])) |
		(int32(b[pos+1]) << 8) |
		(int32(b[pos+2]) << 16) |
		(int32(b[pos+3]) << 24)
}

func fill(r net.Conn, b []byte) error {
	l := len(b)
	n, err := r.Read(b)
	for n != l && err == nil {
		var ni int
		ni, err = r.Read(b[n:])
		n += ni
	}
	return err
}

func setInt32(b []byte, pos int, i int32) {
	b[pos] = byte(i)
	b[pos+1] = byte(i >> 8)
	b[pos+2] = byte(i >> 16)
	b[pos+3] = byte(i >> 24)
}
