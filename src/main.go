package main

import (
	"log"
	"os"
)

func main() {
	file, err := os.Open("script.txt")
	if err != nil {
		log.Println(err)
	}

	view, err := NewView(file)
	if err != nil {
		log.Fatal(err)
	}
	defer view.Close()

	view.Title()

	quit := make(chan struct{})
	go view.Loop(quit)
	<-quit
}
