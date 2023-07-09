package simplecrud




import (
 "database/sql"
 "fmt"
 "math/rand"
 "time"
)

type MyEmployee struct {
	EmpId     int     `json:"empId"`
	EmpName   string  `json:"empName"`
	EmpSalary float64 `json:"empSalary"`
	Email     string  `json:"email"`
}
func main(){
	func ExecuteCrud(db *sql.DB) error {

		// Call to insert
		if err := insert(db); err != nil {
		   return err
		}
	   
		if err := update(db); err != nil {
		   return err
		}
	   
		if err := list(db); err != nil {
		   return err
		}
	   
		if err := delete(db); err != nil {
		   return err
		}
	   
		return nil
	   }
	   
	   // DELETE
	   func delete(db *sql.DB) error {
		  fmt.Println("On delete function")
		 
		  stmt, err := db.Prepare("DELETE FROM users WHERE id=?")
		  if err != nil {
			 fmt.Println(err)
			 return err
		  }
		  res, err := stmt.Exec(123456) // hardcoded id
		  if err != nil {
			 fmt.Println(err)
			 return err
		  }
		 
		  count, err := res.RowsAffected()
		  if err != nil {
			 fmt.Println(err)
			 return err
		  }
		 
		  fmt.Println("Rows affected: ", count)
		 
		  return nil
	   }
	   
	   // UPDATE
	   func update(db *sql.DB) error {
		  fmt.Println("On Update function")
		 
		  stmt, err := db.Prepare("UPDATE users SET name = ? WHERE id = ?")
		  if err != nil {
			 fmt.Println(err)
			 return err
		  }
		 
		  // hardcoded id
		  res, err := stmt.Exec("edited name", 81)
		  if err != nil {
			 fmt.Println(err)
			 return err
		  }
		 
		  count, err := res.RowsAffected()
		  if err != nil {
			 fmt.Println(err)
			 return err
		  }
		 
		  fmt.Println("Rows affected: ", count)
		 
		  return nil
	   }
	   
	   // SELECT ALL
	   func list(db *sql.DB) error {
		  fmt.Println("On list function")
		 
		  var allUsers []MyEmployee
		  rows, err := db.Query("SELECT * FROM users;")
		  if err != nil {
			 fmt.Println(err)
			 return err
		  }
		 
		  defer rows.Close()
		 
		  for rows.Next() {
			 var resultUser MyEmployee
			 err := rows.Scan(&resultUser.EmpId, &resultUser.EmpName, &resultUser.EmpSalary, &resultUser.Email)
			 if err != nil {
				return err
			 }
		   
			 allUsers = append(allUsers, resultUser)
		  }
		 
		  fmt.Println("Employee :", allUsers)
		 
		  return nil
	   }
	   
	   // INSERT
	   func insert(db *sql.DB) error {
		  fmt.Println("On insert function")
		 
		  stmt, err := db.Prepare("INSERT INTO users (empId, empName, empSalary, email) VALUES (?,?,?,?)")
		  if err != nil {
			 fmt.Println(err)
			 return err
		  }
		 
		  newUser := MyEmployee{
			 EmpId:       rand.Intn(1000) + time.Now().Nanosecond()*2,
			 EmpName:     "Jane",
			 EmpSalary: 78000,
			 Email:      "jane@gmail.com",
		  }
		  res, err := stmt.Exec(newUser.EmpId, newUser.EmpName, newUser.EmpSalary, newUser.Email)
		  if err != nil {
			 fmt.Println(err)
			 return err
		  }
		 
		  count, err := res.RowsAffected()
		  if err != nil {
			 fmt.Println(err)
			 return err
		  }
		 
		  fmt.Println("Rows affected: ", count)
		 
		  return nil
	   }
	

}

