// Package mailer is a simple e-mail sender for the Go Programming Language.
package mailer

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/mail"
	"net/smtp"
	"os/exec"
	"strings"

	"github.com/valyala/bytebufferpool"
)

const (
	// Version current version semantic number of the "go-mailer" package.
	Version = "0.1.0"
)

// Mailer is the main struct which contains the nessecary fields
// for sending emails, either with unix command "sendmail"
// or by following the configuration's properties.
type Mailer struct {
	config        Config
	fromAddr      mail.Address
	auth          smtp.Auth
	authenticated bool
}

// New creates and returns a new mail sender.
func New(cfg Config) *Mailer {
	m := &Mailer{config: cfg}
	addr := cfg.FromAddr
	if addr == "" {
		addr = cfg.Username
	}

	if cfg.FromAlias == "" {
		if !cfg.UseCommand && cfg.Username != "" && strings.Contains(cfg.Username, "@") {
			m.fromAddr = mail.Address{Name: cfg.Username[0:strings.IndexByte(cfg.Username, '@')], Address: addr}
		}
	} else {
		m.fromAddr = mail.Address{Name: cfg.FromAlias, Address: addr}
	}
	return m
}

// UpdateConfig overrides the current configuration.
func (m *Mailer) UpdateConfig(cfg Config) {
	m.config = cfg
}

// SendWithReader same as `Send` but it accepts
// an io.Reader that body can be retrieved and call the `SendWithBytes`.
func (m *Mailer) SendWithReader(subject string, bodyReader io.Reader, to ...string) error {
	// Good idea but not expected by the user, she/he can just use the `SendWithBytes`
	// and `Send` if have a 'byter' or a 'stringer', so the below are not important.
	//
	// // check if it can return string data directly without reading, (i.e bytes.Buffer).
	// if bytesReader, ok := bodyReader.(byter); ok {
	// 	return m.SendWithBytes(subject, bytesReader.Bytes(), to...)
	// }
	// // check if it can return string data directly without reading, (i.e bytes.Buffer).
	// if stringReader, ok := bodyReader.(stringer); ok {
	// 	return m.Send(subject, stringReader.String(), to...)
	// }

	// read the data and send with bytes.
	body, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		return err
	}

	// we could check if the `bodyReader` is an io.Closer and close its body
	// but the caller may need to manage this by itself, so we don't do it here
	// we introduce a new function, `SendWithReadCloser`.
	return m.SendWithBytes(subject, body, to...)
}

// SendWithReadCloser same as `SendWithReader` but it closes the reader at the end.
func (m *Mailer) SendWithReadCloser(subject string, bodyReader io.ReadCloser, to ...string) error {
	body, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		return err
	}

	if err = bodyReader.Close(); err != nil {
		return err
	}

	return m.SendWithBytes(subject, body, to...)
}

// Send sends an email to the recipient(s)
// the body can be in HTML format as well.
//
// Note: you can change the UseCommand in runtime.
func (m *Mailer) Send(subject string, body string, to ...string) error {
	return m.SendWithBytes(subject, []byte(body), to...)
}

// SendWithBytes same as `Send` but it accepts the body as raw []byte,
// it's the fastest method to send e-mails.
func (m *Mailer) SendWithBytes(subject string, body []byte, to ...string) error {
	if m.config.UseCommand {
		return m.sendUnix(subject, body, to)
	}

	return m.sendSMTP(subject, body, to)
}

const (
	contentTypeHTML         = `text/html; charset=\"utf-8\"`
	mimeVer                 = "1.0"
	contentTransferEncoding = "base64"
)

type stringWriter interface {
	WriteString(string) (int, error)
}

func writeHeaders(w stringWriter, subject string, body []byte, to []string) {
	w.WriteString(fmt.Sprintf("%s: %s\r\n", "To", strings.Join(to, ",")))
	w.WriteString(fmt.Sprintf("%s: %s\r\n", "Subject", subject))
	w.WriteString(fmt.Sprintf("%s: %s\r\n", "MIME-Version", mimeVer))
	w.WriteString(fmt.Sprintf("%s: %s\r\n", "Content-Type", contentTypeHTML))
	w.WriteString(fmt.Sprintf("%s: %s\r\n", "Content-Transfer-Encoding", contentTransferEncoding))
	w.WriteString(fmt.Sprintf("\r\n%s", base64.StdEncoding.EncodeToString(body)))
}

var bufPool bytebufferpool.Pool

func (m *Mailer) sendSMTP(subject string, body []byte, to []string) error {
	buffer := bufPool.Get()
	defer bufPool.Put(buffer)

	if !m.authenticated {
		cfg := m.config
		if cfg.Username == "" || cfg.Password == "" || cfg.Host == "" || cfg.Port <= 0 {
			return fmt.Errorf("username, password, host or port missing")
		}
		m.auth = smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
		m.authenticated = true
	}

	fullhost := fmt.Sprintf("%s:%d", m.config.Host, m.config.Port)

	buffer.WriteString(fmt.Sprintf("%s: %s\r\n", "From", m.fromAddr.String()))
	writeHeaders(buffer, subject, body, to)

	return smtp.SendMail(
		fmt.Sprintf(fullhost),
		m.auth,
		m.config.Username,
		to,
		buffer.Bytes(),
	)
}

func (m *Mailer) sendUnix(subject string, body []byte, to []string) error {
	buffer := new(bytes.Buffer)
	writeHeaders(buffer, subject, body, to)

	cmd := exec.Command("sendmail", "-F", m.fromAddr.Name, "-f", m.fromAddr.Address, "-t")
	cmd.Stdin = buffer
	_, err := cmd.CombinedOutput()
	return err
}
