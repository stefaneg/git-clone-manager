package fs

type FakeFileSystem struct {
	CreatedFiles []struct {
		FilePath    DirectoryPath
		FileName    string
		FileContent string
	}
	CreatedDirs  []DirectoryPath
	ExistingDirs map[DirectoryPath]bool
}

func (ffc *FakeFileSystem) CreateSmallTextFile(filePath DirectoryPath, fileName string, fileContent string) error {
	ffc.CreatedFiles = append(
		ffc.CreatedFiles, struct {
			FilePath    DirectoryPath
			FileName    string
			FileContent string
		}{filePath, fileName, fileContent},
	)
	return nil
}

func (ffc *FakeFileSystem) MkDir(dir DirectoryPath) error {
	ffc.CreatedDirs = append(ffc.CreatedDirs, dir)
	return nil
}

func (ffc *FakeFileSystem) DirectoryExists(dir DirectoryPath) (bool, error) {
	exists, ok := ffc.ExistingDirs[dir]
	if !ok {
		return false, nil
	}
	return exists, nil
}
