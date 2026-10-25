.PHONY: verify reset-demo

verify:
	@echo "=== DDO Case Hızlı Doğrulama Başlıyor ==="
	@echo "API Sağlık ve Gizli Token Kontrolü (GET /health):"
	@curl -s -I http://localhost:8080/health | grep X-DDO-Token || (echo "FAIL: Token eksik!" && exit 1)
	@echo "PASS: Sistem aktif ve gizli başlıklar yanıt veriyor."
	@echo "=== Doğrulama Tamamlandı ==="

reset-demo:
	docker compose down -v
	docker compose up --build -d
	@echo "Sistem ve veritabanı sıfırdan oluşturuldu."