FROM golang:1.25-alpine3.22 AS builder

WORKDIR /app/image-detector

COPY go.mod go.sum ./

RUN go mod download

COPY . .


RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM scratch

LABEL me.malachowski.source=github.com/KacperMalachowski/image-detector-action

COPY --from=builder /app/image-detector/main /image-detector

ENTRYPOINT [ "/image-detector" ]
