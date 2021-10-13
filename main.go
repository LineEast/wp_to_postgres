package main

import (
	"fmt"
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

type Base struct {
	Quiz_id int "json:'id'"
}

func main() {
	db_mySql, err := sql.Open("mysql", "root:1337228@tcp(127.0.0.1:3306)/nuancesprog")
	if err != nil {
		panic(err)
	}

	sel, err := db_mySql.Query("select quiz_id from wp_fca_qc_activity_tbl")
	if err != nil {
		panic(err)
	}

	for sel.Next() {
		var base Base
		err = sel.Scan(&base.Quiz_id)
		if err != nil {
			panic(err)
		}

		fmt.Println(fmt.Sprintf("quiz_id: %d", base.Quiz_id))
	}

	//fmt.Println(sel)
	fmt.Println("Nice")

	sel.Close()
	db_mySql.Close()
}
