# just for developer's convenience 
# the folder structure follows standard GOPATH

.PHONY: clean all fmt run

all:
	@ GOPATH=`pwd` go install -v ./src/...

fmt:
	@ GOPATH=`pwd` go fmt ./src/... 

clean:
	-rm -r bin pkg

run: all
	bin/run

doc: all
	godoc -http=:6060 -path=.


# .PHONY: test
