package adapters

import "log"

type EmailClient interface {
	SendEmail(from, to string, title string, message string) error
}

// InMemoryEmailClient used for demo purposes
type InMemoryEmailClient struct{}

func (c *InMemoryEmailClient) SendEmail(from, to string, title string, message string) error {
	log.Printf("Sent email from %q to %q with title %q and message: %q\n", from, to, title, message)
	return nil
}

var _ EmailClient = (*InMemoryEmailClient)(nil)
