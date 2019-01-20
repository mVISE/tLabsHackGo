package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

type Question struct {
	ID       string
	Question string
	answer   bool
}

type AnswerPost struct {
	Answers   map[string]bool
	Signature string
	UserID    string
}

type Transaction struct {
	UserID    string
	ItemID    string
	Value     int
	Timestamp time.Time
	Signature string
}

func postAnswer(w http.ResponseWriter, r *http.Request) {
	var answerPost AnswerPost

	// Get AnswerPost from body
	err := json.NewDecoder(r.Body).Decode(&answerPost)
	if err != nil {
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte("Unable to decode answer body"))

		if err != nil {
			log.Printf("Unable to write response body")
		}
		return
	}

	user, err := getUser(answerPost.UserID)
	if err != nil {
		w.WriteHeader(400)
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte("User doesn't exist"))

		if err != nil {
			log.Printf("Unable to write response body")
		}
		return
	}

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

	if len(item.Questions) == 0 {
		w.WriteHeader(500)
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte("Invalid item: no associated questions"))

		if err != nil {
			log.Printf("Unable to write response body")
		}
		return
	}

	// check all answers
	for _, question := range item.Questions {
		answer, ok := answerPost.Answers[question.ID]
		if !ok || answer != question.answer {
			log.Println("User failed a question. ID: " + question.ID)
			json.NewEncoder(w).Encode(user)
			return
		}
	}

	log.Println("User answered successfully, incrementing score")
	user.Score += item.value
	// save users new score
	_, err = db.Exec("update t_user set score = ? where user_id = ?", user.Score, user.UserID)

	if err != nil {
		log.Println("Failed to update user's score: ", err)
	}

	// Unlock item
	_, err = db.Exec("update t_item set locked = 0 where item_id = ?", item.ItemID)

	if err != nil {
		log.Println("Failed to unlock item: ", err)
	}

	postTransaction(*item, *user, answerPost.Signature)

	json.NewEncoder(w).Encode(user)
}

func postTransaction(item Item, user User, signature string) {
	transaction := Transaction{
		UserID:    user.UserID,
		ItemID:    item.ItemID,
		Value:     item.value,
		Timestamp: time.Now(),
		Signature: signature,
	}

	_, err := db.Exec("insert into t_transaction (user_id, item_id, value, timestamp, signature) values (?, ?, ?, ?, ?)",
		transaction.UserID,
		transaction.ItemID,
		transaction.Value,
		transaction.Timestamp.Format("2006-01-02 15:04:05.999999"),
		transaction.Signature,
	)

	if err != nil {
		log.Println("Failed to save transaction: ", err)
	}

	// Create JSON representing hash
	json, err := json.Marshal(transaction)

	// Hash JSON formatted transaction
	hasher := sha256.New()
	hasher.Write(json)

	hashString := hex.EncodeToString(hasher.Sum(nil))

	// Post transaction to blockchain
	req, err := http.NewRequest("POST", "https://developers.cryptowerk.com/platform/API/v6/register?version=6&hashes="+hashString, nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-Key", os.Getenv("CRYPTOWORKAPI")+" "+os.Getenv("CRYPTOWERKPW"))
	client := &http.Client{}
	if resp, err := client.Do(req); err != nil {
		log.Println("Error when sending updating blockchain: ", err)
	} else if resp.StatusCode != 200 {
		// Request sent but server responded something went wrong, try to read body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("Error when reading blockchain response body, message: ", err)
		}
		log.Printf("Blockchain responded with %d, reason: %s", resp.StatusCode, body)
	}
}
