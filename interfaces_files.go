	GetBackend() ApiFileBackend
	GetBackendName() string
	GetBackendID() string
	SetBackendID(string) error

	GetBucket() string
	GetName() string
	GetExtension() string
	GetFullName() string
	GetTitle() string
	GetDescription() string
	// File size in bytes if available.
	GetSize() int64
	GetMime() string
	GetIsImage() bool
	GetWidth() int
	GetHeight() int
	Writer(create bool) (string, *bufio.Writer, ApiError)
	// Lists the buckets that currently exist.
	Writer(f ApiFile, create bool) (string, *bufio.Writer, ApiError)
	WriterById(bucket, id string, create bool) (string, *bufio.Writer, ApiError)
	// Given a file instance with a specified bucket, read the file from filePath, upload it
	// to the backend and then store it in the database.
	// If no file.GetBackendName() is empty, the default backend will be used.
	// The file will be deleted if everything succeeds. Otherwise,
	// it will be left in the file system.
	// If deleteDir is true, the directory holding the file will be deleted
	// also.
	BuildFile(file ApiFile, user ApiUser, filePath string, deleteDir bool) ApiError
