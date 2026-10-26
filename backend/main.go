package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func ddoTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-DDO-Token", "2f0121")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Idempotency-Key")
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeErr(w http.ResponseWriter, status int, code, msg string) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": msg,
		},
	})
}

func main() {
	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		dbUrl = "postgres://ddouser:ddopassword123@db:5432/kutuphanenet?sslmode=disable"
	}

	db, err := sql.Open("pgx", dbUrl)
	if err != nil {
		log.Fatalf("Veritabanına bağlanılamadı: %v", err)
	}
	defer db.Close()

	mux := http.NewServeMux()

	// 1. Sağlık Ucu
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		err := db.Ping()
		dbStatus := "ok"
		if err != nil {
			dbStatus = "error"
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status": "ok", "db": "%s"}`, dbStatus)
	})

	// 2. Gizli DDO Sağlık Ucu
	mux.HandleFunc("/_hz/2f01", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status": "ok"}`)
	})

	// 3. Kimlik Doğrulama (Login)
	mux.HandleFunc("POST /api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErr(w, http.StatusBadRequest, "BAD_REQUEST", "Geçersiz JSON")
			return
		}
		var id int
		var pwHash, role string
		err := db.QueryRow(`SELECT id, password_hash, role FROM "yn-ej-users" WHERE email = $1`, req.Email).Scan(&id, &pwHash, &role)
		if err != nil || pwHash != req.Password {
			writeErr(w, http.StatusUnauthorized, "UNAUTHORIZED", "E-posta veya parola hatalı")
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "success",
			"user":   map[string]any{"id": id, "email": req.Email, "role": role},
		})
	})

	// 4. T-3 & F-5: Türkçe Tam Metin + Trigram Arama (Senin mevcut ucun + boş sorguda tüm liste)
	mux.HandleFunc("/api/books/search", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		if query == "" {
			writeErr(w, http.StatusBadRequest, "BAD_REQUEST", "Arama terimi (q) gerekli")
			return
		}

		sqlQuery := `
			SELECT id, title, author, COALESCE(isbn, ''), COALESCE(subject, 'Genel')
			FROM "yn-books"
			WHERE search_vector @@ websearch_to_tsquery('turkish', $1)
			   OR title % $1
			   OR title ILIKE '%' || $1 || '%'
			LIMIT 50;`

		rows, err := db.Query(sqlQuery, query)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Veritabanı hatası")
			return
		}
		defer rows.Close()

		results := make([]map[string]any, 0)
		for rows.Next() {
			var id int
			var title, author, isbn, subject string
			if err := rows.Scan(&id, &title, &author, &isbn, &subject); err != nil {
				writeErr(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Okuma hatası")
				return
			}
			results = append(results, map[string]any{
				"id": id, "title": title, "author": author, "isbn": isbn, "subject": subject,
			})
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "success",
			"count":  len(results),
			"books":  results,
		})
	})

	// 5. F-1: Katalog Kitap Listesi ve Yeni Kitap Ekleme
	mux.HandleFunc("GET /api/books", func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`SELECT id, title, author, COALESCE(isbn, ''), COALESCE(subject, 'Genel') FROM "yn-books" ORDER BY id DESC LIMIT 50`)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
			return
		}
		defer rows.Close()
		books := make([]map[string]any, 0)
		for rows.Next() {
			var id int
			var title, author, isbn, subject string
			if rows.Scan(&id, &title, &author, &isbn, &subject) == nil {
				books = append(books, map[string]any{"id": id, "title": title, "author": author, "isbn": isbn, "subject": subject})
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "books": books})
	})

	mux.HandleFunc("POST /api/books", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Title   string `json:"title"`
			Author  string `json:"author"`
			ISBN    string `json:"isbn"`
			Subject string `json:"subject"`
			Barcode string `json:"barcode"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" {
			writeErr(w, http.StatusBadRequest, "BAD_REQUEST", "Kitap başlığı zorunludur")
			return
		}
		var bookID int
		err := db.QueryRow(
			`INSERT INTO "yn-books" (title, author, isbn, subject) VALUES ($1, $2, $3, $4) RETURNING id`,
			req.Title, req.Author, req.ISBN, req.Subject).Scan(&bookID)
		if err != nil {
			writeErr(w, http.StatusConflict, "DUPLICATE_ISBN", "ISBN zaten mevcut")
			return
		}
		if req.Barcode != "" {
			_, _ = db.Exec(`INSERT INTO "yn-copies" (book_id, barcode, status) VALUES ($1, $2, 'available')`, bookID, req.Barcode)
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "created", "book_id": bookID})
	})

	// 6. T-1 & F-2: Ödünç Verme (Eşzamanlılık Kilitli & Üye Tipi Limitli)
	mux.HandleFunc("POST /api/loans", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			UserID  int    `json:"user_id"`
			Barcode string `json:"barcode"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.UserID == 0 || req.Barcode == "" {
			writeErr(w, http.StatusBadRequest, "BAD_REQUEST", "user_id ve barcode gerekli")
			return
		}

		tx, err := db.Begin()
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "TX_ERROR", err.Error())
			return
		}
		defer tx.Rollback()

		var copyID int
		var status string
		err = tx.QueryRow(`SELECT id, status FROM "yn-copies" WHERE barcode = $1 FOR UPDATE`, req.Barcode).Scan(&copyID, &status)
		if err != nil {
			writeErr(w, http.StatusNotFound, "NOT_FOUND", "Barkod bulunamadı")
			return
		}
		if status != "available" {
			writeErr(w, http.StatusConflict, "COPY_UNAVAILABLE", "Bu nüsha şu an ödünç verilemez")
			return
		}

		// Üye rolüne göre kitap limiti (ogrenci: 5, akademisyen: 20)
		var role string
		_ = tx.QueryRow(`SELECT role FROM "yn-ej-users" WHERE id = $1`, req.UserID).Scan(&role)
		maxBooks := 5
		loanDays := 14
		if role == "akademisyen" {
			maxBooks = 20
			loanDays = 30
		}

		var activeCount int
		_ = tx.QueryRow(`SELECT COUNT(*) FROM "yn-loans" WHERE user_id = $1 AND returned_at IS NULL`, req.UserID).Scan(&activeCount)
		if activeCount >= maxBooks {
			writeErr(w, http.StatusConflict, "LIMIT_EXCEEDED", "Üye ödünç limitine ulaştı")
			return
		}

		dueDate := time.Now().AddDate(0, 0, loanDays)
		var loanID int
		err = tx.QueryRow(
			`INSERT INTO "yn-loans" (copy_id, user_id, due_date) VALUES ($1, $2, $3) RETURNING id`,
			copyID, req.UserID, dueDate).Scan(&loanID)
		if err != nil {
			writeErr(w, http.StatusConflict, "CONCURRENT_CONFLICT", "Nüsha eşzamanlı işlemde ödünç verildi")
			return
		}

		_, _ = tx.Exec(`UPDATE "yn-copies" SET status = 'borrowed' WHERE id = $1`, copyID)
		_, _ = tx.Exec(`INSERT INTO "yn-audit_logs" (action_type, table_name, record_id, performed_by) VALUES ('BORROW', 'yn-loans', $1, $2)`, loanID, req.UserID)
		_ = tx.Commit()

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "borrowed", "loan_id": loanID, "due_date": dueDate})
	})

	// 7. T-1 & T-2 & F-4: İade, Kuruş Bazlı Ceza Hesabı ve Rezervasyon Kuyruğu Tetikleme
	mux.HandleFunc("POST /api/loans/{id}/return", func(w http.ResponseWriter, r *http.Request) {
		loanID, _ := strconv.Atoi(r.PathValue("id"))
		tx, err := db.Begin()
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "TX_ERROR", err.Error())
			return
		}
		defer tx.Rollback()

		var copyID, bookID int
		var dueDate time.Time
		var returnedAt sql.NullTime
		err = tx.QueryRow(`
			SELECT l.copy_id, c.book_id, l.due_date, l.returned_at
			FROM "yn-loans" l JOIN "yn-copies" c ON c.id = l.copy_id
			WHERE l.id = $1 FOR UPDATE`, loanID).Scan(&copyID, &bookID, &dueDate, &returnedAt)
		if err != nil {
			writeErr(w, http.StatusNotFound, "NOT_FOUND", "Ödünç kaydı bulunamadı")
			return
		}
		if returnedAt.Valid {
			writeErr(w, http.StatusConflict, "ALREADY_RETURNED", "Kitap zaten iade edilmiş")
			return
		}

		now := time.Now()
		fineKurus := 0
		if now.After(dueDate) {
			days := int(now.Sub(dueDate).Hours()/24) + 1
			fineKurus = days * 500 // Günlük 500 kuruş (5 TL) tam sayı ceza
		}

		_, _ = tx.Exec(`UPDATE "yn-loans" SET returned_at = $1, fine_amount = $2 WHERE id = $3`, now, fineKurus, loanID)

		// Rezervasyon kuyruğunda bekleyen ilk üyeye 3 gün tut (T-2)
		var resID int
		err = tx.QueryRow(`
			SELECT id FROM "yn-reservations"
			WHERE book_id = $1 AND status = 'waiting'
			ORDER BY created_at ASC LIMIT 1 FOR UPDATE SKIP LOCKED`, bookID).Scan(&resID)

		nextStatus := "available"
		if err == nil && resID > 0 {
			nextStatus = "reserved"
			_, _ = tx.Exec(`UPDATE "yn-reservations" SET status = 'active', notified_at = $1, expires_at = $2 WHERE id = $3`,
				now, now.Add(72*time.Hour), resID)
		}
		_, _ = tx.Exec(`UPDATE "yn-copies" SET status = $1 WHERE id = $2`, nextStatus, copyID)
		_ = tx.Commit()

		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":      "returned",
			"loan_id":     loanID,
			"fine_amount": fineKurus,
			"copy_status": nextStatus,
		})
	})

	// 8. F-3 & T-2: Rezervasyon Kuyruğuna Ekleme
	mux.HandleFunc("POST /api/reservations", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			BookID int `json:"book_id"`
			UserID int `json:"user_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.BookID == 0 || req.UserID == 0 {
			writeErr(w, http.StatusBadRequest, "BAD_REQUEST", "book_id ve user_id gerekli")
			return
		}
		var resID int
		err := db.QueryRow(`INSERT INTO "yn-reservations" (book_id, user_id, status) VALUES ($1, $2, 'waiting') RETURNING id`,
			req.BookID, req.UserID).Scan(&resID)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "waiting", "reservation_id": resID})
	})

	// 9. T-4 & F-6: Envanter Sayımı ve Durum Makinesi (Ödünçteki nüsha 'lost' olamaz)
	mux.HandleFunc("PATCH /api/copies/{id}/status", func(w http.ResponseWriter, r *http.Request) {
		copyID, _ := strconv.Atoi(r.PathValue("id"))
		var req struct {
			Status string `json:"status"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)

		var currentStatus string
		err := db.QueryRow(`SELECT status FROM "yn-copies" WHERE id = $1`, copyID).Scan(&currentStatus)
		if err != nil {
			writeErr(w, http.StatusNotFound, "NOT_FOUND", "Nüsha bulunamadı")
			return
		}
		if currentStatus == "borrowed" && req.Status == "lost" {
			writeErr(w, http.StatusConflict, "INVALID_TRANSITION", "Ödünçteki nüsha kayıp (lost) olarak işaretlenemez")
			return
		}
		_, _ = db.Exec(`UPDATE "yn-copies" SET status = $1 WHERE id = $2`, req.Status, copyID)
		_ = json.NewEncoder(w).Encode(map[string]any{"copy_id": copyID, "old_status": currentStatus, "new_status": req.Status})
	})

	// 10. F-7: Raporlar
	mux.HandleFunc("GET /api/reports/summary", func(w http.ResponseWriter, r *http.Request) {
		var books, users, activeLoans, waitingRes, totalFines int
		_ = db.QueryRow(`SELECT COUNT(*) FROM "yn-books"`).Scan(&books)
		_ = db.QueryRow(`SELECT COUNT(*) FROM "yn-ej-users"`).Scan(&users)
		_ = db.QueryRow(`SELECT COUNT(*) FROM "yn-loans" WHERE returned_at IS NULL`).Scan(&activeLoans)
		_ = db.QueryRow(`SELECT COUNT(*) FROM "yn-reservations" WHERE status IN ('waiting', 'active')`).Scan(&waitingRes)
		_ = db.QueryRow(`SELECT COALESCE(SUM(fine_amount), 0) FROM "yn-loans"`).Scan(&totalFines)

		_ = json.NewEncoder(w).Encode(map[string]any{
			"total_books":         books,
			"total_users":         users,
			"active_loans":        activeLoans,
			"active_reservations": waitingRes,
			"total_fines_kurus":   totalFines,
		})
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: ddoTokenMiddleware(mux),
	}

	fmt.Println("Backend 8080 portunda çalışıyor, veritabanı bağlandı...")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Sunucu hatası: %v", err)
	}
}
