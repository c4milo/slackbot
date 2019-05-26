package slackbot

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/user"
	"strconv"
	"syscall"

	"golang.org/x/crypto/blake2b"
)

// File defines a file management module
// Available states are: present or absent
type File struct {
	task `yaml:",inline"`
	// Content declares the content of the file
	Content string `yaml:"content"`
	// Mode declares the permissions to set to the file
	Mode os.FileMode `yaml:"mode"`
	// Owner declares the user who owns the file
	Owner string `yaml:"owner"`
	// Group declares the group who has access to the file
	Group string `yaml:"group"`
	// DestPath is the file's destination path
	DestPath string `yaml:"dest"`
	// checksum is the hash of the file content. Used to determine whether or not
	// the file content needs to be rewritten.
	checksum []byte
}

// Init does nothing in this module
func (f *File) Init() error {
	if f.State == "" {
		f.State = "present"
	}
	return nil
}

// Validate validates whether a destination path was declared
func (f *File) Validate() error {
	if f.DestPath == "" {
		return errors.New("file: destination path required")
	}
	return nil
}

// Apply applies the declared file state if needed.
func (f *File) Apply() ([]byte, error) {
	currentFile := f.state()
	var err error

	if currentFile == nil && f.State == "absent" {
		return nil, nil
	}

	if f.State == "absent" {
		return nil, os.Remove(f.DestPath)
	}

	file := new(os.File)
	if currentFile == nil {
		file, err = os.Create(f.DestPath)
	} else {
		file, err = os.OpenFile(f.DestPath, os.O_RDWR, currentFile.Mode)
	}

	if err != nil {
		return nil, fmt.Errorf("file: %s", err)
	}
	defer file.Close()

	if f.Content != "" {
		checksum := blake2b.Sum512([]byte(f.Content))
		if currentFile == nil || !bytes.Equal(currentFile.checksum, checksum[:]) {
			if err := file.Truncate(0); err != nil {
				return nil, fmt.Errorf("file: %s", err)
			}

			_, err := file.WriteString(f.Content)
			if err != nil {
				return nil, fmt.Errorf("file: %s", err)
			}
			f.changed = true
		}
	}

	if f.Mode != 0 && currentFile.Mode != f.Mode {
		if err := file.Chmod(f.Mode); err != nil {
			return nil, fmt.Errorf("file: %s", err)
		}
		f.changed = true
	}

	if f.Owner != "" && currentFile.Owner != f.Owner {
		usr, err := user.Lookup(f.Owner)
		if err != nil {
			return nil, fmt.Errorf("file: %s", err)
		}

		uid, err := strconv.Atoi(usr.Uid)
		if err != nil {
			return nil, fmt.Errorf("file: %s", err)
		}

		if err := file.Chown(uid, -1); err != nil {
			return nil, fmt.Errorf("file: %s", err)
		}
		fmt.Println("file: owner changed")
		f.changed = true
	}

	if f.Group != "" && currentFile.Group != f.Group {
		group, err := user.LookupGroup(f.Group)
		if err != nil {
			return nil, fmt.Errorf("file: %s", err)
		}

		gid, err := strconv.Atoi(group.Gid)
		if err != nil {
			return nil, fmt.Errorf("file: %s", err)
		}

		if err := file.Chown(-1, gid); err != nil {
			return nil, fmt.Errorf("file: %s", err)
		}
		f.changed = true
	}

	return nil, nil
}

// state retrieves the current metadata state of the file
func (f *File) state() *File {
	meta, err := os.Stat(f.DestPath)
	if os.IsNotExist(err) {
		return nil
	}
	current := &File{
		Mode:     meta.Mode().Perm(),
		DestPath: f.DestPath,
	}

	if f.Owner != "" {
		uid := meta.Sys().(*syscall.Stat_t).Uid
		user, err := user.LookupId(strconv.FormatUint(uint64(uid), 10))
		if err == nil {
			current.Owner = user.Username
		}
	}

	if f.Group != "" {
		gid := meta.Sys().(*syscall.Stat_t).Gid
		group, err := user.LookupGroupId(strconv.FormatUint(uint64(gid), 10))
		if err == nil {
			current.Group = group.Name
		}
	}

	if f.Content == "" {
		return current
	}

	fileReader, err := os.Open(f.DestPath)
	if err != nil {
		return nil
	}
	defer fileReader.Close()

	h, err := blake2b.New512(nil)
	if err != nil {
		return nil
	}

	if _, err := io.Copy(h, fileReader); err != nil {
		return nil
	}

	current.checksum = h.Sum(nil)
	return current
}
