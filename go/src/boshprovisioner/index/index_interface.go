package index

import (
	"errors"
)

var (
	ErrNotFound = errors.New("Record is not found")
)

type Index interface {
	ListKeys(interface{}) error
	List(interface{}) error

	Find(interface{}, interface{}) error
	Save(interface{}, interface{}) error
	Remove(interface{}) error
}
