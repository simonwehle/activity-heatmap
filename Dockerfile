FROM golang:alpine AS build
RUN apk --no-cache add git build-base
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=0
RUN go build -o main .

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/main .
COPY --from=build /app/web ./web
COPY --from=build /app/style ./style
RUN mkdir -p /app/activities
EXPOSE 3465
CMD ["./main"]