package tar

type Compressor interface {
	Compress(string) (string, error)
	CleanUp(string) error
}
