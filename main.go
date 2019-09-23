package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/nlopes/slack"
	database "github.com/sirkz/standuplytest/db"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
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

	router := mux.NewRouter()

	router = router.PathPrefix("/api/v1").Subrouter()

	router.HandleFunc("/order", Order(config.ApiKey)).Methods("POST")

	router.Use(LogMiddleware)
	router.Use(PanicMiddleware)

	if err := http.ListenAndServe(":8443", router); err != nil {
		log.Warn(err)
	}
}

func Order(apiKey string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		verifier, err := slack.NewSecretsVerifier(r.Header, apiKey)
		if err != nil {
			log.Error(err)
			http.Error(w, "InternalError", 500)
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

		nameIndex := strings.IndexByte(s.Text, ' ')
		fmt.Println(s.Text[:nameIndex])
		sizeIndex := strings.IndexByte(s.Text, ' ')
		fmt.Println(s.Text[nameIndex:sizeIndex])
		fmt.Println(s.Text[sizeIndex:])
		sendMsg(slack.Msg{Text: "Your order has been accepted and is being processed"}, w)
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
				log.Error("recovered", err)
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
