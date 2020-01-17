# Mailer

Simple E-mail sender written in Go.
Mailer supports rich e-mails and, optionally, *nix built'n `sendmail` command.



<a href="https://travis-ci.org/kataras/go-mailer"><img src="https://img.shields.io/travis/kataras/go-mailer.svg?style=flat-square" alt="Build Status"></a>
<a href="https://github.com/kataras/go-mailer/blob/master/LICENSE"><img src="https://img.shields.io/badge/%20license-MIT%20%20License%20-E91E63.svg?style=flat-square" alt="License"></a>
<a href="https://github.com/kataras/go-mailer/releases"><img src="https://img.shields.io/badge/%20release%20-%20v0.1.0-blue.svg?style=flat-square" alt="Releases"></a>
<a href="https://godoc.org/github.com/kataras/go-mailer"><img src="https://img.shields.io/badge/%20docs-reference-5272B4.svg?style=flat-square" alt="Godocs"></a>
<a href="https://kataras.rocket.chat/channel/go-mailer"><img src="https://img.shields.io/badge/%20community-chat-00BCD4.svg?style=flat-square" alt="Build Status"></a>
<a href="https://golang.org"><img src="https://img.shields.io/badge/powered_by-Go-3362c2.svg?style=flat-square" alt="Built with GoLang"></a>
<a href="#"><img src="https://img.shields.io/badge/platform-Any--OS-yellow.svg?style=flat-square" alt="Platforms"></a>

## Installation

The only requirement is the [Go Programming Language](https://golang.org/dl).

```bash
$ go get -u github.com/kataras/go-mailer
```

## Getting Started

- `New` returns a new, e-mail sender service.
- `Mailer#Send` send an e-mail, supports text/html and `sendmail` unix command

```go
// New returns a new *Mailer, which contains the Send methods.
New(cfg Config) *Mailer
```

```go
// SendWithBytes same as `Send` but it accepts the body as raw []byte,
// it's the fastest method to send e-mails.
SendWithBytes(subject string, body []byte, to ...string) error 

// Send sends an email to the recipient(s)
// the body can be in HTML format as well.
Send(subject string, body string, to ...string) error

// SendWithReader same as `Send` but it accepts
// an io.Reader that body can be retrieved and call the `SendWithBytes`.
SendWithReader(subject string, bodyReader io.Reader, to ...string) error

// SendWithReadCloser same as `SendWithReader` but it closes the reader at the end.
SendWithReadCloser(subject string, bodyReader io.ReadCloser, to ...string) error
```

### Configuration

```go
// Config contains those necessary fields that Mailer needs to send e-mails.
type Config struct {
    // Host is the server mail host, IP or address.
    Host string
    // Port is the listening port.
    Port int
    // Username is the auth username@domain.com for the sender.
    Username string
    // Password is the auth password for the sender.
    Password string
    // FromAddr is the 'from' part of the mail header, it overrides the username.
    FromAddr string
    // FromAlias is the from part, if empty this is the first part before @ from the Username field.
    FromAlias string
    // UseCommand enable it if you want to send e-mail with the mail command  instead of smtp.
    //
    // Host,Port & Password will be ignored.
    // ONLY FOR UNIX.
    UseCommand bool
}
```

### Example

```sh
$ cat example.go
```

```go
package main

import "github.com/kataras/go-mailer"

func main() {
    // sender configuration.
    config := mailer.Config{
        Host:     "smtp.mailgun.org",
        Username: "postmaster",
        Password: "38304272b8ee5c176d5961dc155b2417",
        FromAddr: "postmaster@sandbox661c307650f04e909150b37c0f3b2f09.mailgun.org",
        Port:     587,
        // Enable UseCommand to support sendmail unix command,
        // if this field is true then Host, Username, Password and Port are not required,
        // because these info already exists in your local sendmail configuration.
        //
        // Defaults to false.
        UseCommand: false,
    }

    // initalize a new mail sender service.
    sender := mailer.New(config)

    // the subject/title of the e-mail.
    subject := "Hello subject"

    // the rich message body.
    content := `<h1>Hello</h1> <br/><br/> <span style="color:red"> This is the rich message body </span>`

    // the recipient(s).
    to := []string{"kataras2006@hotmail.com", "kataras2018@hotmail.com"}

    // send the e-mail.
    err := sender.Send(subject, content, to...)

    if err != nil {
        println("error while sending the e-mail: " + err.Error())
    }
}
```

```sh
$ go run example.go
```

## FAQ

Explore [these questions](https://github.com/kataras/go-mailer/issues?go-mailer=label%3Aquestion) or navigate to the [community chat](https://kataras.rocket.chat/channel/go-mailer).

## Versioning

Current: **v0.1.0**

Read more about Semantic Versioning 2.0.0

 - http://semver.org/
 - https://en.wikipedia.org/wiki/Software_versioning
 - https://wiki.debian.org/UpstreamGuide#Releases_and_Versions

### Upgrading from version 0.0.3 to 0.1.0

One breaking change:

The `Send` commands accept a `to ...string` instead of `to []string` now, this is an API Change, if you got multiple `to` emails then just append three dots at the end `...` and you'll be fine, i.e

```go
sender := mailer.New(mailer.Config{...})
to := []string{"recipient1@example.com", "recipient2@example.com"}
sender.Send("subject", "<p>body</p>", to...)
```

## People

The author of go-mailer is [@kataras](https://github.com/kataras).

## Contributing

If you are interested in contributing to the go-mailer project, please make a PR.

### TODO

- [ ] Add a simple CLI tool for sending emails

## License

This project is licensed under the MIT License. License file can be found [here](LICENSE).
