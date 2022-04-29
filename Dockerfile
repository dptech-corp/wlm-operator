FROM golang:alpine as build-env

WORKDIR /root/wlm-operator
COPY . .
ENV PATH=$PATH:/usr/local/go/bin
ENV HOME=/root

RUN apk update
RUN apk add git && apk add ca-certificates

####################################################################################################

FROM build-env as operator

RUN CGO_ENABLED=0 go build -mod vendor -ldflags "-X main.version=`(git describe  --dirty --always 2>/dev/null || echo "unknown") | sed -e "s/^v//;s/-/_/g;s/_/-/;s/_/./g"`" -o main cmd/operator/main.go

ENTRYPOINT [ "./main" ]

####################################################################################################

FROM build-env as configurator

RUN CGO_ENABLED=0 go build -mod vendor -ldflags "-X main.version=`(git describe  --dirty --always 2>/dev/null || echo "unknown") | sed -e "s/^v//;s/-/_/g;s/_/-/;s/_/./g"`" -o main cmd/configurator/main.go

ENTRYPOINT [ "./main" ]

####################################################################################################

FROM build-env as results

RUN CGO_ENABLED=0 go build -mod vendor -ldflags "-X main.version=`(git describe  --dirty --always 2>/dev/null || echo "unknown") | sed -e "s/^v//;s/-/_/g;s/_/-/;s/_/./g"`" -o main cmd/results/main.go

ENTRYPOINT [ "./main" ]
