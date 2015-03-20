package nameexception

type Provider interface {
	Name() string
	URL() string
	GetExceptions() ([]byte, error)
	ProcessExceptions([]byte) error
}
