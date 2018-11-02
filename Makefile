install:
	gofmt -w .
	go install

test:
	test/test.sh
	@rm -f tmp*

clean:
	rm -f tmp*

.PHONY: install test clean
