##
## Build
##
FROM golang:1.17-alpine AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY *.go ./

RUN go build -o ./terraform_http_backend

##
## Deploy
##
FROM alpine

WORKDIR /app

RUN addgroup -S tf; adduser -h /app tf -G tf -D
RUN mkdir /app/store

COPY --from=build /app/terraform_http_backend /app/terraform_http_backend
COPY .env.dist ./.env
RUN chown tf:tf -R /app

EXPOSE 8080

USER tf:tf

ENTRYPOINT [ "/app/terraform_http_backend" ]