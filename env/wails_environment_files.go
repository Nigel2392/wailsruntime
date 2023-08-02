//go:build !js && !wasm
// +build !js,!wasm

package env

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (f *Environment) OpenFile(constraint *FileConstraint) WailsResponse[File] {
	var (
		p       string
		filters []runtime.FileFilter
		file    File
		err     error
	)
	if constraint != nil && len(constraint.AllowedExtensions) > 0 {
		filters = make([]runtime.FileFilter, len(constraint.AllowedExtensions))
		for i, filter := range constraint.AllowedExtensions {
			filters[i] = runtime.FileFilter{
				DisplayName: filter.DisplayName,
				Pattern:     strings.Join(filter.Extensions, ";"),
			}
		}
	} else {
		filters = make([]runtime.FileFilter, 0)
	}
	if constraint != nil && constraint.Path != "" {
		p = constraint.Path
	} else if constraint != nil && constraint.OpenDirectory {
		p, err = runtime.OpenDirectoryDialog(f.ctx, runtime.OpenDialogOptions{
			Title:   "Select Directory",
			Filters: filters,
		})
		file.IsDir = true
	} else {
		p, err = runtime.OpenFileDialog(f.ctx, runtime.OpenDialogOptions{
			Title:   "Select File",
			Filters: filters,
		})
	}
	if err != nil {
		return WailsResponse[File]{
			Data: file,
			Err:  err.Error(),
		}
	}

	p = filepath.ToSlash(p)

	file.Name = path.Base(p)
	file.Path = p

	if !file.IsDir {
		if constraint != nil {
			err = readFile(constraint.BaseConstraint, &file)
		} else {
			err = readFile(nil, &file)
		}
		if err != nil {
			return WailsResponse[File]{
				Err: err.Error(),
			}
		}
	}

	return WailsResponse[File]{
		Data: file,
		OK:   true,
	}
}

func (f *Environment) OpenMultipleFiles(constraint *MultipleFileConstraint) WailsResponse[[]File] {
	var (
		paths   []string
		filters []runtime.FileFilter
		files   []File
		err     error
	)
	if constraint != nil && len(constraint.AllowedExtensions) > 0 {
		filters = make([]runtime.FileFilter, len(constraint.AllowedExtensions))
		for i, filter := range constraint.AllowedExtensions {
			filters[i] = runtime.FileFilter{
				DisplayName: filter.DisplayName,
				Pattern:     strings.Join(filter.Extensions, ";"),
			}
		}
	}
	paths, err = runtime.OpenMultipleFilesDialog(f.ctx, runtime.OpenDialogOptions{
		Title:   "Select Files",
		Filters: filters,
	})
	if err != nil {
		return WailsResponse[[]File]{
			Err: err.Error(),
		}
	}

	files = make([]File, len(paths))
	for i, p := range paths {
		p = filepath.ToSlash(p)
		var f = &File{
			Name: path.Base(p),
			Path: p,
		}
		err = readFile(constraint.BaseConstraint, f)
		if err != nil {
			return WailsResponse[[]File]{
				Err: err.Error(),
			}
		}
		files[i] = *f
	}

	return WailsResponse[[]File]{
		Data: files,
		OK:   true,
	}
}

func readFile(constraint *BaseConstraint, file *File) error {
	var f, err = os.Open(file.Path)
	if err != nil {
		return err
	}
	defer f.Close()

	var stat os.FileInfo
	stat, err = f.Stat()
	if err != nil {
		return err
	}

	if constraint != nil {
		if stat.Size() > constraint.MaxSize {
			return errors.New("File too large")
		}
		if stat.Size() < constraint.MinSize {
			return errors.New("File too small")
		}
	}

	file.Size = stat.Size()
	file.Extension = path.Ext(file.Path)
	var data = make([]byte, file.Size)
	_, err = f.Read(data)
	if err != nil {
		return err
	}
	file.Data = string(data)
	return nil
}

func (f *Environment) SaveFile(file File, flags FileFlags) WailsResponse[bool] {
	var err error
	var p string
	if file.Path == "" {
		p, err = runtime.SaveFileDialog(f.ctx, runtime.SaveDialogOptions{
			Title:                "Save File",
			ShowHiddenFiles:      true,
			CanCreateDirectories: true,
			DefaultFilename:      file.Name,
		})
	} else {
		p = file.Path
	}
	if err != nil {
		runtime.LogError(f.ctx, err.Error())
		return WailsResponse[bool]{
			Err: err.Error(),
			OK:  false,
		}
	}
	var osFile *os.File

	osFile, err = os.OpenFile(p, convertFlags(flags), 0644)
	if err != nil {
		runtime.LogError(f.ctx, err.Error())
		return WailsResponse[bool]{
			Err: err.Error(),
			OK:  false,
		}
	}
	defer osFile.Close()
	_, err = osFile.Write([]byte(file.Data))
	if err != nil {
		runtime.LogError(f.ctx, err.Error())
		return WailsResponse[bool]{
			Err: err.Error(),
			OK:  false,
		}
	}
	return WailsResponse[bool]{
		Data: true,
		OK:   true,
	}
}
