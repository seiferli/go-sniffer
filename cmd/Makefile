.PHONY: all build run gotool install clean help

BINARY_NAME?=go-sniffer
BIN_DIR=../bin/

VERSION?=v1.0.1
#buildTime=`date -u '+%Y-%m-%d_%H:%M:%S'`

DLVPORT?=:10499
TAGS?="etcdv3"
BUOS?=linux
BUArch?=amd64
flags=" -w -s \
 -X main.SvcName=${BINARY_NAME} \
 -X main.Version=$(VERSION) \
 -X main.Commit=`git rev-parse --short HEAD` \
 -X main.Branch=`git rev-parse --abbrev-ref HEAD` \
 -X main.BuildFlag=$(TAGS)-$(BUOS)-$(BUArch) \
 -X main.BuildTime=`date -u '+%Y-%m-%d_%H:%M:%S_%A'` \
 -X main.OsArch=`go env GOOS`/`go env GOARCH` \
 -X main.GoVersion=`go env GOVERSION`"

all: gotool build

# make buildtag TAGS=kubeapi
build:
	CGO_ENABLED=1 GOOS=${BUOS} GOARCH=$(BUArch) go build -tags=${TAGS} -ldflags=${flags} -o ${BINARY_NAME} main.go

debug:
	GOOS=${BUOS} GOARCH=$(BUArch) go build -tags=${TAGS} -gcflags="all=-N -l" -o ${BINARY_NAME}_dlv main.go

dlv:
	LD_LIBRARY_PATH=/usr/local/lib/ dlv --listen=$(DLVPORT) --headless=true --api-version=2 --accept-multiclient --log exec ./${BINARY_NAME}_dlv -- enp0s3 mongodb -p 30215

run:
	@go run -tags=${TAGS}  ./

test:
	GOMAXPROCS=4 go test -timeout 90s -race ../...

gotool:
	go fmt ../...
	go vet -tags=${TAGS} ../...

install:
	make build
	mkdir -p ${BIN_DIR}
	mv ${BINARY_NAME} ${BIN_DIR}

clean:
	@if [ -f ${BINARY_NAME} ] ; then rm ${BINARY_NAME} ; fi

