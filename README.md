# Github Cat

Cat a file path across all repos in github org.

## Usage

```
github-cat ORG PATH
    -n, --no-repo-name               Do not prefix lines with repo name
        --json                       JSON output
    -h, --help                       Prints this help
```

## Configuration

Reads your Github API token from the environment variable `GITHUB_API_TOKEN`

## Examples

### Defaults

    $ github-cat someorg .ruby-version
    someorg/app-foo:2.6.3
    someorg/app-bar:2.4.6
    someorg/app-baz:2.5.5

### With Multiline files

    $ github-cat someorg Dockerfile
    someorg/app-foo:FROM ruby:2.6.3
    someorg/app-foo:COPY . /app/
    someorg/app-foo:CMD ["foo"]
    someorg/app-bar:FROM ruby:2.4.6
    someorg/app-bar:RUN apt-get update && apt-get install curl
    someorg/app-bar:COPY . /app/
    someorg/app-bar:CMD ["bar"]

### No repo prefix

    $ github-cat --no-repo-name someorg .ruby-version
    2.6.3
    2.4.6
    2.5.5

### JSON Output

    $ github-cat --json someorg .ruby-version | jq .
    [
      {
        "repo": "someorg/app-foo",
        "content": "2.6.3\n"
      },
      {
        "repo": "someorg/app-bar",
        "content": "2.4.6\n"
      },
      {
        "repo": "someorg/app-baz",
        "content": "2.5.5\n"
      }
    ]

## Run with Docker

    $ docker build -t github-cat .
    $ export GITHUB_API_TOKEN=setecastronomy
    $ docker run --rm -e GITHUB_API_TOKEN github-cat someorg .ruby-version
