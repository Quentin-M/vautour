package mailer

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

// DefaultConfig returns the default configs for Mailer
// returns just an empty Config struct.
func DefaultConfig() Config {
	return Config{}
}

// IsValid returns true if the configuration is valid, otherwise false.
func (m Config) IsValid() bool {
	return (m.Host != "" && m.Port > 0 && m.Username != "" && m.Password != "") || m.UseCommand
}
