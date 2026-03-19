package domain

import "fmt"

// Commit is an immutable value object representing a single Git commit.
type Commit struct {
	Hash      string // full 40-char SHA
	ShortHash string // first 7 chars
	Subject   string // first line of commit message
	Author    string
	RelDate   string // e.g. "2 days ago"
}

// Display returns a single-line string for terminal output.
func (c Commit) Display(index int) string {
	subject := c.Subject
	if len(subject) > 60 {
		subject = subject[:57] + "..."
	}
	return fmt.Sprintf("  %2d)  %s  %-60s  %s  (%s)",
		index, c.ShortHash, subject, c.Author, c.RelDate)
}
