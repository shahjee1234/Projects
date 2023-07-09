package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
	//"github.com/denisenkom/go-mssqldb/mssql"
	"github.com/dgrijalva/jwt-go"

	"github.com/gorilla/mux"
)

var jwtKey = []byte("sIoVC8OFOgmxbk9XRYtY2zMKXuYXBGL2d3x1IV37")

var users = map[string]string{
	"user1": "password1",
	"user2": "password2",
}

type Movie struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	UserName  string `json:"username"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func generateToken(username string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour) // Token expires in 24 hours

	claims := &Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func login(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&credentials)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	expectedPassword, ok := users[credentials.Username]
	if !ok || expectedPassword != credentials.Password {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	tokenString, err := generateToken(credentials.Username)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	response := struct {
		Token string `json:"token"`
	}{
		Token: tokenString,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		tokenString := authorizationHeader[len("Bearer "):]

		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				http.Error(w, "Invalid token signature", http.StatusUnauthorized)
				return
			}
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getEmployee(conn *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		rows, err := conn.Query("SELECT * FROM employeedetails")
		if err != nil {
			log.Println("Query failed:", err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		movies := []Movie{}
		for rows.Next() {
			var movie Movie
			err := rows.Scan(&movie.ID, &movie.FirstName, &movie.LastName, &movie.Role, &movie.Email, &movie.UserName, &movie.Password)
			if err != nil {
				log.Println("Scan failed:", err.Error())
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			movies = append(movies, movie)
		}

		if err := rows.Err(); err != nil {
			log.Println("Error occurred while iterating over rows:", err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		//json.NewEncoder(w).Encode(movies)

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200") // Allow requests from any origin
		response, err := json.Marshal(movies)
		if err != nil {
			log.Println("JSON marshaling failed:", err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.Write(response)
	}
}

func getEmployeeByID(w http.ResponseWriter, r *http.Request) {

	c, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			// If the cookie is not set, return an unauthorized status
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// For any other type of error, return a bad request status
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get the JWT string from the cookie
	tknStr := c.Value

	// Initialize a new instance of `Claims`
	claims := &Claims{}

	// Parse the JWT string and store the result in `claims`.
	// Note that we are passing the key in this method as well. This method will return an error
	// if the token is invalid (if it has expired according to the expiry time we set on sign in),
	// or if the signature does not match
	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !tkn.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	params := mux.Vars(r)
	movieID := params["id"]

	connString := "server=DESKTOP-TTBIGOK;user id=mubashir;password=hashmi;port=1435;database=movie;"
	conn, err := sql.Open("mssql", connString)
	if err != nil {
		log.Println("Open connection failed:", err.Error())
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	query := "SELECT ID, FirstName, LastName, Role, Email, UserName, Password FROM employeedetails WHERE id = ?"
	stmt, err := conn.Prepare(query)
	if err != nil {
		log.Println("Failed to prepare SQL statement:", err.Error())
		http.Error(w, "Failed to retrieve movie", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	var movie Movie
	err = stmt.QueryRow(movieID).Scan(&movie.ID, &movie.FirstName, &movie.LastName, &movie.Role, &movie.Email, &movie.UserName, &movie.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Movie not found", http.StatusNotFound)
		} else {
			log.Println("Error retrieving movie:", err.Error())
			http.Error(w, "Failed to retrieve movie", http.StatusInternalServerError)
		}
		return
	}

	response, err := json.Marshal(movie)
	if err != nil {
		log.Println("Error marshaling movie response:", err.Error())
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}
func updateEmployee(w http.ResponseWriter, r *http.Request) {
	// Extract the movie details from the request body
	var movie Movie
	err := json.NewDecoder(r.Body).Decode(&movie)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update the movie in the database
	connString := "server=DESKTOP-TTBIGOK;user id=mubashir;password=hashmi;port=1435;database=movie;"
	conn, err := sql.Open("mssql", connString)
	if err != nil {
		http.Error(w, "Failed to connect to the database", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// Perform the database update
	query := "UPDATE employeedetails SET FirstName=?, LastName=?, Role=?, Email=?, UserName=?, Password=? WHERE ID=?"
	result, err := conn.Exec(query, movie.FirstName, movie.LastName, movie.Role, movie.Email, movie.ID, movie.UserName, movie.Password)
	if err != nil {
		http.Error(w, "Failed to update movie in the database", http.StatusInternalServerError)
		return
	}

	// Check the number of rows affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Failed to get the number of rows affected", http.StatusInternalServerError)
		return
	}

	// Check if any rows were affected
	if rowsAffected == 0 {
		http.Error(w, "Movie not found", http.StatusNotFound)
		return
	}

	// Respond with a success message
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Movie updated successfully")
}

func deleteEmployee(w http.ResponseWriter, r *http.Request) {
	// Extract the movie ID from the request parameters
	vars := mux.Vars(r)
	movieID := vars["id"]

	connString := "server=DESKTOP-TTBIGOK;user id=mubashir;password=hashmi;port=1435;database=movie;"
	conn, err := sql.Open("mssql", connString)
	if err != nil {
		log.Println("Open connection failed:", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	err = conn.Ping() // Test the connection
	if err != nil {
		log.Println("Connection test failed:", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Execute the delete query
	stmt, err := conn.Prepare("DELETE FROM employeedetails WHERE id = ?")
	if err != nil {
		log.Println("Prepare query failed:", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(movieID)
	if err != nil {
		log.Println("Delete query failed:", err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Respond with success message
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Movie deleted successfully"))
}

func createEmployee(w http.ResponseWriter, r *http.Request) {
	var newMovie Movie

	// Decode the JSON request body into a Movie struct
	err := json.NewDecoder(r.Body).Decode(&newMovie)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create a new database connection
	connString := "server=DESKTOP-TTBIGOK;user id=mubashir;password=hashmi;port=1435;database=movie;"
	db, err := sql.Open("mssql", connString)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Test the database connection
	err = db.Ping()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Insert the movie into the database
	stmt, err := db.Prepare("INSERT INTO employeedetails (ID, FirstName,LastName, Role, Email, UserName, Password) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(newMovie.ID, newMovie.FirstName, newMovie.LastName, newMovie.Role, newMovie.Email, newMovie.UserName, newMovie.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the created movie as JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newMovie)
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	// connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s;",
	// 	"DESKTOP-TTBIGOK-", "mubashir", "ahmed", "1435", "movie")
	// conn, err := sql.Open("mssql", connString)
	connString := "server=DESKTOP-TTBIGOK;user id=mubashir;password=hashmi;port=1435;database=movie;"
	conn, err := sql.Open("mssql", connString)
	if err != nil {
		log.Fatal("Open connection failed:", err.Error())
	}
	defer conn.Close()

	err = conn.Ping() // Test the connection
	if err != nil {
		log.Fatal("Connection test failed:", err.Error())
	}
	fmt.Println("Connected to SQL Server")
	// c := cors.New(cors.Options{
	// 	AllowedOrigins:   []string{"http://localhost:4200"}, // Replace with your frontend origin
	// 	AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	// 	AllowedHeaders:   []string{"Origin", "Authorization", "Content-Type"},
	// 	AllowCredentials: true,
	// })
	// http.Handle("/getemployee", c.Handler(http.HandlerFunc(handleAngularURL)))

	r := mux.NewRouter()
	//handler := c.Handler(r)

	//r.Use(c.Handler)

	// Apply the CORS handler to all routes
	//r.Use(cors.AllowAll().Handler)

	// Apply JWT authentication middleware to protected routes
	protected := r.PathPrefix("/").Subrouter()
	protected.Use(authenticate)

	// ...

	r.HandleFunc("/login", login).Methods("POST")
	// ...

	// Apply JWT authentication middleware to protected routes

	protected.Path("/getemployee").Handler(http.HandlerFunc(getEmployee(conn))).Methods("GET")
	//r.HandleFunc("/getemployee", getEmployee(conn)).Methods("GET")
	//protected.HandleFunc("/getemployee", getEmployee(conn)).Methods("GET")
	//r.HandleFunc("/getemployeebyid/{id}", getEmployeeByID).Methods("GET")

	protected.HandleFunc("/getemployeebyid/{id}", getEmployeeByID).Methods("GET")

	//r.HandleFunc("/delete_emp/{id}", deleteEmployee).Methods("DELETE")

	protected.HandleFunc("/delete_emp/{id}", deleteEmployee).Methods("DELETE")
	//r.HandleFunc("/create_emp", createEmployee).Methods("POST")
	protected.HandleFunc("/create_emp", createEmployee).Methods("POST")
	//r.HandleFunc("/update_emp", updateEmployee).Methods("PUT")
	protected.HandleFunc("/update_emp", updateEmployee).Methods("PUT")

	http.HandleFunc("http://localhost:4200", handleAngularURL)

	fmt.Printf("Starting server at port 9000\n")
	//handler := cors.Default().Handler(r)
	// c := cors.New(cors.Options{
	// 	AllowedOrigins:   []string{"*"},
	// 	AllowedHeaders:   []string{"Authorization", "Role", "UserName"},
	// 	AllowCredentials: true,
	// 	Debug:            true,
	// })

	// Insert the middleware
	//handler = c.Handler(handler)
	//log.Fatal(http.ListenAndServe(":9000", r))
	//http.ListenAndServe(":8080", r)

	http.ListenAndServe(":9000", enableCORS(r))
}

func handleAngularURL(w http.ResponseWriter, r *http.Request) {
	// Handle the request from Angular
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
	switch r.Method {
	case http.MethodGet:
		// Handle GET request from Angular
		fmt.Fprintf(w, "This is the response from the Go server for the Angular URL GET request.")
	case http.MethodPost:
		// Handle POST request from Angular
		// You can access the request body and process the data sent by Angular
		fmt.Fprintf(w, "This is the response from the Go server for the Angular URL POST request.")
	default:
		// Return an error for unsupported methods
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	//w.WriteHeader(http.StatusOK)
}
