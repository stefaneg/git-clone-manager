# Git Clone Manager

## Scratching personal itches which may scratch someone else's itch as well.

## Getting Started

### Prerequisites

- Go 1.23.2 or later
- Git
- A GitLab API token with access to the projects you want to clone

### Installation

1. Clone the repository:
    ```sh
    git clone https://github.com/stefaneg/git-clone-manager.git
    cd git-clone-manager
    ```

2. Install dependencies:
    ```sh
    go mod tidy
    ```

### Configuration

1. Create a `workingCopies.yaml` file in your home directory with the following structure:
    ```yaml
    gitlab:
      - tokenEnvVar: "GITLAB_API_TOKEN"
        hostName: 'gitlab.example.com'
        cloneDirectory: '/path/to/clone/directory'
        groups:
          - id: "group-id-1"
            cloneArchived: false
          - id: "group-id-2"
            cloneArchived: true
        projects:
          - name: "Project Name"
            fullPath: "group/project"
    ```

Note that you also need to be authenticated in git with permissions to clone projects with an ssh key.

2. Set the environment variable for your GitLab API token:
    ```sh
    export GITLAB_API_TOKEN=your_token_here
    ```

### Compilation

Compile the git clone manager using the Go compiler:
```shell
go build -o gcm
```

### Unit test

```shell
go test ./...
```

# Use

Currently, the only command is "clone". Running ```gcm``` will clone all groups and projects specified in your 
configuration file.


# To do
- Collect statistics - how many projects processed - checked out - archived
- Support GitHub api to clone organisations.
- Potentially destructive commands are always local in scope by default. --global option to make global in scope.
- 
- Create command to add to management "gcm clone <URL>". Adds a repo/group/organisation to clone management, and clones. 
- Create command to delete branches without remote. "gcm cleanup" && "gcm --global cleanup"
- Create command that reports projects that are not on main branch
- Create command that reports projects with dirty index
- Create command that reports projects with un-pushed changes.
- Create command to report all projects that  
  - a) have uncommitted changes 
  - b) are behind origin or without a tracked remote branch 
  - c) are checked out on a branch.
- Create command to pull changes on projects on main and with a clean index.
- Create command to open webUI for repo "gcm webui". Opens Gitlab or Github on page of repo.

- Log commands issued on each repository in separate files. Where exactly is a bit tricky...in checkout root directory ?
