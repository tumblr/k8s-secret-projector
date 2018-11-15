package key

// Key is an interface for a Key that is returned from an Module
type Key interface {
	Plaintext() string
}
