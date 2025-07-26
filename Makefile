.PHONY: bench snapshot clean

bin/deepclean: *.go ./cmd/deepclean/*.go
	go build -o $@ ./cmd/deepclean

bench: bin/deepclean
	hyperfine -w1 "$< ~/src"

snapshot:
	goreleaser --snapshot --clean

clean:
	rm -rf bin
	rm -rf dist
