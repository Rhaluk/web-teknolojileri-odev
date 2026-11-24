# Yemekhane uygulaması · iki aşamalı imaj
# 1) golang imajında derlenir  2) yalnızca çalıştırılabilir dosya küçük bir imaja kopyalanır

FROM golang:1.27-alpine AS derleme
WORKDIR /kaynak
# Önce bağımlılıklar: go.mod değişmedikçe bu katman önbellekten gelir
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /uygulama .

FROM alpine:3.22
RUN adduser -D -H uygulama
COPY --from=derleme /uygulama /usr/local/bin/uygulama
USER uygulama
EXPOSE 8000
CMD ["uygulama"]
