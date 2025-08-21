package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/marcboeker/go-duckdb"
)

func main() {
	db, _ := sql.Open("duckdb", "")
	defer db.Close()

	rows, err := db.Query(
		`SELECT amount
		   FROM read_csv_auto('statistics_experiment_data.csv', delim=',')`,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var prices []float64
	for rows.Next() {
		var p float64
		if err := rows.Scan(&p); err != nil {
			log.Fatal(err)
		}
		prices = append(prices, p)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("所有单价：", prices)

}
