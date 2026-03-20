package domain

// Commit is an immutable value object representing a single Git commit.
type Commit struct {
	Hash      string // full 40-char SHA
	ShortHash string // first 7 chars
	Subject   string // first line of commit message
	Author    string
	RelDate   string // e.g. "2 days ago"
}
