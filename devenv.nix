{ pkgs, ... }:

{
  languages.go.enable = true;

  packages = [ pkgs.golangci-lint ];

  scripts.run-cli = {
    exec = ''go run ./cmd/github-cat "$@"'';
    description = "Run the github-cat CLI (forwards arguments)";
  };

  scripts.run-build = {
    exec = ''go build -o github-cat ./cmd/github-cat'';
    description = "Build the github-cat binary";
  };

  scripts.run-tests = {
    exec = ''go test ./... "$@"'';
    description = "Run the test suite";
  };

  scripts.run-fmt = {
    exec = ''gofmt -w cmd'';
    description = "Format Go source files";
  };

  scripts.run-lint = {
    exec = ''
      set -e
      unformatted=$(gofmt -l cmd)
      if [ -n "$unformatted" ]; then
        echo "The following files are not gofmt-formatted:"
        echo "$unformatted"
        exit 1
      fi
      go vet ./...
      golangci-lint run ./...
    '';
    description = "Run gofmt check, go vet, and golangci-lint";
  };

  enterShell = ''
    echo "github-cat dev environment"
    echo "  go            $(go version | cut -d' ' -f3)"
    echo "  golangci-lint $(golangci-lint version --short 2>/dev/null || golangci-lint --version)"
    echo ""
    echo "Available scripts:"
    echo "  run-cli    -- go run the CLI (e.g. run-cli someorg go.mod)"
    echo "  run-build  -- build the github-cat binary"
    echo "  run-tests  -- run the test suite"
    echo "  run-fmt    -- format Go source files"
    echo "  run-lint   -- gofmt check, go vet, and golangci-lint"
  '';
}
