package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var OTPGEN string

// Global variable for the database connection pool
var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("mysql", "root:dexter@tcp(127.0.0.1:3306)/mykeys")
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return // Corrected: Return after a connection error
	}
	defer db.Close()

	// Create a handler that serves files from the "static" directory
	fs := http.FileServer(http.Dir("build"))
	http.Handle("/", fs)
	// Register the file server and form handler
	http.HandleFunc("/submit-form", formHandler)
	http.HandleFunc("/validate-no", validN)
	http.HandleFunc("/otp-val", valOTP)
	http.HandleFunc("/changemypasswordreq", changereq)

	fmt.Println("Server listening on http://localhost:8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
func valOTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	i1 := r.FormValue("OTPno")
	if i1 == OTPGEN {
		w.Write([]byte("1"))
	} else {
		w.Write([]byte("bad OTP"))
	}

}
func changereq(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	phoneNO := r.FormValue("Pno")
	pass := r.FormValue("newpassword")

	// use Exec, not Query
	_, err := db.Exec(`UPDATE userpass SET password = ? WHERE pNO = ?`, pass, phoneNO)
	if err != nil {
		http.Error(w, "DB update failed", http.StatusInternalServerError)
		return
	}

	row := db.QueryRow(`SELECT password FROM userpass WHERE pNO = ?`, phoneNO)

	var passgot string
	err = row.Scan(&passgot)
	if err != nil {
		http.Error(w, "Failed to read password", http.StatusInternalServerError)
		return
	}

	if pass == passgot {
		w.Write([]byte("1"))
	} else {
		w.Write([]byte("-1"))
	}
}

func validN(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	i2 := r.FormValue("FPNo")
	rows1, err := db.Query("SELECT * from userpass where pNO =? ", i2)
	if err != nil {
		fmt.Println("Error selecting userpass:", err)
	}
	if !rows1.Next() {
		w.Write([]byte("-1"))
		return
	} else {
		w.Write([]byte("1"))
		OTPGEN = ""
		rand.NewSource(time.Now().UnixNano())
		n := rand.Intn(100000)
		OTPGEN += strconv.Itoa(n)
		println(OTPGEN)
	}
	defer rows1.Close()
}

func formHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	input1 := r.FormValue("in1")
	input2 := r.FormValue("in2")

	// Corrected: Use a prepared statement to prevent SQL Injection
	rows, err := db.Query(`SELECT password FROM userpass WHERE username=?`, input1)
	if err != nil {
		http.Error(w, "Failed to fetch data from database", http.StatusInternalServerError)
		return // Corrected: Return after an error
	}

	// Corrected: Ensure rows are closed to prevent resource leaks
	defer rows.Close()

	for rows.Next() {
		var pass string
		if err := rows.Scan(&pass); err != nil {
			http.Error(w, "Error reading data.", http.StatusInternalServerError)
			return
		}

		if pass == input2 {
			fmt.Fprint(w, "<p>Your password is correct.</p>")
			return // Return after a successful login
		}
	}

	// This message is only reached if no matching username/password pair was found
	fmt.Fprint(w, "<p>Invalid password.</p>")
}
