// API приложения GoNews.
package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"GoNews/pkg/storage"

	"github.com/gorilla/mux"
)

// API приложения.
type API struct {
	db *storage.DB
	r  *mux.Router
}

// Конструктор API.
func New(db *storage.DB) *API {
	a := API{db: db, r: mux.NewRouter()}
	a.endpoints()
	return &a
}

// Router возвращает маршрутизатор для использования
// в качестве аргумента HTTP-сервера.
func (api *API) Router() *mux.Router {
	return api.r
}

// Регистрация методов API в маршрутизаторе запросов.
func (api *API) endpoints() {
	api.r.Use(api.logMiddleware)                                                                // middleware
	api.r.HandleFunc("/news", api.getAllNews).Methods(http.MethodGet)                           // получить список всех новостей
	api.r.HandleFunc("/news/{n}", api.posts).Methods(http.MethodGet, http.MethodOptions)        // получить n последних новостей
	api.r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./webapp")))) // веб-приложение
}

func (api *API) posts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json") // сообщает клиенту, что тело ответа будет в формате JSON
	w.Header().Set("Access-Control-Allow-Origin", "*") // разрешает запросы с любых доменов
	if r.Method == http.MethodOptions {
		return
	}
	s := mux.Vars(r)["n"]
	n, err := strconv.Atoi(s)
	if err != nil {
		http.Error(w, "Invalid parameter 'n'", http.StatusBadRequest)
		return
	}
	news, err := api.db.News(n)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(news)
}

type paginatedResponse struct {
	Data       []storage.Post `json:"data"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
	Total      int            `json:"total"`
	TotalPages int            `json:"total_pages"`
}

func (api *API) getAllNews(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	query := r.URL.Query().Get("q")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	var (
		news       []storage.Post
		total      int
		queryError error
	)

	if query != "" {
		news, total, queryError = api.db.SearchNews(query, offset, limit)
	} else {
		news, total, queryError = api.db.AllNewsPaginated(offset, limit)
	}

	if queryError != nil {
		http.Error(w, "Ошибка при получении новостей", http.StatusInternalServerError)
		return
	}

	totalPages := (total + limit - 1) / limit

	resp := paginatedResponse{
		Data:       news,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}

	json.NewEncoder(w).Encode(resp)
}
