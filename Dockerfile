FROM golang:1.24-alpine AS build
WORKDIR /test
COPY . .

RUN go build main.go
EXPOSE 8080

FROM scratch
WORKDIR /test
COPY --from=build /test/main .
ENTRYPOINT [ "./main" ]