# Gunakan image Golang official berbasis Alpine sebagai base image
FROM golang:alpine AS builder

# Set working directory di dalam container
WORKDIR /app

# Copy file kode sumber ke dalam container
COPY . .

# Kompilasi aplikasi Go ke binary static.
# CGO_ENABLED=0 untuk memastikan binary yang dibuat compatible secara luas
# -o main untuk output file bernama main
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Gunakan image Alpine untuk menjalankan aplikasi Go
FROM alpine:latest  

# Copy binary aplikasi dari builder stage ke image final
COPY --from=builder /app/main .

# Expose port 8080
EXPOSE 8099

# Jalankan binary aplikasi
CMD ["./main"]
