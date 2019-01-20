package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type User struct {
	UserID string
	Score  int
}

func getUserAPI(w http.ResponseWriter, r *http.Request) {

	user, err := getUser(mux.Vars(r)["user"])
	if err != nil {
		user = &User{
			UserID: mux.Vars(r)["user"],
			Score:  0,
		}
		db.Exec("insert into t_user (user_id, score) values (?, ?)", user.UserID, user.Score)
	}
	err = json.NewEncoder(w).Encode(user)

	if err != nil {
		log.Println("Failed to encode: ", err)
	}

}

func getUser(userID string) (*User, error) {
	userRow := db.QueryRow("select user_id, score from t_user where user_id = ?", userID)

	var result User
	err := userRow.Scan(&result.UserID, &result.Score)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
