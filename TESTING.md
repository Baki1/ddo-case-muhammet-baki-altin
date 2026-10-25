# KütüphaneNet - Test ve Doğrulama Rehberi

## A. Demo Hesapları
Sistem ayağa kalktığında aşağıdaki hesaplar otomatik olarak oluşturulur:
- **Yönetici:** seed.kayin@ddo-case.invalid / ornek_sifre_hashi
- **Öğrenci:** ogrenci@ddo-case.invalid / 123456
- **Kütüphaneci:** personel@ddo-case.invalid / 123456

## B. Doğrulama Komutu (make verify)
Tüm sistem sağlık kontrollerini ve gizli token doğrulamasını uçtan uca test etmek için kök dizinde şu komutu çalıştırın:
\`make verify\`

## C. Manuel Test Senaryoları
Sistemi sıfırlamak ve ilk haline döndürmek için: \`make reset-demo\`

**T-3: Türkçe Tam Metin Arama Testi (Görev 3)**
- **Nasıl Denenir:** `GET http://localhost:8080/api/books/search?q=algoritma`
- **Beklenen Sonuç:** Başlık, yazar veya konuda "algoritma" geçen kitapların JSON listesi dönmelidir. İstek süresi p95 < 150ms altında olmalıdır (Veritabanında GIN ve pg_trgm indeksleri kullanılmıştır).