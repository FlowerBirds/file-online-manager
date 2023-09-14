FROM  golang:1.18 as build
ENV GO111MODULE=on
ENV CGO_ENABLED=0
WORKDIR /app

COPY vendor vendor
COPY handler handler
COPY model model
COPY util util
COPY go.mod go.mod
COPY main.go main.go

RUN go build -o manage

FROM alpine:3.17 as ship
RUN apk update && \
    apk add zip --no-cache
WORKDIR /home/app
COPY --from=build /app/manage .
COPY static static

EXPOSE 8080
ENV CONTEXT_PATH /
ENV ROOT_PATH /app
RUN mkdir -p ${ROOT_PATH}/apps && mkdir -p ${ROOT_PATH}/file

CMD ["./manage"]