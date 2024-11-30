package gitremote

// GitRemoteProjectConfig Configuration that points directly to a remote project. Not dependent on type of RepoManager, but needs HostName from RepoManager for full config.
type GitRemoteProjectConfig struct {
	Name     string `yaml:"name"`
	FullPath string `yaml:"fullPath"`
}
