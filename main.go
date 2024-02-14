package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

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

type Saldo_extrato struct {
	Total        int       `json:"total"`
	Data_extrato time.Time `json:"data_extrato"`
	Limite       int       `json:"limite"`
}

type Transacao struct {
	Valor        int    `json:"valor"`
	Tipo         string `json:"tipo"`
	Descricao    string `json:"descricao"`
	Realizada_em string `json:"realizada_em"`
}

type Transacoes []Transacao

type extrato_response struct {
	Saldo_extrato `json:"saldo"`
	Transacoes    `json:"ultimas_transacoes"`
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
	r.Post("/clientes/{id}/extrato", handleExtrato)
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

	limite, limiteErr := getLimite(id)

	if limiteErr != nil {
		w.WriteHeader(404)
		w.Write([]byte("User not found"))
		return
	}

	var t transaction

	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		w.WriteHeader(422)
		w.Write([]byte("Invalid values."))
		return
	}

	t.Client_id = id

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

		res := trasactionResponse{
			Limite: limite,
			Saldo:  saldo,
		}

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

	if t.Tipo == "d" {
		saldo, err := getSaldo(id)
		fmt.Println(saldo)
		if err != nil {
			fmt.Println(err)
			return
		}

		totalCredit := saldo + limite
		fmt.Println(saldo)

		if totalCredit >= t.Valor {
			saldo = saldo - t.Valor
			fmt.Println(saldo)
		} else {
			w.WriteHeader(422)
			w.Write([]byte("Insuficient Funds"))
			return
		}

		res := trasactionResponse{
			Limite: limite,
			Saldo:  saldo,
		}

		newTransaction := createTransaction(t)
		if newTransaction != nil {
			w.WriteHeader(422)
			w.Write([]byte("Somthing went wrong"))
			return
		}

		fmt.Println(saldo)

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
}

func handleExtrato(w http.ResponseWriter, r *http.Request) {
	id, idErr := strconv.Atoi(chi.URLParam(r, "id"))

	if idErr != nil {
		w.WriteHeader(422)
		w.Write([]byte("ID should be an int"))
		return
	}

	limite, limiteErr := getLimite(id)

	if limiteErr != nil {
		w.WriteHeader(404)
		w.Write([]byte("User not found"))
		return
	}

	saldo, saldoerr := getSaldo(id)

	if saldoerr != nil {
		w.WriteHeader(422)
		w.Write([]byte("something went wrong"))
		return
	}

	saldot := Saldo_extrato{
		Total:        saldo,
		Data_extrato: time.Now(),
		Limite:       limite,
	}

	transacoes, trerr := getTransactions(id)

	if trerr != nil {
		w.WriteHeader(422)
		w.Write([]byte("Something went wrong when getting transactions"))
		return
	}

	res := extrato_response{
		Saldo_extrato: saldot,
		Transacoes:    transacoes,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(res)
	return

}
