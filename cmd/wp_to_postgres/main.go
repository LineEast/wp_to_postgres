package main

import (
	"wp_to_postgres/internal/base"
)

func main() {
	dbMySql, dbPostgres := base.StartBase()
	base.AllInfo(dbMySql, dbPostgres)
	
	dbPostgres.Close()
	dbMySql.Close()
}