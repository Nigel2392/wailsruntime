package env

import "github.com/Nigel2392/jsext/v2/errs"

type FileFlags int

const (
	O_RDONLY FileFlags = 1
	O_WRONLY FileFlags = 2
	O_RDWR   FileFlags = 4
	O_APPEND FileFlags = 8
	O_CREATE FileFlags = 16
	O_EXCL   FileFlags = 32
	O_SYNC   FileFlags = 64
	O_TRUNC  FileFlags = 128
)

type WailsResponse[T any] struct {
	Data T      `json:"data,omitempty"`
	OK   bool   `json:"ok,omitempty"`
	Err  string `json:"error,omitempty"`
}

func (w WailsResponse[T]) AsError() error {
	if w.Err == "" {
		return nil
	}
	return errs.Error(w.Err)
}

type Filter struct {
	DisplayName string   `json:"name"`
	Extensions  []string `json:"extensions"`
}

type BaseConstraint struct {
	MaxSize           int64    `json:"maxSize"`
	MinSize           int64    `json:"minSize"`
	AllowedExtensions []Filter `json:"allowedExtensions"`
	AllowedMimeTypes  []string `json:"allowedMimeTypes"`
}

type FileConstraint struct {
	Path            string `json:"path"`
	OpenDirectory   bool   `json:"openDirectory"`
	*BaseConstraint `json:",inline"`
}

type MultipleFileConstraint struct {
	MaxFiles int `json:"maxFiles"`
	*BaseConstraint
}

type File struct {
	Name      string `json:"name"`
	Extension string `json:"extension"`
	Path      string `json:"path"`
	Data      string `json:"data"`
	Size      int64  `json:"size"`
	IsDir     bool   `json:"isDir"`
}
