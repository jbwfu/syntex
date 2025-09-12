package packer

// Formatter defines the contract for output formatters.
type Formatter interface {
	Format(filename, language string, content []byte) ([]byte, error)
}
