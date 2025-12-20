package database

// Config defines configuration data for the databases.
type Config struct {
	// Users are the database users.
	Users []string `yaml:"users,omitempty"`
	// Database contains database connection details.
	Database map[string]string `yaml:"database,omitempty"`
}
