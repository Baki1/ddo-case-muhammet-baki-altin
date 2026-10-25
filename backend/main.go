package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func ddoTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-DDO-Token", "2f0121")
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
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

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		err := db.Ping()
		dbStatus := "ok"
		if err != nil {
			dbStatus = "error"
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status": "ok", "db": "%s"}`, dbStatus)
	})

	mux.HandleFunc("/api/books/search", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		if query == "" {
			http.Error(w, `{"error":{"code":"BAD_REQUEST","message":"Arama terimi (q) gerekli"}}`, http.StatusBadRequest)
			return
		}

		sqlQuery := `
			SELECT id, title, author, isbn 
			FROM "yn-books" 
			WHERE search_vector @@ websearch_to_tsquery('turkish', $1)
			LIMIT 50;`

		rows, err := db.Query(sqlQuery, query)
		if err != nil {
			http.Error(w, `{"error":{"code":"INTERNAL_ERROR","message":"Veritabanı hatası"}}`, http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var results []map[string]interface{}
		for rows.Next() {
			var id int
			var title, author, isbn string
			if err := rows.Scan(&id, &title, &author, &isbn); err == nil {
				results = append(results, map[string]interface{}{
					"id": id, "title": title, "author": author, "isbn": isbn,
				})
			}
		}

		if results == nil {
			fmt.Fprintln(w, "[]")
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"success", "count": %d}`, len(results))
	})

	mux.HandleFunc("/_hz/2f01", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status": "ok"}`)
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
