package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type Book struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
}

var books = make(map[string]Book)

var respChan = make(chan []byte)

var errChan = make(chan []byte)

func main() {
	http.HandleFunc("/books", handleBooks)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleBooks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		go getAllBooks(respChan, errChan)
		sendResponse(w)
	case http.MethodPost:
		go addBook(r, respChan, errChan)
		sendResponse(w)
	case http.MethodDelete:
		go deleteBook(r, respChan, errChan)
		sendResponse(w)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func addBook(r *http.Request, respChan, errChan chan<- []byte) {
	var book Book
	err := json.NewDecoder(r.Body).Decode(&book)
	if err != nil {
		log.Println("Error decoding request:", err)
		errChan <- []byte(err.Error())
		return
	}

	books[book.ID] = book
	response, err := json.Marshal(books[book.ID])
	if err != nil {
		log.Println("Error marshaling response:", err)
		errChan <- []byte(err.Error())
		return
	}

	respChan <- response
}

func getAllBooks(respChan, errChan chan<- []byte) {
	response, err := json.Marshal(books)
	if err != nil {
		log.Println("Error marshaling response:", err)
		errChan <- []byte(err.Error())
		return
	}

	respChan <- response
}

func deleteBook(r *http.Request, respChan, errChan chan<- []byte) {
	id := r.URL.Query().Get("id")

	delete(books, id)

	response, err := json.Marshal(books)
	if err != nil {
		log.Println("Error marshaling response:", err)
		errChan <- []byte(err.Error())
		return
	}

	respChan <- response
}

func sendResponse(w http.ResponseWriter) {
	select {
	case resp := <-respChan:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	case err := <-errChan:
		http.Error(w, string(err), http.StatusInternalServerError)
	}
}