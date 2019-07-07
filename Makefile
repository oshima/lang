install:
	@rm -f tmp*
	gofmt -w .
	go vet ./...
	golint ./...
	go install

test:
	.test/test.sh
	@rm -f tmp*

clean:
	rm -f tmp*

.PHONY: install test clean
