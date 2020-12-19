package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	contactsDatabase "github.com/tejustiwari/contact_api_project/contacts"
	"github.com/tejustiwari/contact_api_project/schema"
	usersDatabase "github.com/tejustiwari/contact_api_project/users"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Connecting with mongoDB
var users = usersDatabase.ConnectDB()
var contacts = contactsDatabase.ConnectDB()

// 1.   POST: /users to Create a User
func createUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var user schema.User

	// we decode our body request params in JSON
	_ = json.NewDecoder(r.Body).Decode(&user)

	result, err := users.InsertOne(context.TODO(), user)

	if err != nil {
		log.Fatal(err)
	}

	// we decode the recieved params in JSON
	json.NewEncoder(w).Encode(result)
}

// 2.   GET: /users/<id here> to Get a user using id
func getUser(w http.ResponseWriter, r *http.Request) {
	// set header.
	w.Header().Set("Content-Type", "application/json")

	var user schema.User

	id := strings.TrimPrefix(r.URL.Path, "/users/")
	// OR
	// re := regexp.MustCompile("/users/([!-z]+)")
	// id := re.FindStringSubmatch(r.URL.Path)[1]

	filter := bson.M{"id": id}
	// fmt.Println(id)
	err := users.FindOne(context.TODO(), filter).Decode(&user)

	if err != nil {
		log.Fatal(err, w)
	}

	json.NewEncoder(w).Encode(user)
}

// 3.   POST /contacts to Add a contact
func createContact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var contact schema.Contact

	// we decode our body request params
	_ = json.NewDecoder(r.Body).Decode(&contact)

	// insert our book model.
	result, err := contacts.InsertOne(context.TODO(), contact)

	if err != nil {
		log.Fatal(err)
	}

	json.NewEncoder(w).Encode(result)
}

// 4.   GET: //contacts?user=<user id>&infection_timestamp=<timestamp> to List all primary contacts within the last 14 days of infection
func getContacts(w http.ResponseWriter, r *http.Request) {

	query := r.URL.Query()
	userID := query.Get("user")
	t, err := time.Parse("0102030405060700", query.Get("infection_timestamp"))
	if err != nil {
		log.Fatal(err)
	}
	infectionTimestamp := t
	var fourteenDaysBeforeTimestamp = infectionTimestamp.AddDate(0, 0, -14)

	var contactsArray []string

	matchID := bson.M{"useridone": userID, "$or": bson.M{"useridtwo": userID}}
	matchTime := bson.M{"timeofcontact": bson.M{"$gt": fourteenDaysBeforeTimestamp}}
	cur, err := contacts.Aggregate(ctx, mongo.Pipeline{matchID, matchTime})

	if err != nil {
		log.Fatal(err, w)
	}

	// Close the cursor once finished
	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {

		// create a value into which the single document can be decoded
		var contact schema.Contact
		// & character returns the memory address of the following variable.
		err := cur.Decode(&contact) // decode similar to deserialize process.
		if err != nil {
			log.Fatal(err)
		}

		// add item our array
		// contactsArray = append(contactsArray, contact)
		if userID == contact.UserIDTwo {
			contactsArray = append(contactsArray, contact.UserIDOne)
		} else {
			contactsArray = append(contactsArray, contact.UserIDTwo)
		}

	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	json.NewEncoder(w).Encode(contactsArray) // encode similar to serialize process.
}

func main() {
	// Arrange the routes
	http.HandleFunc("/users", createUser)
	http.HandleFunc("/", getUser) // Generic Route, regular exp matching is done in getUser Handler
	http.HandleFunc("/contacts", createContact)
	// http.HandleFunc("/users/uid_58", getUser)
	// http.HandleFunc(fmt.Sprintf("/users/%s", userID), getUser)
	// Rewriter(http.HandleFunc("/users/{id}", getUser))
	// http.HandleFunc("/contacts", getContacts)

	// Set PORT address
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		panic(err)
	}
}
