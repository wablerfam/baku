package main

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type Server struct {
	Router *mux.Router
	Port   int
}

func InitServer(config ServerConfig) *Server {
	s := new(Server)
	s.Router = mux.NewRouter()
	s.Port = config.Port
	return s
}

func (srv Server) Run(job JobConfig, database Database) {
	Logger("info", "baku.server", "baku server up")

	handler := Handler{job, database}
	handler.Use(srv.Router)

	http.ListenAndServe((strings.Join([]string{":", strconv.Itoa(srv.Port)}, "")), srv.Router)
}
