package connection

type DatabaseConnection interface {
	Test() error
	Close() error
}
