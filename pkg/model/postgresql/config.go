package postgresql

// Config holds the PostgreSQL connection configuration.
type Config struct {
	// Address is the PostgreSQL server address.
	Address string
	// Port is the PostgreSQL server port.
	Port int
	// Username is the username for PostgreSQL authentication.
	Username string
	// Password is the password for PostgreSQL authentication.
	//nolint:gosec // This is not a hardcoded secret, it's a configuration field.
	Password string
}
