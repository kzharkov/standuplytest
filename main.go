package main

import (
	"encoding/json"
	"github.com/nlopes/slack"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	apiKey := "xoxp-770274988487-770274988743-770642068102-6163d973a4d1d798cdf96bf0a01c8dc8"
	router := http.NewServeMux()

	router.HandleFunc("/order", Order(apiKey))

	if err := http.ListenAndServe(":8443", router); err != nil {
		log.Println(err)
	}
}

func Order(apiKey string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		verifier, err := slack.NewSecretsVerifier(r.Header, apiKey)
		if err != nil {
			log.Println(err)
			http.Error(w, "InternalError", 500)
		}
		r.Body = ioutil.NopCloser(io.TeeReader(r.Body, &verifier))
		s, err := slack.SlashCommandParse(r)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err = verifier.Ensure(); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		params := &slack.Msg{Text: s.Text}
		b, err := json.Marshal(params)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(b)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
