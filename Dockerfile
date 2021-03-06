FROM golang:1.4-cross

RUN mkdir -p /go/src/github.com/docker \
	&& git clone https://github.com/docker/libcontainer.git /go/src/github.com/docker/libcontainer \
	&& cd /go/src/github.com/docker/libcontainer \
	&& git checkout --quiet b322073f27b0e9e60b2ab07eff7f4e96a24cb3f9

ENV GOPATH $GOPATH:/go/src/github.com/docker/libcontainer/vendor

# disable CGO for ALL THE THINGS (to help ensure no libc)
ENV CGO_ENABLED 0

COPY *.go /go/src/github.com/bpowers/gosu/
WORKDIR /go/src/github.com/bpowers/gosu

# gosu-$(dpkg --print-architecture)
RUN GOARCH=amd64       go build -v -ldflags -d -o /go/bin/gosu-amd64
RUN GOARCH=386         go build -v -ldflags -d -o /go/bin/gosu-i386
RUN GOARCH=arm GOARM=5 go build -v -ldflags -d -o /go/bin/gosu-armel
RUN GOARCH=arm GOARM=6 go build -v -ldflags -d -o /go/bin/gosu-armhf
#RUN GOARCH=arm GOARM=7 go build -v -ldflags -d -o /go/bin/gosu-armhf # boo Raspberry Pi, making life hard
