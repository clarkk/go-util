package websocket

import (
	"io"
	"fmt"
	"bufio"
	"strings"
	"net"
	"net/http"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
)

const websocket_guid = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

type Client struct {
	conn	net.Conn
	rw		*bufio.ReadWriter
}

func Handler(w http.ResponseWriter, r *http.Request){
	if r.Header.Get("Upgrade") != "websocket" {
		http.Error(w, "Invalid websocket request", http.StatusBadRequest)
		return
	}
	
	key := r.Header.Get("Sec-WebSocket-Key")
	if key == "" {
		http.Error(w, "Missing websocket key", http.StatusBadRequest)
		return
	}
	
	h, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Websocket unsupported", http.StatusInternalServerError)
		return
	}
	
	conn, rw, err := h.Hijack()
	if err != nil {
		return
	}
	
	hash	:= sha1.Sum([]byte(key + websocket_guid))
	accept	:= base64.StdEncoding.EncodeToString(hash[:])
	header	:= "HTTP/1.1 101 Switching Protocols\r\n"+
		"Upgrade: websocket\r\n"+
		"Connection: Upgrade\r\n"+
		"Sec-WebSocket-Accept: %s\r\n"+
		"\r\n"
	
	fmt.Fprintf(rw, header, accept)

	if err := rw.Flush(); err != nil {
		conn.Close()
		return
	}
	
	client := &Client{
		conn:	conn,
		rw:		rw,
	}
	
	client.run()
}

func (c *Client) run(){
	defer c.conn.Close()
	
	for {
		message, err := c.read()
		if err != nil {
			return
		}
		
		fmt.Printf("WebSocket: %s\n", message)
		
		if err := c.write([]byte(`{"ok":true}`)); err != nil {
			return
		}
	}
}

func (c *Client) read() ([]byte, error){
	var header [2]byte
	
	if _, err := io.ReadFull(c.rw, header[:]); err != nil {
		return nil, err
	}
	
	final	:= header[0]&0x80 != 0
	opcode	:= header[0] & 0x0f
	
	if !final {
		return nil, fmt.Errorf("Fragmented messages are not supported")
	}
	
	switch opcode {
	case 0x1:
		//	Text
		
	case 0x8:
		return nil, io.EOF
		
	case 0x9:
		payload, err := c.read_payload(header[1])
		if err != nil {
			return nil, err
		}
		
		if err := c.write_frame(0xA, payload); err != nil {
			return nil, err
		}
		
		return c.read()
		
	default:
		return nil, fmt.Errorf("Unsupported opcode: %d", opcode)
	}
	
	return c.read_payload(header[1])
}

func (c *Client) read_payload(second_byte byte) ([]byte, error){
	masked := second_byte&0x80 != 0
	if !masked {
		return nil, fmt.Errorf("Client message is not masked")
	}
	
	length := uint64(second_byte & 0x7f)
	
	switch length {
	case 126:
		var data [2]byte
		if _, err := io.ReadFull(c.rw, data[:]); err != nil {
			return nil, err
		}
		
		length = uint64(binary.BigEndian.Uint16(data[:]))
		
	case 127:
		var data [8]byte
		if _, err := io.ReadFull(c.rw, data[:]); err != nil {
			return nil, err
		}
		
		length = binary.BigEndian.Uint64(data[:])
	}
	
	if length > 1024*1024 {
		return nil, fmt.Errorf("message too large")
	}
	
	var mask [4]byte
	if _, err := io.ReadFull(c.rw, mask[:]); err != nil {
		return nil, err
	}
	
	payload := make([]byte, int(length))
	if _, err := io.ReadFull(c.rw, payload); err != nil {
		return nil, err
	}
	
	for i := range payload {
		payload[i] ^= mask[i%4]
	}
	
	return payload, nil
}

func (c *Client) write(payload []byte) error {
	return c.write_frame(0x1, payload)
}

func (c *Client) write_frame(opcode byte, payload []byte) error {
	header := []byte{
		0x80 | opcode,
	}
	
	switch length := len(payload); {
	case length <= 125:
		header = append(header, byte(length))
		
	case length <= 65535:
		header = append(
			header,
			126,
			byte(length>>8),
			byte(length),
		)
		
	default:
		header = append(
			header,
			127,
			byte(uint64(length)>>56),
			byte(uint64(length)>>48),
			byte(uint64(length)>>40),
			byte(uint64(length)>>32),
			byte(uint64(length)>>24),
			byte(uint64(length)>>16),
			byte(uint64(length)>>8),
			byte(uint64(length)),
		)
	}
	
	if _, err := c.rw.Write(header); err != nil {
		return err
	}
	
	if _, err := c.rw.Write(payload); err != nil {
		return err
	}
	
	return c.rw.Flush()
}