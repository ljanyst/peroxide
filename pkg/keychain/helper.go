// Copied from github.com/docker/docker-credential-helpers to aviod dependency
// on cgo. MIT License. Copyright (c) 2016 David Calavera

package keychain

type Credentials struct {
	ServerURL string
	Username  string
	Secret    string
}

type Helper interface {
	// Add appends credentials to the store.
	Add(*Credentials) error
	// Delete removes credentials from the store.
	Delete(serverURL string) error
	// Get retrieves credentials from the store.
	// It returns username and secret as strings.
	Get(serverURL string) (string, string, error)
	// List returns the stored serverURLs and their associated usernames.
	List() (map[string]string, error)
}
