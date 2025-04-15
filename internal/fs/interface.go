package fs

type Path string
type DirectoryPath Path
type FilePath Path

type FileName string

type MkDirFn = func(path DirectoryPath) error

type CreateSmallTextFileFn = func(filePath DirectoryPath, fileName FileName, fileContent string) error

//type DirectoryExistsCheckFn func(DirectoryPath) (bool, error)

type FileSystem interface {
	MkDir(DirectoryPath) error
	DirectoryExists(DirectoryPath) (bool, error)
	CreateSmallTextFile(filePath DirectoryPath, fileName FileName, fileContent string) error
}
