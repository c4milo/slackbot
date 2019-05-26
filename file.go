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

type File struct {
	Content  string      `yaml:"content"`
	Mode     os.FileMode `yaml:"mode"`
	Owner    string      `yaml:"owner"`
	Group    string      `yaml:"group"`
	DestPath string      `yaml:"dest"`
	task     `yaml:",inline"`
	checksum []byte
}

func (f *File) Init() error {
	return nil
}

func (f *File) Validate() error {
	if f.DestPath == "" {
		return errors.New("file: destination path required")
	}
	return nil
}

func (f *File) Apply() ([]byte, error) {
	currentFile := f.state()
	var err error

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
