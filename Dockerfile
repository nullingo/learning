FROM golang:1.17 AS build

WORKDIR /HTTPSERVER/

COPY . .

RUN go env -w GOPROXY=https://goproxy.cn,direct && \
	go env -w GOOS=linux && \
	go env -w CGO_ENABLED=0

RUN go build -installsuffix cgo -o httpserver httpserver.go


FROM busybox

COPY --from=build /HTTPSERVER/httpserver /HTTPSERVER/httpserver 

EXPOSE 8360

ENV ENV local

WORKDIR /HTTPSERVER/

ENTRYPOINT ["./httpserver"]

