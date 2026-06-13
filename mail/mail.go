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
	to_email		string
	to_name			string
	from_email		string
	from_name		string
	reply_to_email	string
	reply_to_name	string
	subject			string
	body			string
	html			string
	unsubscribe_url	string
}

func NewMail() *Mail {
	return &Mail{}
}

func (m *Mail) To(email, name string){
	m.to_email			= email
	m.to_name			= name
}

func (m *Mail) From(email, name string){
	m.from_email		= email
	m.from_name			= name
}

func (m *Mail) Reply_to(email, name string){
	m.reply_to_email	= email
	m.reply_to_name		= name
}

func (m *Mail) Subject(subject string){
	m.subject			= subject
}

func (m *Mail) Body(body string){
	m.body				= body
}

func (m *Mail) HTML(html string){
	m.html				= html
}

func (m *Mail) Unsubscribe(url string){
	m.unsubscribe_url	= url
}

func (m *Mail) Message() (string, error){
	from_email, err := valid_email(m.from_email)
	if err != nil {
		return "", err
	}
	
	to_email, err := valid_email(m.to_email)
	if err != nil {
		return "", err
	}
	
	boundary := fmt.Sprintf("boundary_%d", time.Now().UnixNano())
	
	var b strings.Builder
	
	b.WriteString("Return-Path: <")
	b.WriteString(from_email)
	b.WriteByte('>')
	b.WriteString(CRLF)
	
	b.WriteString("Date: ")
	b.WriteString(time.Now().UTC().Format(time.RFC1123Z))
	b.WriteString(CRLF)
	
	b.WriteString("From: ")
	b.WriteString(format_address(m.from_name, from_email))
	b.WriteString(CRLF)
	
	b.WriteString("To: ")
	b.WriteString(format_address(m.to_name, to_email))
	b.WriteString(CRLF)
	
	if m.reply_to_email != "" {
		reply_to_email, err := valid_email(m.reply_to_email)
		if err != nil {
			return "", err
		}
		
		b.WriteString("Reply-To: ")
		b.WriteString(format_address(m.reply_to_name, reply_to_email))
		b.WriteString(CRLF)
	}
	
	msg_id, err := message_id(from_email)
	if err != nil {
		return "", err
	}
	
	b.WriteString("Message-ID: ")
	b.WriteString(msg_id)
	b.WriteString(CRLF)
	
	if m.unsubscribe_url != "" {
		b.WriteString("List-Unsubscribe: <")
		b.WriteString(m.unsubscribe_url)
		b.WriteString(">")
		b.WriteString(CRLF)
		
		b.WriteString("List-Unsubscribe-Post: List-Unsubscribe=One-Click")
		b.WriteString(CRLF)
	}
	
	b.WriteString("Subject: ")
	b.WriteString(mime.BEncoding.Encode(UTF8, sanitize_header(m.subject)))
	b.WriteString(CRLF)
	
	b.WriteString("MIME-Version: 1.0")
	b.WriteString(CRLF)
	
	if m.html != "" && m.body != "" {
		b.WriteString(`Content-Type: multipart/alternative; boundary="`)
		b.WriteString(boundary)
		b.WriteString(`"`)
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
	return b.String(), nil
}

func message_id(email string) (string, error){
	at := strings.LastIndex(email, "@")
	domain := "localhost"
	if at != -1 {
		domain = email[at+1:]
	}
	timestamp := time.Now().UnixNano()
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	random := hex.EncodeToString(b)
	return fmt.Sprintf("<%d.%s@%s>", timestamp, random, domain), nil
}

func format_address(name, email string) string {
	name = sanitize_header(name)
	if name == "" {
		return email
	}
	addr := mail.Address{
		Name:		name,
		Address:	email,
	}
	return addr.String()
}

func valid_email(email string) (string, error){
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return "", fmt.Errorf("Invalid email format")
	}
	
	alias, domain, _ := strings.Cut(addr.Address, "@")
	
	puny_domain, err := idna.ToASCII(domain)
	if err != nil {
		return "", err
	}
	
	return alias+"@"+puny_domain, nil
}

func sanitize_header(s string) string {
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")
	return s
}