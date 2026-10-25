# 1. Aşama: Go kodunu derlemek için resmi Go imajını indiriyoruz
FROM golang:1.27-alpine AS builder

# Konteyner içindeki çalışma dizinini ayarlıyoruz
WORKDIR /app

# Go modül dosyalarını kopyalayıp bağımlılıkları indiriyoruz
COPY go.mod ./
# (Eğer ileride go.sum olursa onu da kopyalarız)

# Yazdığımız main.go dosyasını konteynere kopyalıyoruz
COPY . .

# Go kodumuzu "main" adında çalıştırılabilir bir dosyaya dönüştürüyoruz (derliyoruz)
RUN go build -o main .

# 2. Aşama: Uygulamayı çalıştırmak için çok hafif ve boş bir imaj kullanıyoruz (boyutu küçültmek için)
FROM alpine:latest
WORKDIR /app

# Sadece derlenmiş dosyayı ilk aşamadan buraya kopyalıyoruz
COPY --from=builder /app/main .

# 8080 portunu dışa açıyoruz
EXPOSE 8080

# Konteyner ayağa kalktığında bu komutu çalıştırıyoruz
CMD ["./main"]