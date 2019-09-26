package main

import (
	"context"
	"encoding/json"
	"flag"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
	"github.com/nlopes/slack"
	database "github.com/sirkz/standuplytest/db"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	config, err := LoadConfig()
	if err != nil {
		log.Warn(err)
		return
	}

	db, err := database.Conn(config.ConfigDB)
	if err != nil {
		log.Warn(err)
		return
	}
	defer func() {
		err = db.Close(context.Background())
		log.Warn(err)
	}()
	log.Info("Connect to database")

	syncDB := flag.Bool("sync-db", false, "Creates all necessary tables")

	flag.Parse()

	if *syncDB {
		err = database.CreateModels(db)
		if err != nil {
			log.Warn(err)
			return
		}
		log.Info("Successful")
		return
	}

	indexFile, err := os.Open("static/index.html")
	if err != nil {
		log.Warn(err)
		return
	}
	defer func() {
		err = indexFile.Close()
		log.Error(err)
	}()

	indexHtml, err := ioutil.ReadAll(indexFile)
	if err != nil {
		log.Error(err)
		return
	}

	router := mux.NewRouter()
	router.Use(LogMiddleware)
	router.Use(PanicMiddleware)

	router.HandleFunc("/", IndexHandle(indexHtml)).Methods("GET")

	apiRouter := router.PathPrefix("/api/v1").Subrouter()

	apiRouter.HandleFunc("/Order", SetOrder(config.ApiKey, db)).Methods("POST")
	apiRouter.HandleFunc("/listOrders", ListOrders(db)).Methods("GET")

	if err := http.ListenAndServe(":8443", router); err != nil {
		log.Warn(err)
	}
}

func SetOrder(apiKey string, db *pgx.Conn) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		verifier, err := slack.NewSecretsVerifier(r.Header, apiKey)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		r.Body = ioutil.NopCloser(io.TeeReader(r.Body, &verifier))
		s, err := slack.SlashCommandParse(r)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err = verifier.Ensure(); err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		createOrder(s, w, db)
		sendMsg(slack.Msg{Text: "Your Order has been accepted and is being processed"}, w)
	}
}

func createOrder(s slack.SlashCommand, w http.ResponseWriter, db *pgx.Conn) {
	params := strings.SplitN(s.Text, " ", 3)
	if len(params) < 3 {
		sendMsg(slack.Msg{Text: "First enter the name of the pizza, then the size and address"}, w)
		return
	}
	pizza := params[0]
	size, err := strconv.Atoi(params[1])
	if err != nil {
		sendMsg(slack.Msg{Text: "Invalid size"}, w)
		return
	}
	address := params[2]

	userId, err := database.FindUserIdBySlackId(db, s.UserID)
	if err != nil {
		log.Error(err)
		return
	}
	if userId == "-1" {
		userId, err = database.CreateUser(db, s.UserID, s.UserName)
		if err != nil {
			log.Error(err)
			return
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	_, err = db.Exec(ctx, "INSERT INTO Order(user_id, pizza, size, address) VALUES ($1, $2, $3, $4);", userId, pizza, size, address)
	if err != nil {
		log.Error(err)
	}
	return
}

type Order struct {
	Number  int    `json:"number" db:"number"`
	Pizza   string `json:"pizza" db:"pizza"`
	Size    int    `json:"size" db:"size"`
	Address string `json:"address" db:"address"`
}

func ListOrders(db *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var listOrders []Order
		rows, err := db.Query(context.Background(), "SELECT number,pizza,size,address FROM orders;")
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		var order Order
		for rows.Next() {
			err = rows.Scan(&order.Number, &order.Pizza, &order.Size, &order.Address)
			if err != nil {
				log.Error(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			listOrders = append(listOrders, order)
		}
		listOrdersJson, err := json.Marshal(listOrders)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = w.Write(listOrdersJson)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func LogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timeReq := time.Now()
		defer func(start time.Time) {
			log.Info(r.URL.Path, time.Since(start))
		}(timeReq)
		next.ServeHTTP(w, r)
	})
}

func PanicMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Error("recovered ", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func sendMsg(msg slack.Msg, w http.ResponseWriter) {
	b, err := json.Marshal(msg)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(b)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func IndexHandle(indexHTML []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write(indexHTML)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
