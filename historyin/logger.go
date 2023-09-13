package historyin

// Logger is the logger this package uses
type Logger interface {
	Error(error)
}

type nopLogger struct {
}

func (l nopLogger) Error(err error) {}
