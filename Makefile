.PHONY: bench snapshot

bin/deepclean: *.go
	go build -o $@

bench: bin/deepclean
	hyperfine -w1 "$< ~/src"

snapshot:
	goreleaser --snapshot --rm-dist

# release process:
# $ git tag -a v0.1.0 -m "First release"
# $ git push origin v0.1.0
# $ goreleaser
#
# default token location if not in env: ~/.config/goreleaser/github_token
