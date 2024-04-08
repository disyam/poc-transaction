package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/microsoft/go-mssqldb"
	"github.com/oklog/ulid/v2"
)

var db *sql.DB

func debit(id string, value int) {
	var err error
	var tx *sql.Tx
	if tx, err = db.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelReadCommitted}); err != nil {
		log.Fatalln(err)
	}
	defer tx.Rollback()
	var before int
	if err = tx.QueryRow(`SELECT "balance" FROM "wallets" WITH (XLOCK, HOLDLOCK, ROWLOCK, INDEX(0)) WHERE "id" = @p1`, id).Scan(&before); err != nil {
		log.Fatalln(err)
	}
	if before < value {
		err = fmt.Errorf("balance not enough")
		log.Fatalln(err)
	}
	after := before - value
	if _, err = tx.Exec(`UPDATE "wallets" SET "balance" = @p1 WHERE "id" = @p2`, after, id); err != nil {
		log.Fatalln(err)
	}
	if _, err = tx.Exec(`INSERT INTO "ledgers" ("id", "wallet_id", "credit", "debit", "before", "after") VALUES (@p1, @p2, @p3, @p4, @p5, @p6)`, ulid.Make().String(), id, 0, value, before, after); err != nil {
		log.Fatalln(err)
	}
	if err = tx.Commit(); err != nil {
		log.Fatalln(err)
	}
	fmt.Println()
	fmt.Println("id: ", id)
	fmt.Println("before: ", before)
	fmt.Println("debit: ", value)
	fmt.Println("after: ", after)
}

func main() {
	var err error
	if db, err = sql.Open("sqlserver", "sqlserver://sa:R00Tmssql@127.0.0.1:1433?database=poc-transaction"); err != nil {
		return
	}
	defer db.Close()
	if _, err = db.Exec(`DROP TABLE IF EXISTS "ledgers"`); err != nil {
		log.Fatalln(err)
	}
	if _, err = db.Exec(`DROP TABLE IF EXISTS "wallets"`); err != nil {
		log.Fatalln(err)
	}
	if _, err = db.Exec(`CREATE TABLE "wallets" ("id" varchar(255) NOT NULL PRIMARY KEY, "balance" int NOT NULL)`); err != nil {
		log.Fatalln(err)
	}
	if _, err = db.Exec(`CREATE TABLE "ledgers" ("id" varchar(255) NOT NULL PRIMARY KEY, "wallet_id" varchar(255) NOT NULL REFERENCES "wallets", "credit" int NOT NULL, "debit" int NOT NULL, "before" int NOT NULL, "after" int NOT NULL)`); err != nil {
		log.Fatalln(err)
	}
	if _, err = db.Exec(`INSERT INTO "wallets" ("id", "balance") VALUES (@p1, @p2)`, "1", 1000); err != nil {
		log.Fatalln(err)
	}
	if _, err = db.Exec(`INSERT INTO "ledgers" ("id", "wallet_id", "credit", "debit", "before", "after") VALUES (@p1, @p2, @p3, @p4, @p5, @p6)`, ulid.Make().String(), 1, 1000, 0, 0, 1000); err != nil {
		log.Fatalln(err)
	}
	if _, err = db.Exec(`INSERT INTO "wallets" ("id", "balance") VALUES (@p1, @p2)`, "2", 2000); err != nil {
		log.Fatalln(err)
	}
	if _, err = db.Exec(`INSERT INTO "ledgers" ("id", "wallet_id", "credit", "debit", "before", "after") VALUES (@p1, @p2, @p3, @p4, @p5, @p6)`, ulid.Make().String(), 2, 2000, 0, 0, 2000); err != nil {
		log.Fatalln(err)
	}
	count := 20
	go func() {
		for range count {
			go debit("1", 20)
		}
	}()
	go func() {
		for range count {
			go debit("1", 30)
		}
	}()
	go func() {
		for range count {
			go debit("2", 40)
		}
	}()
	go func() {
		for range count {
			go debit("2", 60)
		}
	}()
	time.Sleep(3 * time.Second)
	var id string
	var balance int
	if err = db.QueryRow(`SELECT "id", "balance" FROM "wallets" WHERE "id" = @p1`, "1").Scan(&id, &balance); err != nil {
		log.Fatalln(err)
	}
	fmt.Println()
	fmt.Printf("id: %+v\n", id)
	fmt.Printf("balance: %+v\n", balance)
	if err = db.QueryRow(`SELECT "id", "balance" FROM "wallets" WHERE "id" = @p1`, "2").Scan(&id, &balance); err != nil {
		log.Fatalln(err)
	}
	fmt.Println()
	fmt.Printf("id: %+v\n", id)
	fmt.Printf("balance: %+v\n", balance)
}
