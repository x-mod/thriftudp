PROJECT=echo

clean:
	@go clean
	@rm -rf thrift
gen: $(shell which thrift)
	@mkdir -p thrift
	thrift -out thrift -r --gen go idl/*.thrift 

build: gen	
	go build
