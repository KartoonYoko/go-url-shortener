package main

import (
	"fmt"

	"github.com/KartoonYoko/go-url-shortener/internal/app"
)

func main() {
	app.Serve()

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Got panic", r)
			panic("Something gone wrong")
		}
	}()
}
