# Gulli's Personal Tools written in Go

## Scratching personal itches which may become tools to scratch someone else's itch as well.

## Getting Started

### Prerequisites

- Go 1.23.2 or later
- Git
- A GitLab API token with access to the projects you want to clone

### Installation

1. Clone the repository:
    ```sh
    git clone https://github.com/yourusername/yourrepository.git
    cd yourrepository
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

Note that you also need to be authenticated in git with permissions to clone projects.

2. Set the environment variable for your GitLab API token:
    ```sh
    export GITLAB_API_TOKEN=your_token_here
    ```

### Compilation

Compile the project using the Go compiler:
```shell
go build -o wcmanager
```


# To do
- Collect error counts, archived counts. 
- Separate into goroutine-free access / repository classes and pipeline classes to collect results.
- Delete branches without remote.
- Create command that reports projects that are not on main branch
- Create command that reports projects with dirty index
- Create command that reports projects with unpushed changes.
- Collect statistics - how many projects processed - checked out - archived
- Report all projects that have a) have uncommitted changes b) are behind origin or without a tracked remote branch c) are checked out on a branch.
- Create command to pull changes on projects on main and with a clean index.
