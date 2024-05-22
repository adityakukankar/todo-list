package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/thedevsaddam/renderer"
	mgo "gopkg.in/mgo.v2"

	"github.com/adityakukankar/todo/handlers"
	"github.com/adityakukankar/todo/utils"
)

var rnd *renderer.Render
var db *mgo.Database

func init() {
	rnd = renderer.New()
	sess, err := mgo.Dial(utils.HostName)
	utils.CheckErr(err)
	sess.SetMode(mgo.Monotonic, true)
	db = sess.DB(utils.DBName)
}

func main() {

	stopChannel := make(chan os.Signal)
	signal.Notify(stopChannel, os.Interrupt)

	homeHandler := func(w http.ResponseWriter, r *http.Request) {
		handlers.HomeHandler(w, r, rnd)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", homeHandler)
	r.Mount("/todo", handlers.TodoHandlers(rnd, db))

	// server definition
	srv := &http.Server{
		Addr:         utils.Port,
		Handler:      r,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// channel to start server
	go func() {
		log.Println("Listening on port", utils.Port)
		if err := srv.ListenAndServe(); err != nil {
			log.Println("listen:%s\n", err)
		}
	}()

	<-stopChannel
	log.Println("Shutting down server..")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	srv.Shutdown(ctx)
	defer cancel()
	log.Println("server gracefully stopped!!")

}
