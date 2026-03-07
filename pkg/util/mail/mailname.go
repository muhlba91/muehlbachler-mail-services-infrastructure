package mail

import "fmt"

// Mailname constructs a mail server name for the given domain.
// domain: The domain for which to create the mail server name.
func Mailname(domain string) string {
	return fmt.Sprintf("mail.%s", domain)
}
