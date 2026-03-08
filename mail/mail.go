package mail

import (
	"fmt"
	"time"
	"mime"
	"strings"
	"net/mail"
	"crypto/rand"
	"encoding/hex"
	"golang.org/x/net/idna"
)

const (
	UTF8	= "UTF-8"
	CRLF	= "\r\n"
)

type Mail struct {
	to_email	string
	to_name		string
	from_email	string
	from_name	string
	subject		string
	body		string
	html		string
}

func NewMail() *Mail {
	return &Mail{}
}

func (m *Mail) To(email, name string){
	m.to_email		= email
	m.to_name		= name
}

func (m *Mail) From(email, name string) {
	m.from_email	= email
	m.from_name		= name
}

func (m *Mail) Subject(subject string){
	m.subject = subject
}

func (m *Mail) Body(body string){
	m.body = body
}

func (m *Mail) HTML(html string){
	m.html = "<html><body>"+html+"</body></html>"
}

func (m *Mail) Message() (string, error){
	var err error
	m.to_email, err = punycode(m.to_email)
	if err != nil {
		return "", err
	}
	return m.source(), nil
}

func (m *Mail) source() string {
	addr_from := mail.Address{
		Name:		m.from_name,
		Address:	m.from_email,
	}
	
	addr_to := mail.Address{
		Name:		m.to_name,
		Address:	m.to_email,
	}
	
	boundary := fmt.Sprintf("boundary_%d", time.Now().UnixNano())
	
	var b strings.Builder
	
	b.WriteString("Return-Path: <")
	b.WriteString(m.from_email)
	b.WriteByte('>')
	b.WriteString(CRLF)
	
	b.WriteString("Date: ")
	b.WriteString(time.Now().Format(time.RFC1123Z))
	b.WriteString(CRLF)
	
	b.WriteString("From: ")
	b.WriteString(addr_from.String())
	b.WriteString(CRLF)
	
	b.WriteString("To: ")
	b.WriteString(addr_to.String())
	b.WriteString(CRLF)
	
	b.WriteString("Message-ID: ")
	b.WriteString(m.message_id())
	b.WriteString(CRLF)
	
	b.WriteString("Subject: ")
	b.WriteString(mime.QEncoding.Encode(UTF8, m.subject))
	b.WriteString(CRLF)
	
	b.WriteString("MIME-Version: 1.0")
	b.WriteString(CRLF)
	
	if m.html != "" && m.body != "" {
		b.WriteString("Content-Type: multipart/alternative; boundary=")
		b.WriteString(boundary)
		b.WriteString(CRLF)
		b.WriteString(CRLF)
		
		b.WriteString("--")
		b.WriteString(boundary)
		b.WriteString(CRLF)
		
		b.WriteString("Content-Type: text/plain; charset=")
		b.WriteString(UTF8)
		b.WriteString(CRLF)
		
		b.WriteString("Content-Transfer-Encoding: 8bit")
		b.WriteString(CRLF)
		b.WriteString(CRLF)
		
		b.WriteString(m.body)
		b.WriteString(CRLF)
		
		b.WriteString("--")
		b.WriteString(boundary)
		b.WriteString(CRLF)
		
		b.WriteString("Content-Type: text/html; charset=")
		b.WriteString(UTF8)
		b.WriteString(CRLF)
		
		b.WriteString("Content-Transfer-Encoding: 8bit")
		b.WriteString(CRLF)
		b.WriteString(CRLF)
		
		b.WriteString(m.html)
		b.WriteString(CRLF)
		
		b.WriteString("--")
		b.WriteString(boundary)
		b.WriteString("--")
		b.WriteString(CRLF)
	} else {
		b.WriteString("Content-Type: ")
		if m.html != "" {
			b.WriteString("text/html; charset=")
		} else {
			b.WriteString("text/plain; charset=")
		}
		b.WriteString(UTF8)
		b.WriteString(CRLF)
		
		b.WriteString("Content-Transfer-Encoding: 8bit")
		b.WriteString(CRLF)
		b.WriteString(CRLF)
		
		if m.html != "" {
			b.WriteString(m.html)
		} else {
			b.WriteString(m.body)
		}
		b.WriteString(CRLF)
	}
	return b.String()
}

func (m *Mail) message_id() string {
	at := strings.LastIndex(m.from_email, "@")
	domain := "localhost"
	if at != -1 {
		domain = m.from_email[at+1:]
	}
	timestamp := time.Now().UnixNano()
	b := make([]byte, 8)
	rand.Read(b)
	random := hex.EncodeToString(b)
	return fmt.Sprintf("<%d.%s@%s>", timestamp, random, domain)
}

func punycode(email string) (string, error){
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "", fmt.Errorf("Invalid email format")
	}
	
	alias	:= parts[0]
	domain	:= parts[1]
	
	puny_domain, err := idna.ToASCII(domain)
	if err != nil {
		return "", err
	}
	
	return alias+"@"+puny_domain, nil
}