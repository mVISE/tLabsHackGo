package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Item struct {
	ItemID      string
	Questions   []Question
	Description string
	ImageRef    string
	value       int
	userID      string
	locked      bool
}

type notFoundError struct {
}

func (e *notFoundError) Error() string {
	return "Not Found"
}

// If item is already locked by another user, return 400
// If item is already locked by this user, normal response
// If item is not locked and owned by a diff user, change owner and normal response
// If item is not locked and owned by this user, return 400 (they have already earned points)
func getItemAPI(w http.ResponseWriter, r *http.Request) {
	item, err := getItem(mux.Vars(r)["item"])
	if err != nil {
		w.WriteHeader(404)
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte("Item not found"))

		if err != nil {
			log.Printf("Unable to write response body")
		}
		return
	}

	// If user has already submitted
	if item.userID == mux.Vars(r)["user"] && !item.locked {
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte("User has already submitted this item"))

		if err != nil {
			log.Printf("Unable to write response body")
		}
		return
	}

	// Locked by another user
	if item.userID != mux.Vars(r)["user"] {
		if item.locked {
			w.WriteHeader(400)
			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write([]byte("Item locked by another user"))

			if err != nil {
				log.Printf("Unable to write response body")
			}
			return
		} else {
			_, err = db.Exec("update t_item set locked = 1, user_id = ? where item_id = ?", mux.Vars(r)["user"], item.ItemID)
			if err != nil {
				log.Println("Unable to reassign and lock item: ", err)
			}
		}
	}

	err = json.NewEncoder(w).Encode(item)

	if err != nil {
		log.Println("Failed to encode: ", err)
	}
}

func getItem(itemID string) (*Item, error) {
	rows, err := db.Query("SELECT i.item_id, i.value, i.user_id, i.locked, i.description, i.image_ref, q.question_id, q.question, q.answer FROM t_item i INNER JOIN t_questions q ON q.item_id = i.item_id WHERE i.item_id = ?", itemID)

	if err != nil {
		log.Println("Can't read item: ", err)
		return nil, err
	}

	var result Item
	var cur Question
	ok := false
	for rows.Next() {
		ok = true
		err = rows.Scan(&result.ItemID, &result.value, &result.userID, &result.locked, &result.Description, &result.ImageRef, &cur.ID, &cur.Question, &cur.answer)
		if err != nil {
			log.Println("Failed to read row: ", err)
		}
		result.Questions = append(result.Questions, cur)
	}

	if !ok {
		return nil, &notFoundError{}
	}

	return &result, nil
}

func getUserItems(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT item_id from t_item WHERE user_id = ?", mux.Vars(r)["user"])
	if err != nil {
		log.Println(err)
	}
	var result []Item
	var itemID string
	for rows.Next() {
		rows.Scan(&itemID)

		item, _ := getItem(itemID)
		result = append(result, *item)
	}

	err = json.NewEncoder(w).Encode(result)

	if err != nil {
		log.Println("Failed to encode: ", err)
	}

}
