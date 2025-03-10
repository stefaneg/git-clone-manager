package fs

type DirectoryPath string

type FileName string

type MkDirFn = func(path DirectoryPath) error

type CreateSmallTextFileFn = func(filePath DirectoryPath, fileName FileName, fileContent string) error

type DirectoryExistsCheckFn func(DirectoryPath) (bool, error)
