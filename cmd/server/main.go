package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	server := http.NewServeMux()

	server.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s, %s, %s, %v\n", r.Proto, r.Host, r.URL, r.Header)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(209)
		w.Write([]byte(`{ "hello" : 1.5 , "good" : true }`))
	})

	fmt.Println("Running server at '127.0.0.1:4007'")

	err := http.ListenAndServe("127.0.0.1:4007", server)
	if err != nil {
		log.Fatal(err)
	}
}
