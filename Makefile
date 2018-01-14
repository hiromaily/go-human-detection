CURDIR := $(PWD)

ins:
	go get -u -d gocv.io/x/gocv

ins-linux:
	cd $(GOPATH)/src/gocv.io/x/gocv
	make deps
	#
	make download
	make build
	make clean

# it doesn't work...so `source` should be executed outside Makefile.
# source $GOPATH/src/gocv.io/x/gocv/env.sh
bld-gocv:
	cd $(GOPATH)/src/gocv.io/x/gocv && \
	chmod 755 ./env.sh && \
	source ./env.sh
	echo $(CGO_LDFLAGS)
	#cd $(CURDIR)

bld:
	go build -race -v -o ${GOPATH}/bin/go-cv ./main.go
	#go build -i -race -v -o ${GOPATH}/bin/go-cv ./main.go

exec:
	go-cv

exec1:
	#face detection
	go-cv -mode 1 -gh 'https://xxxxx.ngrok.io/google-home-notifier'

exec2:
	#motion detection
	go-cv -mode 2 -gh 'https://xxxxx.ngrok.io/google-home-notifier'

exec3:
	#web stream
	go-cv -mode 3 -port 8080 -gh 'https://xxxxx.ngrok.io/google-home-notifier'

run:
	go run -race ./main.go
