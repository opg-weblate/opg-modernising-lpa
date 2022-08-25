FROM node:18.8.0-alpine3.16 as asset-env

WORKDIR /app

RUN mkdir -p web/static

COPY package.json .
COPY yarn.lock .
RUN yarn

COPY app/web/assets web/assets
RUN yarn build

FROM golang:1.18 as build-env

WORKDIR /app

COPY app/go.mod .
COPY app/go.sum .

RUN go mod download

COPY /app .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/mlpab

FROM alpine:3.16.1 as production

WORKDIR /go/bin

COPY --from=build-env /go/bin/mlpab mlpab
COPY --from=asset-env /app/web/static web/static
COPY app/web/template web/template
COPY app/lang lang

RUN addgroup -S app && \
    adduser -S -g app app && \
    chown -R app:app mlpab web/template web/static
USER app

ENTRYPOINT ["./mlpab"]
