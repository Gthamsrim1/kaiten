package errs

import (
	"errors"
	"syscall"
)

var (
	// Generic
	ErrInvalidName = errors.New("invalid name")
	ErrEmptyName   = errors.New("empty name")
	ErrNameTooLong = errors.New("name too long")
	ErrInvalidPath = errors.New("invalid path")
	ErrNotEmpty    = errors.New("not empty")

	// Lookup
	ErrNotFound      = errors.New("node not found")
	ErrAlreadyExists = errors.New("node already exists")

	// Directory
	ErrNotDirectory      = errors.New("not a directory")
	ErrDirectoryNotEmpty = errors.New("directory not empty")

	// File
	ErrNotFile            = errors.New("not a file")
	ErrReadOnlyFileSystem = errors.New("read only filesystem")
	ErrPermissionDenied   = errors.New("permission denied")

	// Parent
	ErrNilParent     = errors.New("parent directory is nil")
	ErrInvalidParent = errors.New("invalid parent")

	// Storage
	ErrObjectMissing   = errors.New("backing object missing")
	ErrObjectCorrupted = errors.New("backing object corrupted")
	ErrRefUnderflow    = errors.New("refcount underflow")

	//Snapshot
	ErrSnapshotIDEmpty = errors.New("snapshot id is empty")

	// Runtime
	ErrInvalidInode     = errors.New("invalid inode")
	ErrInvalidOperation = errors.New("invalid operation")
)

func ToErrno(err error) syscall.Errno {
	switch {
	case err == nil:
		return 0

	case errors.Is(err, ErrEmptyName):
		return syscall.EINVAL

	case errors.Is(err, ErrInvalidName):
		return syscall.EINVAL

	case errors.Is(err, ErrNameTooLong):
		return syscall.ENAMETOOLONG

	case errors.Is(err, ErrInvalidPath):
		return syscall.ENOENT

	case errors.Is(err, ErrAlreadyExists):
		return syscall.EEXIST

	case errors.Is(err, ErrNotFound):
		return syscall.ENOENT

	case errors.Is(err, ErrNotDirectory):
		return syscall.ENOTDIR

	case errors.Is(err, ErrNotFile):
		return syscall.EISDIR

	case errors.Is(err, ErrDirectoryNotEmpty):
		return syscall.ENOTEMPTY

	case errors.Is(err, ErrPermissionDenied):
		return syscall.EACCES

	case errors.Is(err, ErrReadOnlyFileSystem):
		return syscall.EROFS

	case errors.Is(err, ErrNilParent):
		return syscall.EFAULT

	case errors.Is(err, ErrInvalidParent):
		return syscall.EINVAL

	case errors.Is(err, ErrObjectMissing):
		return syscall.EIO

	case errors.Is(err, ErrObjectCorrupted):
		return syscall.EIO

	case errors.Is(err, ErrInvalidInode):
		return syscall.EBADF

	case errors.Is(err, ErrInvalidOperation):
		return syscall.ENOSYS

	case errors.Is(err, ErrNotEmpty):
		return syscall.ENOTEMPTY

	case errors.Is(err, ErrRefUnderflow):
		return syscall.EIO

	default:
		return syscall.EIO
	}
}
