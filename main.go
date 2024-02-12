package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func init() {
	connectDB()
}

type transaction struct {
	Client_id int
	Valor     int
	Tipo      string
	Descricao string
}

type trasactionResponse struct {
	Limite int `json:"limite"`
	Saldo  int `json:"saldo"`
}

func main() {
	port := fmt.Sprintf(":%s", os.Getenv("PORT"))
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("alive"))
	})

	r.Post("/clientes/{id}/transacoes", handleTrasactions)
	fmt.Printf("Running on port %v", port)

	http.ListenAndServe(port, r)

}

func handleTrasactions(w http.ResponseWriter, r *http.Request) {

	id, idErr := strconv.Atoi(chi.URLParam(r, "id"))

	if idErr != nil {
		w.WriteHeader(422)
		w.Write([]byte("ID should be an int"))
		return
	}

	var t transaction

	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		w.WriteHeader(422)
		w.Write([]byte("Invalid values."))
		return
	}

	if len(t.Descricao) > 10 || t.Descricao == "" {
		w.WriteHeader(422)
		w.Write([]byte("Descricao cannot be empty or greater than 10 characters"))
		return
	}

	if t.Tipo != "c" && t.Tipo != "d" {
		w.WriteHeader(422)
		w.Write([]byte("Wrong transaction type"))
		return
	}

	if t.Tipo == "c" {
		saldo, err := getSaldo(id)
		fmt.Println(saldo)
		if err != nil {
			fmt.Println(err)
			return
		}
		saldo = saldo + t.Valor
		fmt.Println(saldo)

		limite, err := getLimite(id)
		if err != nil {
			fmt.Println(err)
			return
		}

		res := trasactionResponse{
			Limite: limite,
			Saldo:  saldo,
		}

		t.Client_id = id
		newTransaction := createTransaction(t)
		if newTransaction != nil {
			w.WriteHeader(422)
			w.Write([]byte("Somthing went wrong"))
			return
		}

		errSaldo := updateSaldo(saldo, id)
		if errSaldo != nil {
			w.WriteHeader(422)
			w.Write([]byte("Somthing went wrong"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(res)
		return
	}

	// fmt.Println(id)
	// fmt.Println(t)
	// w.Write([]byte("alive"))

}
