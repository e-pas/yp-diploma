package main

import (
	"log"

	"yp-diploma/internal/app"
)

func main() {
	app := app.New()

	err := app.Run()
	if err != nil {
		log.Println(err)
	}
}
