MINGW_DIR=/usr/x86_64-w64-mingw32
WINDOWS_CC=x86_64-w64-mingw32-gcc
WINDOWS_CCX=x86_64-w64-mingw32-g++
WINDOWS_VARS=GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=${WINDOWS_CC} CCX=${WINDOWS_CCX} CGO_CFLAGS="-I${MINGW_DIR}/include/" CGO_LDFLAGS="-lusb-1.0"

FILES = $(wildcard *.go)

all: $(FILES) go.mod go.sum windows native

native:
	go build

windows: 
	${WINDOWS_VARS} go build -ldflags "-linkmode external -extldflags -static"