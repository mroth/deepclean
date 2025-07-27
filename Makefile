.PHONY: bench snapshot clean

bin/deepclean: */**/*.go go.mod go.sum
	go build -o $@ .

bench: bin/deepclean
	hyperfine -w1 "$< ~/src"

snapshot:
	goreleaser --snapshot --clean

clean:
	rm -rf bin
	rm -rf dist
