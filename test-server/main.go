package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	data_dir_path := os.Args[1]
	log.Fatal(http.ListenAndServe(":8080", http.FileServer(http.Dir(data_dir_path))))
}
