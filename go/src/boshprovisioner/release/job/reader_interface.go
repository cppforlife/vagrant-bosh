package job

type Reader interface {
	Read() (Job, error)
	Close() error
}
