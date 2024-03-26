package main

import (
	"fmt"
	"net/http"

	"github.com/spf13/viper"
)

// handler yang akan dipanggil setiap kali ada request ke "/"
func helloWorldHandler(w http.ResponseWriter, r *http.Request) {
	// Menulis respons "Hello, World!" ke browser
	myVariable := viper.GetString("DATA_OKE") // Menggunakan nama variable tanpa prefix jika Anda tidak menetapkan SetEnvPrefix
	fmt.Println(myVariable)
	fmt.Fprintf(w, "Hello, World!"+myVariable)
}

func main() {
	// Mengatur handler untuk route "/"
	viper.AutomaticEnv() // Membaca semua environment variables

	http.HandleFunc("/", helloWorldHandler)
	myVariable := viper.GetString("DATA_OKE") // Menggunakan nama variable tanpa prefix jika Anda tidak menetapkan SetEnvPrefix
	fmt.Println(myVariable)
	// Menjalankan server di port 8080
	fmt.Println("Server is running on http://localhost:8099" + myVariable)
	if err := http.ListenAndServe(":8099", nil); err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}
