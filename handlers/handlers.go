package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/thedevsaddam/renderer"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	. "github.com/adityakukankar/todo/models"
	"github.com/adityakukankar/todo/utils"
)

type TodoHandler struct {
	rnd *renderer.Render
	db  *mgo.Database
}

func (h *TodoHandler) createTodo(w http.ResponseWriter, r *http.Request) {
	var t Todo

	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		h.rnd.JSON(w, http.StatusProcessing, err)
		return
	}

	if t.Title == "" {
		h.rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "the title is required",
		})
		return
	}

	tm := TodoModel{
		ID:        bson.NewObjectId(),
		Title:     t.Title,
		Completed: false,
		CreatedAt: time.Now(),
	}

	if err := h.db.C(utils.CollectionName).Insert(&tm); err != nil {
		h.rnd.JSON(w, http.StatusProcessing, renderer.M{
			"message": "Failed to save todo",
			"error":   err,
		})
		return
	}

	h.rnd.JSON(w, http.StatusCreated, renderer.M{
		"message": "todo created successfully",
		"todo_id": tm.ID.Hex(),
	})
}

func (h *TodoHandler) updateTodo(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "id"))

	if !bson.IsObjectIdHex(id) {
		h.rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "Id is invalid",
		})
		return
	}

	var t Todo

	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		h.rnd.JSON(w, http.StatusProcessing, err)
		return
	}

	if t.Title == "" {
		h.rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "the title is required",
		})
		return
	}

	if err := h.db.C(utils.CollectionName).Update(
		bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"title": t.Title, "completed": t.Completed},
	); err != nil {
		h.rnd.JSON(w, http.StatusProcessing, renderer.M{
			"message": "failed to update todo",
			"error":   err,
		})
		return
	}
}

func (h *TodoHandler) fetchTodos(w http.ResponseWriter, r *http.Request) {
	todos := []TodoModel{}
	if err := h.db.C(utils.CollectionName).Find(bson.M{}).All(&todos); err != nil {
		h.rnd.JSON(w, http.StatusProcessing, renderer.M{
			"message": "Failed to fetch todo",
			"error":   err,
		})
		return
	}
	todoList := []Todo{}

	for _, i := range todos {
		todoList = append(todoList, Todo{
			ID:        i.ID.Hex(),
			Title:     i.Title,
			Completed: i.Completed,
			CreatedAt: i.CreatedAt,
		})
	}
	h.rnd.JSON(w, http.StatusOK, renderer.M{
		"data": todoList,
	})
}

func (h *TodoHandler) deleteTodo(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "id"))

	if !bson.IsObjectIdHex(id) {
		h.rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "Id is invalid",
		})
		return
	}

	if err := h.db.C(utils.CollectionName).RemoveId(bson.ObjectIdHex(id)); err != nil {
		h.rnd.JSON(w, http.StatusProcessing, renderer.M{
			"message": "failed to delete todo",
			"error":   err,
		})
		return
	}

	h.rnd.JSON(w, http.StatusOK, renderer.M{
		"message": "todo deleted successfully",
	})
}

func HomeHandler(w http.ResponseWriter, r *http.Request, rnd *renderer.Render) {
	err := rnd.Template(w, http.StatusOK, []string{"resources/home.tpl"}, nil)
	utils.CheckErr(err)
}

func TodoHandlers(Render *renderer.Render, DB *mgo.Database) http.Handler {
	handler := &TodoHandler{
		rnd: Render,
		db:  DB,
	}

	rg := chi.NewRouter()
	rg.Group(func(r chi.Router) {
		r.Get("/", handler.fetchTodos)
		r.Post("/", handler.createTodo)
		r.Put("/{id}", handler.updateTodo)
		r.Delete("/{id}", handler.deleteTodo)
	})
	return rg
}
