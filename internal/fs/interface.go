package fs

type DirectoryPath string

type MkDirFn = func(path DirectoryPath) error

type CreateSmallTextFileFn = func(filePath DirectoryPath, fileName string, fileContent string) error

type DirectoryExistsCheckFn func(DirectoryPath) (bool, error)
