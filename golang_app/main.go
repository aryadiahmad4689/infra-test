package main

import (
	"fmt"
	"net/http"

	"os"

	"github.com/spf13/viper"
)

// handler yang akan dipanggil setiap kali ada request ke "/"
func helloWorldHandler(w http.ResponseWriter, r *http.Request) {
	// Menulis respons "Hello, World!" ke browser

	fmt.Fprintf(w, "Hello, World!")
}

func main() {
	// Mengatur handler untuk route "/"
	viper.AutomaticEnv() // Membaca semua environment variables

	http.HandleFunc("/", helloWorldHandler)

	fmt.Println("Testing", viper.GetString("DATA_OKE"))
	fmt.Println("Testing", os.Getenv("DATA_OKE"))
	fmt.Println("Testing", os.Getenv("DATA_OKE"))

	// Menjalankan server di port 8080
	fmt.Println("Server is running on http://localhost:8099")
	if err := http.ListenAndServe(":8099", nil); err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}
