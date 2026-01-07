FROM golang:1.24 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/bankdash ./cmd/bankdash

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /out/bankdash /app/bankdash
ENV PORT=8080
EXPOSE 8080
ENTRYPOINT ["/app/bankdash"]
