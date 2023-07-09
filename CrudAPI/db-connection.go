// package main

// import "gorm.io/gorm"

// var DataBase *gorm.DB
// var urlDSN =

// func DataMigration() {

// }


package main

import (
    "database/sql"
    "fmt"

    _ "github.com/denisenkom/go-mssqldb"
	"github.com/josue/database_connections/internal/simplecrud"

)

func main() {


	
    // Set up the connection string
    server := "localhost"
    port := 1433
    user := "mubashir"
    password := "ahmed"
    database := "employee"
    connectionString := fmt.Sprintf("server=%s;port=%d;user id=%s;password=%s;database=%s;",
        server, port, user, password, database)

    // Connect to the database
    db, err := sql.Open("sqlserver", connectionString)
    if err != nil {
        panic(err.Error())
    }

    // Ping the database to check the connection
    err = db.Ping()
    if err != nil {
        panic(err.Error())
    }

    // Query the database
    rows, err := db.Query("SELECT * FROM employees")
    if err != nil {
        panic(err.Error())
    }
    defer rows.Close()

    // Loop through the results and print them
    // for rows.Next() {
    //     var EmpId int
    //     var EmpName string
    //     err := rows.Scan(&EmpId, &EmpName)
    //     if err != nil {
    //         panic(err.Error())
    //     }
    //     fmt.Printf("EmpId: %d, EmpName: %s\n", EmpId, EmpName)
    // }

	errsc := simplecrud.ExecuteCrud(db)

 if errsc != nil {
  fmt.Println(errsc)
 }

}

