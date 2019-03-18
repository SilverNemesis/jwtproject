package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/silvernemesis/jwtproject/internal/api"
	"github.com/silvernemesis/jwtproject/internal/dotenv"
	"github.com/silvernemesis/jwtproject/internal/security"
)

// User is the interface for user authentication/authorization
type User interface {
	AddUser(username, password string) error
	VerifyUser(username, password string) error
	CreateToken() (string, error)
}

func main() {
	err := dotenv.ProcessFile(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	fmt.Println("Starting the application...")

	username := os.Getenv("ADMIN_USERNAME")
	password := os.Getenv("ADMIN_PASSWORD")
	secretKey := os.Getenv("SECRET_KEY")
	certFile := os.Getenv("CERT_FILE")
	keyFile := os.Getenv("KEY_FILE")

	security.SetSecretKey(secretKey)
	security.AddUser(username, password, "admin")

	useTLS := false
	httpPrefix := "http"

	if certFile != "" && keyFile != "" {
		useTLS = true
		httpPrefix = "https"
	}

	router := mux.NewRouter()

	router.HandleFunc("/authenticate", api.HandleLogin).Methods("POST")
	router.HandleFunc("/test", api.ValidateMiddleware(testEndpoint)).Methods("GET")

	go func() {
		if useTLS {
			log.Fatal(http.ListenAndServeTLS(":5000", certFile, keyFile, router))
		} else {
			log.Fatal(http.ListenAndServe(":5000", router))
		}
	}()

	client := &http.Client{}
	headers := make(map[string]string)

	message := map[string]interface{}{
		"username": username,
		"password": password,
	}

	responseData := testPost(client, httpPrefix+"://127.0.0.1:5000/authenticate", "application/json", message)

	if responseData["token"] != nil {
		token := responseData["token"].(string)
		headers["Authorization"] = "Bearer " + token
		testGet(client, httpPrefix+"://127.0.0.1:5000/test", headers)
	}
}

func testEndpoint(w http.ResponseWriter, req *http.Request) {
	if token := req.Context().Value(security.Claims); token != nil {
		json.NewEncoder(w).Encode(token)
	} else {
		fmt.Fprintf(w, `{"message", "not authorized"}`)
	}
}

func testGet(client *http.Client, url string, headers map[string]string) {
	fmt.Println(url)
	req, _ := http.NewRequest("GET", url, nil)
	for k, v := range headers {
		fmt.Println(k, v)
		req.Header.Add(k, v)
	}
	handleResponse(client.Do(req))
}

func testPost(client *http.Client, url string, contentType string, data interface{}) map[string]interface{} {
	fmt.Println(url)
	return handleResponse(postData(client, url, contentType, data))
}

func postData(client *http.Client, url string, contentType string, data interface{}) (resp *http.Response, err error) {
	dataAsBytes, err := json.Marshal(data)

	if err == nil {
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(dataAsBytes))
		req.Header.Add("Content-Type", "application/json")
		resp, err = client.Do(req)
	}

	return
}

func handleResponse(resp *http.Response, err error) (data map[string]interface{}) {
	if err != nil {
		fmt.Println(err)
	} else {
		defer resp.Body.Close()
		fmt.Println(resp.Status)
		body, _ := ioutil.ReadAll(resp.Body)
		if len(body) > 0 {
			fmt.Println(string(body))
			json.Unmarshal(body, &data)
		}
	}
	return
}
