package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

type Wallet struct {
	ID      int
	Balance int
}

type Ledger struct {
	ID       int
	WalletID int
	Wallet   Wallet
	Credit   int
	Debit    int
	Balance  int
}

var db *gorm.DB

func purchase(walletID int, nominal int) {
	var err error
	tx := db.Begin()
	var wallet Wallet
	if err = tx.Clauses(clause.Locking{
		Strength: "UPDATE",
	}).Find(&wallet, walletID).Error; err != nil {
		log.Fatalln(err)
	}
	before := wallet.Balance
	after := before
	if wallet.Balance >= nominal {
		after = before - nominal
		if err = tx.Model(&wallet).Update("balance", after).Error; err != nil {
			log.Fatalln(err)
		}
		if err = tx.Create(&Ledger{
			WalletID: walletID,
			Debit:    nominal,
			Balance:  after,
		}).Error; err != nil {
			log.Fatalln(err)
		}
		tx.Commit()
		fmt.Println("\nSUCCESS")
	} else {
		tx.Rollback()
		fmt.Println("\nFAILED")
	}
	fmt.Println("wallet: ", walletID)
	fmt.Println("before: ", before)
	fmt.Println("debit: ", nominal)
	fmt.Println("after: ", after)
}

func main() {
	var err error
	dsn := "host=127.0.0.1 user=postgres password=R00Tpostgres dbname=database port=5432"
	if db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             1 * time.Second,
				LogLevel:                  logger.Warn,
				IgnoreRecordNotFoundError: false,
				Colorful:                  true,
			}),
	}); err != nil {
		log.Fatalln(err)
	}
	if err = db.AutoMigrate(&Wallet{}, &Ledger{}); err != nil {
		log.Fatalln(err)
	}
	if err = db.Where("1 = 1").Delete(&Ledger{}).Error; err != nil {
		log.Fatalln(err)
	}
	if err = db.Where("1 = 1").Delete(&Wallet{}).Error; err != nil {
		log.Fatalln(err)
	}
	if err = db.Create(&Wallet{
		ID:      1,
		Balance: 1000,
	}).Error; err != nil {
		log.Fatalln(err)
	}
	if err = db.Create(&Ledger{
		WalletID: 1,
		Credit:   1000,
		Balance:  1000,
	}).Error; err != nil {
		log.Fatalln(err)
	}
	if err = db.Create(&Wallet{
		ID:      2,
		Balance: 2000,
	}).Error; err != nil {
		log.Fatalln(err)
	}
	if err = db.Create(&Ledger{
		WalletID: 2,
		Credit:   2000,
		Balance:  2000,
	}).Error; err != nil {
		log.Fatalln(err)
	}
	count := 20
	go func() {
		for i := 0; i < count; i++ {
			go purchase(1, 10)
		}
	}()
	go func() {
		for i := 0; i < count; i++ {
			go purchase(1, 20)
		}
	}()
	go func() {
		for i := 0; i < count; i++ {
			go purchase(1, 30)
		}
	}()
	go func() {
		for i := 0; i < count; i++ {
			go purchase(2, 40)
		}
	}()
	go func() {
		for i := 0; i < count; i++ {
			go purchase(2, 50)
		}
	}()
	time.Sleep(10 * time.Second)
}
