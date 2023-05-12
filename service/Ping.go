package service

import (
	"net/http"
)

func Ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Pong"))
}

func Test(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("BaseUrl: " + Config.BaseUrl))
}
