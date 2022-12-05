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
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var rnd *renderer.Render
var db *mgo.Database

const (
	hostName   string = "localhost:27017"
	dbName     string = "demo_todo"
	collection string = "todo"
	port       string = ":9000"
)

type (
	todoModel struct {
		ID         bson.Objectid `bson:"_id,omitempty"`
		Title      string        `bson:"title"`
		Completed  bool          `bson:"completed"`
		creadtedAt time.Time     `bson:"createdAt"`
	}

	todo struct {
		ID        string    `json:"id"`
		Title     string    `json:"title"`
		Completed string    `json:"completed"`
		createdAt time.Time `json:"created_at"`
	}
)

func init() {
	rnd = renderer.New()
	sess, err := mgo.Dial(hostName)
	checkErr(err)
	sess.SetMode(mgo.Monotonic, true)

	// setmode - function, changes the consistency mode for the session - strong, monotonic, eventual

	db = sess.DB(dbName)
}

func homeHandler(w http.ResponseWriter,r *http.Request){
	err : rnd.Template(w, http.StatusOK, []string("static/home.tpl"))
	checkErr(err)
}

func fetchTodo(w http.ResponseWriter,r *http.Request){
	todos := []todoModel{}
	if err != db.C(collectionName).Find(bson.M{}).All(&todo); err!=nil{
		rnd.JSON(w, http.StatusProcessing, renderer.M{
			"message":"Failed to fetch todo",
			"error": err,
		})
		return
	}
	todoList := []todo{}

	for _, t := range todos{
		todoList = append(todoList, todo{
			ID: t.ID.Hex(),
			Title: t.Title,
			Completed: t.Completed,
			createdAt: t.creadtedAt,
		})
	}
	rnd.JSON(w, http.StatusOK, renderer.M{
		"data": todoList,
	})
}

func main() {

	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", homeHandler)
	r.Mount("/todo", todoHandler())

	srv := &http.Server{
		Addr:         port,
		Handler:      r,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}


	//thi is a channel which will start our server
	//listeandserver function is inside http package, helps us to create a server
	go func() {
		log.Println("Listening on port", port)
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("listen:%s\n", err)
		}
	}()

	<- stopChan
	log.Println("shutting down the function")
	ctx, cancel := context.WithTimeout(context.Background(),5*time.Second)
	srv.Shutdown(ctx)
	defer cancel(
		log.Println("server gracefully stopped")
	)
}

func todoHandler() http.Handler {
	rg := chi.NewRouter()
	rg.Group(func(r chi.Router) {
		r.Get("/", fetchTodos)
		r.Post("/", createTodo)
		r.Put("/", updateTodo)
		r.Delete("/(id)", deleteTodo)
	})
	return rg
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}