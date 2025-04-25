package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
)

// Config содержит адреса внутренних сервисов и секрет для JWT
type Config struct {
	AuthServiceURL     string
	TaskServiceURL     string
	TemplateServiceURL string
	JWTSecret          string
}

// APIResponse общий формат ответа
type APIResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
	Error  string      `json:"error,omitempty"`
}

// NewAPIResponse создает новый APIResponse
func NewAPIResponse(status string, data interface{}, err string) APIResponse {
	return APIResponse{
		Status: status,
		Data:   data,
		Error:  err,
	}
}

// proxyRequest перенаправляет запрос к внутреннему сервису
func proxyRequest(client *http.Client, url, method string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(method, url, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return client.Do(req)
}

// JWTMiddleware проверяет JWT-токен
func JWTMiddleware(secret string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(NewAPIResponse("error", nil, "Authorization header is required"))
			return
		}

		// Ожидаем формат: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(NewAPIResponse("error", nil, "Invalid Authorization header format"))
			return
		}

		tokenStr := parts[1]
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(NewAPIResponse("error", nil, "Invalid or expired token"))
			return
		}

		// Токен валиден, продолжаем обработку запроса
		next(w, r)
	}
}

func main() {
	// Конфигурация сервисов
	config := Config{
		AuthServiceURL: "http://auth-service:8080",
		//TaskServiceURL:     "http://task-service:8080",
		//TemplateServiceURL: "http://template-service:8080",
		JWTSecret: "your_jwt_secret_key", // В продакшене храните в переменной окружения
	}

	// HTTP клиент
	client := &http.Client{}

	// Маршрутизатор
	router := mux.NewRouter()

	// Эндпоинты для Auth Service (без JWT)
	router.HandleFunc("/api/v1/register", func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		resp, err := proxyRequest(client, config.AuthServiceURL+"/register", r.Method, body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(NewAPIResponse("error", nil, err.Error()))
			return
		}
		defer resp.Body.Close()
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		json.NewEncoder(w).Encode(NewAPIResponse("success", result, ""))
	}).Methods("POST")

	router.HandleFunc("/api/v1/login", func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		resp, err := proxyRequest(client, config.AuthServiceURL+"/login", r.Method, body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(NewAPIResponse("error", nil, err.Error()))
			return
		}
		defer resp.Body.Close()
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		json.NewEncoder(w).Encode(NewAPIResponse("success", result, ""))
	}).Methods("POST")

	//// Эндпоинты для Task Service (с JWT)
	//router.HandleFunc("/api/tasks", JWTMiddleware(config.JWTSecret, func(w http.ResponseWriter, r *http.Request) {
	//	body, _ := ioutil.ReadAll(r.Body)
	//	resp, err := proxyRequest(client, config.TaskServiceURL+"/tasks", r.Method, body)
	//	if err != nil {
	//		w.WriteHeader(http.StatusInternalServerError)
	//		json.NewEncoder(w).Encode(NewAPIResponse("error", nil, err.Error()))
	//		return
	//	}
	//	defer resp.Body.Close()
	//	var result map[string]interface{}
	//	json.NewDecoder(resp.Body).Decode(&result)
	//	json.NewEncoder(w).Encode(NewAPIResponse("success", result, ""))
	//})).Methods("GET", "POST")
	//
	//// Эндпоинты для Template Service (с JWT)
	//router.HandleFunc("/api/templates", JWTMiddleware(config.JWTSecret, func(w http.ResponseWriter, r *http.Request) {
	//	body, _ := ioutil.ReadAll(r.Body)
	//	resp, err := proxyRequest(client, config.TemplateServiceURL+"/templates", r.Method, body)
	//	if err != nil {
	//		w.WriteHeader(http.StatusInternalServerError)
	//		json.NewEncoder(w).Encode(NewAPIResponse("error", nil, err.Error()))
	//		return
	//	}
	//	defer resp.Body.Close()
	//	var result map[string]interface{}
	//	json.NewDecoder(resp.Body).Decode(&result)
	//	json.NewEncoder(w).Encode(NewAPIResponse("success", result, ""))
	//})).Methods("GET", "POST")

	// Запуск сервера
	log.Println("API Gateway running on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
