package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-pkgz/auth/v2"
	"github.com/go-pkgz/auth/v2/token"

	_ "modernc.org/sqlite"
)

var tpl = template.Must(template.ParseFiles("templates/index.html"))

type App struct {
	db  *sql.DB
	svc *auth.Service
}

func main() {
	db, err := sql.Open("sqlite", "file:locations.db?_foreign_keys=1")
	if err != nil {
		log.Fatal(err)
	}
	if err := initDB(db); err != nil {
		log.Fatal(err)
	}

	opts := auth.Opts{
		SecretReader: token.SecretFunc(func(aud string) (string, error) {
			s := os.Getenv("AUTH_SECRET")
			if s == "" {
				s = "devsecret"
			}
			return s, nil
		}),
		TokenDuration:  time.Hour * 24,
		CookieDuration: time.Hour * 24 * 7,
		Issuer:         "socialmap",
		URL:            os.Getenv("BASE_URL"),
	}

	svc := auth.NewService(opts)

	// Add providers: dev for fast local login and optional Google
	svc.AddProvider("dev", "", "")
	if cid := os.Getenv("GOOGLE_CID"); cid != "" {
		svc.AddProvider("google", cid, os.Getenv("GOOGLE_CSECRET"))
	}

	authRoutes, avaRoutes := svc.Handlers()
	m := svc.Middleware()

	app := &App{db: db, svc: svc}

	mux := http.NewServeMux()
	mux.Handle("/auth/", http.StripPrefix("/auth", authRoutes))
	mux.Handle("/avatar/", http.StripPrefix("/avatar", avaRoutes))

	mux.HandleFunc("/", app.index)

	// API to get/save location (protected)
	mux.Handle("/api/location", m.Auth(http.HandlerFunc(app.locationHandler)))

	// serve templates/static from working dir

	addr := ":8080"
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func initDB(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS users (
        user_id TEXT PRIMARY KEY,
        lat REAL,
        lon REAL
    )`)
	return err
}

func (a *App) index(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"MapboxToken": os.Getenv("MAPBOX_TOKEN"),
	}
	tpl.Execute(w, data)
}

func (a *App) locationHandler(w http.ResponseWriter, r *http.Request) {
	u, err := token.GetUserInfo(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	switch r.Method {
	case http.MethodGet:
		lat, lon, ok, err := a.loadLocation(r.Context(), u.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !ok {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		json.NewEncoder(w).Encode(map[string]float64{"lat": lat, "lon": lon})
	case http.MethodPost:
		var p struct{ Lat, Lon float64 }
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := a.saveLocation(r.Context(), u.ID, p.Lat, p.Lon); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *App) saveLocation(ctx context.Context, userID string, lat, lon float64) error {
	_, err := a.db.ExecContext(ctx, `INSERT INTO users(user_id, lat, lon) VALUES(?,?,?) ON CONFLICT(user_id) DO UPDATE SET lat=excluded.lat, lon=excluded.lon`, userID, lat, lon)
	return err
}

func (a *App) loadLocation(ctx context.Context, userID string) (float64, float64, bool, error) {
	var lat, lon float64
	row := a.db.QueryRowContext(ctx, `SELECT lat, lon FROM users WHERE user_id = ?`, userID)
	if err := row.Scan(&lat, &lon); err != nil {
		if err == sql.ErrNoRows {
			return 0, 0, false, nil
		}
		return 0, 0, false, err
	}
	return lat, lon, true, nil
}
