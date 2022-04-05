package postgres

import (
	"errors"
	"fmt"
	"strings"
)

type User struct {
	ID          int
	Username    string
	Name        string
	Surname     string
	Description string
}

// The function returns the User ID of the username
// -1 if the user does not exist
func exist(username string) int {
	username = strings.ToLower(username)

	db, err := openConnection()
	if err != nil {
		fmt.Println(err)
		return -1
	}
	defer db.Close()

	userId := -1
	stmt := fmt.Sprintf(`SELECT "id" FROM "users" where username = '%s'`, username)
	rows, err := db.Query(stmt)

	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		if err != nil {
			fmt.Println("Scan", err)
			return -1
		}
		userId = id
	}
	defer rows.Close()

	return userId
}

// AddUser adds a new user to the database
// Returns new User ID
// -1 if there was an error
func AddUser(u User) int {
	u.Username = strings.ToLower(u.Username)

	db, err := openConnection()
	if err != nil {
		fmt.Println("AddUser:openConnection:", err)
		return -1
	}
	defer db.Close()

	userId := exist(u.Username)
	if userId != -1 {
		fmt.Println("User already exists:", Username)
		return -1
	}

	// This is how construct a query that accepts parameters
	insertStmt := `INSERT INTO "users" ("username") VALUES ($1)`
	// This is how you pass the desired value into the insertStmt
	_, err = db.Exec(insertStmt, u.Username)
	if err != nil {
		fmt.Println("AddUser:insert user:", err)
		return -1
	}

	userId = exist(u.Username)
	if userId == -1 {
		return userId
	}

	insertStmt = `INSERT INTO "userdata" ("userid", "name", "surname", "description") VALUES ($1, $2, $3, $4)`
	_, err = db.Exec(insertStmt, userId, u.Name, u.Surname, u.Description)
	if err != nil {
		fmt.Println("AddUser:insert userdata:", err)
		return -1
	}

	return userId
}

// DeleteUser deletes an existing user
func DeleteUser(id int) error {
	db, err := openConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	// Does the ID exists?
	stmt := fmt.Sprintf(`SELECT "username" FROM "users" WHERE id = %d`, id)
	rows, err := db.Query(stmt)
	var username string
	for rows.Next() {
		err = rows.Scan(&username)
		if err != nil {
			return err
		}
	}
	defer rows.Close()

	// double check
	if exist(username) != id {
		return fmt.Errorf("User with ID %d does not exits", id)
	}

	// Delete from userdata
	delStmt := `DELETE FROM "userdata" WHERE userid=$1`
	_, err = db.Exec(delStmt, id)
	if err != nil {
		return err
	}

	// Delete from users
	delStmt = `DELETE FROM "users" WHERE id=$1`
	_, err = db.Exec(delStmt, id)
	if err != nil {
		return err
	}

	return nil
}

// ListUsers is to list all users
func ListUsers() ([]User, error) {
	users := []User{}
	db, err := openConnection()
	if err != nil {
		return users, err
	}
	defer db.Close()

	rows, err := db.Query(`SELECT "id", "username", "name", surname", "description" 
	FROM "users", "userdata" 
	WHERE users.id = userdata.userid`)
	if err != nil {
		return users, fmt.Errorf("ListUsers: %s", err)
	}

	for rows.Next() {
		var id int
		var username string
		var name string
		var surname string
		var description string
		err = rows.Scan(&id, &username, &name, &surname, &description)
		if err != nil {
			return users, err
		}
		temp := User{ID: id, Username: username, Name: name, Surname: surname, Description: description}
		users = append(users, temp)
	}
	defer rows.Close()
	return users, nil
}

// UpdateUser is for updating an existing user
func UpdateUser(u User) error {
	db, err := openConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	// Check if the given user exists
	userId := exist(u.Username)
	if userId == -1 {
		return errors.New("User does not exist")
	}

	u.ID = userId
	updateStmt := `UPDATE "userdata" SET "name"=$1, "surname"=$2, description=$3 
	WHERE "userid"=$4`
	_, err = db.Exec(updateStmt, u.Name, u.Surname, u.Description, u.ID)
	if err != nil {
		return err
	}

	return nil
}
