# Usa un'immagine ufficiale di Go per buildare
FROM golang:1.22 AS builder

WORKDIR /app

# Copia go.mod e go.sum per caching delle dipendenze
COPY go.mod go.sum ./
RUN go mod download

# Copia tutto il codice sorgente
COPY . .

# Compila l'app in modalità statica (no dipendenze libc)
RUN CGO_ENABLED=0 GOOS=linux go build -o server main.go

# Usa un'immagine minimale per il runtime
FROM alpine:3.19

WORKDIR /root/

# Installiamo certificati CA per connessioni HTTPS (PostgreSQL ecc.)
RUN apk --no-cache add ca-certificates

# Copia il binario dalla fase di build
COPY --from=builder /app/server .

# Espone la porta dell'app (stessa usata nel codice)
EXPOSE 8080

# Avvia il server
CMD ["./server"]
