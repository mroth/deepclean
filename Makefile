.PHONY: bench snapshot clean

bin/deepclean: *.go ./cmd/deepclean/*.go
	go build -o $@ ./cmd/deepclean

bench: bin/deepclean
	hyperfine -w1 "$< ~/src"

snapshot:
	goreleaser --snapshot --rm-dist

clean:
	rm -rf bin
	rm -rf dist

# release process:
# $ git tag -a v0.1.0 -m "First release"
# $ git push origin v0.1.0
# $ goreleaser
#
# default token location if not in env: ~/.config/goreleaser/github_token
