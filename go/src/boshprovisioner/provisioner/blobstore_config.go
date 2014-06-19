package provisioner

import (
	bosherr "bosh/errors"
)

const (
	BlobstoreConfigTypeLocal = "local"
)

type BlobstoreConfig struct {
	Type    string                 `json:"provider"`
	Options map[string]interface{} `json:"options"`
}

func (c BlobstoreConfig) Validate() error {
	// Only local blobstore provides options[blobstore_path]
	if c.Type == BlobstoreConfigTypeLocal {
		_, err := c.extractLocalPath()
		if err != nil {
			return err
		}
	}

	return nil
}

func (c BlobstoreConfig) LocalPath() string {
	if c.Type != BlobstoreConfigTypeLocal {
		return ""
	}

	path, err := c.extractLocalPath()
	if err != nil {
		return ""
	}

	return path
}

func (c BlobstoreConfig) extractLocalPath() (string, error) {
	path, ok := c.Options["blobstore_path"]
	if !ok {
		return "", bosherr.New("Missing blobstore_path in options")
	}

	pathStr, ok := path.(string)
	if !ok {
		return "", bosherr.New("Must provide blobstore_path as a string")
	}

	if pathStr == "" {
		return "", bosherr.New("Must provide non-empty blobstore_path in options")
	}

	return pathStr, nil
}

// AsMap is used to populate agent infrastructure configuration
func (c BlobstoreConfig) AsMap() map[string]interface{} {
	return map[string]interface{}{
		"provider": c.Type,
		"options":  c.Options,
	}
}
