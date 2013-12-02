# just for developer's convenience 
# the folder structure follows standard GOPATH

.PHONY: clean all fmt run

all:
	@ GOPATH=`pwd` go install -v ./src/...

fmt:
	@ GOPATH=`pwd` go fmt ./src/... 

test:
	GOPATH=`pwd` go test -v ./src/...

clean:
	-rm -r bin pkg

run: all
	bin/drill

doc: all
	@ GOPATH=`pwd` godoc -http=:8000

vet:
	@ GOPATH=`pwd` go vet ./src/...

tags:
	@ gotags `find src -name *.go` > tags


# .PHONY: test
