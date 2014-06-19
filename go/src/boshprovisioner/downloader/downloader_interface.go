package downloader

type Downloader interface {
	Download(string) (string, error)
	CleanUp(string) error
}
