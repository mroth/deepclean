.PHONY: bench

bin/deepclean: *.go
	go build -o $@

bench: bin/deepclean
	hyperfine -w1 "$< ~/src"
