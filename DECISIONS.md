# Mimari Karar Kayıtları (ADR)

## 1. Tam Metin Arama (Full-Text Search) Altyapısı
- **Bağlam:** 500.000 kayıtta p95 < 150ms hızında, Türkçe karakter ve yazım hatası toleranslı arama isteniyordu.
- **Seçenekler:** Elasticsearch gibi harici bir servis kurmak vs. PostgreSQL'in yerleşik arama yeteneklerini kullanmak.
- **Karar:** PostgreSQL `tsvector`, GIN indeksi ve `pg_trgm` eklentisi kullanıldı.
- **Sonuç:** Harici servis bağımlılığı ortadan kalktı, Docker mimarisi hafif tutuldu ve performans kriteri `EXPLAIN` ölçümleriyle sağlandı.

## 2. Backend HTTP Yönlendirici (Router) Seçimi
- **Bağlam:** Go 1.22+ ile REST API geliştirilmesi gerekiyordu.
- **Seçenekler:** Gin, Fiber, Echo gibi dış kütüphaneler vs. Go standart kütüphanesi (`net/http`).
- **Karar:** Go 1.22 ile gelen geliştirilmiş `http.NewServeMux` standart kütüphanesi tercih edildi.
- **Sonuç:** Dış bağımlılıklar en aza indirildi, derleme süresi kısaldı ve `X-DDO-Token` gibi özel middleware yapıları standart arayüzlerle kolayca kurgulandı.

## 3. Frontend Teknoloji Yığını
- **Bağlam:** Zaman kısıtı altında, API ile haberleşen ve Docker üzerinden tek komutla çalışan bir arayüz isteniyordu.
- **Seçenekler:** React/Vue.js (Node.js gerektirir) vs. Standart HTML + Vanilla JS.
- **Karar:** Nginx imajı üzerinden sunulan saf HTML ve JavaScript kullanıldı.
- **Sonuç:** `node_modules` yükü ve build süreleri tamamen sıfırlandı. Sistem hedeflendiği gibi soğuk makinede çok kısa sürede ayağa kalkabilir hale geldi.