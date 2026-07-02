# Github Cat

Cat a file path across all repos owned by a github org or user.

## Install

    $ go install github.com/danajp/github-cat/cmd/github-cat@latest

This installs the `github-cat` binary into `$(go env GOPATH)/bin` (make sure that directory is on your `PATH`).

## Usage

```
github-cat OWNER PATH [flags]

Flags:
  -j, --concurrency int    Number of repos to fetch in parallel (default 20)
  -h, --help               help for github-cat
      --include-archived   Include archived repos (off by default)
      --include-forks      Include forked repos (off by default)
      --json               JSON output
  -n, --no-repo-name       Do not prefix lines with repo name
  -r, --regex string       Filter lines in file by regex
```

`OWNER` may be either an organization or a user.
For a user, only their public repositories are searched.

Files larger than 1 MB cannot be fetched through the GitHub contents API, so `github-cat` will report an error if `PATH` matches a file over that size.

## Configuration

Reads your Github API token from the environment variable `GITHUB_TOKEN`.

### Token scopes

`github-cat` lists an organization's repositories and reads file contents, so the token needs read access to those repositories.

- **Fine-grained personal access token:** grant the token access to the target organization's repositories with the **Contents: Read-only** repository permission. (**Metadata: Read-only** is required by GitHub and is granted automatically.)
- **Classic personal access token:** use the `repo` scope to read private repositories. For public repositories only, `public_repo` (or no scope at all) is sufficient.

If you have the [GitHub CLI](https://cli.github.com/) installed, you can use its token directly:

    $ GITHUB_TOKEN=$(gh auth token) github-cat someorg go.mod

## Build

    $ go build -o github-cat ./cmd/github-cat

## Examples

### Defaults

    $ github-cat someorg go.mod
    someorg/app-foo:module github.com/someorg/app-foo
    someorg/app-foo:
    someorg/app-foo:go 1.21
    someorg/app-bar:module github.com/someorg/app-bar
    someorg/app-bar:
    someorg/app-bar:go 1.22
    someorg/app-baz:module github.com/someorg/app-baz
    someorg/app-baz:
    someorg/app-baz:go 1.20

### Extract the go version with a regex

    $ github-cat --regex '^go ' someorg go.mod
    someorg/app-foo:go 1.21
    someorg/app-bar:go 1.22
    someorg/app-baz:go 1.20

### No repo prefix

    $ github-cat --no-repo-name --regex '^go ' someorg go.mod
    go 1.21
    go 1.22
    go 1.20

### JSON Output

    $ github-cat --json --regex '^go ' someorg go.mod | jq .
    [
      {
        "repo": "someorg/app-foo",
        "content": "go 1.21"
      },
      {
        "repo": "someorg/app-bar",
        "content": "go 1.22"
      },
      {
        "repo": "someorg/app-baz",
        "content": "go 1.20"
      }
    ]

## Run with Docker

    $ docker build -t github-cat .
    $ export GITHUB_TOKEN=setecastronomy
    $ docker run --rm -e GITHUB_TOKEN github-cat someorg go.mod
