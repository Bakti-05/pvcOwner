FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pvc-usage-api .

FROM alpine:latest  

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/pvc-usage-api .

RUN mkdir -p /root/.kube && touch /root/.kube/config && chmod 600 /root/.kube/config

EXPOSE 8080

CMD ["./pvc-usage-api"]