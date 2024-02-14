package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

var db *sql.DB

func connectDB() {
	var err error

	connStr := "postgres://admin:123@localhost/rinha?sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	// defer db.Close()

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println("Connected!")
}

func getSaldo(id int) (int, error) {

	var saldo int
	err := db.QueryRow("SELECT valor FROM saldos WHERE cliente_id=$1", id).Scan(&saldo)
	if err != nil {
		fmt.Printf("error getting saldo: %v\n", err)
		return 0, err
	}
	return saldo, nil

}

func getLimite(id int) (int, error) {
	var limite int
	err := db.QueryRow("SELECT limite FROM clientes WHERE id=$1;", id).Scan(&limite)
	if err != nil {
		fmt.Printf("error getting limite: %v\n", err)
		return 0, err
	}
	return limite, nil
}

func createTransaction(transac transaction) error {
	_, err := db.Query("INSERT INTO transacoes (cliente_id, valor, tipo, descricao) VALUES ($1, $2, $3, $4);", transac.Client_id, transac.Valor, transac.Tipo, transac.Descricao)
	if err != nil {
		fmt.Printf("error creating transaction: %v\n", err)
		return err
	}
	return nil
}

func updateSaldo(saldo int, cliente_id int) error {
	fmt.Println(cliente_id)
	_, err := db.Query("UPDATE saldos SET valor = $1 WHERE cliente_id = $2;", saldo, cliente_id)
	if err != nil {
		fmt.Printf("error updating saldo: %v\n", err)
		return err
	}
	return nil
}

func getTransactions(cliente_id int) ([]Transacao, error) {
	rows, err := db.Query("SELECT valor, tipo, descricao, realizada_em FROM transacoes WHERE cliente_id = $1 ORDER BY realizada_em DESC LIMIT 10;", cliente_id)
	if err != nil {
		fmt.Printf("Error when getting transacoes: %v\n", err)
		return nil, err
	}
	var transacoes []Transacao
	for rows.Next() {
		var transaction Transacao
		if err := rows.Scan(&transaction.Valor, &transaction.Tipo, &transaction.Descricao, &transaction.Realizada_em); err != nil {
			fmt.Printf("Somthing went wrong when serializing rows: %v", err)
			return nil, err
		}
		transacoes = append(transacoes, transaction)
	}

	return transacoes, nil
}
