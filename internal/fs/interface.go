package fs

type Path string
type DirectoryPath Path
type FilePath Path

type FileName string

type MkDirFn = func(path DirectoryPath) error

type CreateSmallTextFileFn = func(filePath DirectoryPath, fileName FileName, fileContent string) error

type DirectoryExistsCheckFn func(DirectoryPath) (bool, error)
