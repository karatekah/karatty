FROM golang:1.27.3 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /server ./cmd/server && mkdir /state && chmod 700 /state

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /server /server
COPY --from=build --chown=65532:65532 /state /state
USER 65532:65532
WORKDIR /state
ENTRYPOINT ["/server"]
