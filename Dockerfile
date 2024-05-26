FROM golang:1.22.3

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /podprox cmd/podprox/main.go

EXPOSE 3000

# Run
CMD ["/podprox"]