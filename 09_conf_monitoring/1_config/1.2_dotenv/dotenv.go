package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// go run dotenv.go
func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalln(err)
	}

	dbPWD := os.Getenv("DB_PASSWORD")
	fmt.Println("password:", dbPWD)
}
