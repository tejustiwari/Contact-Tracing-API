package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	contactsDatabase "github.com/tejustiwari/contact_api_project/contacts"
	"github.com/tejustiwari/contact_api_project/schema"
	usersDatabase "github.com/tejustiwari/contact_api_project/users"
	"go.mongodb.org/mongo-driver/bson"
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

// 2.   GET: /users/<id here> to Get a user using id. This is a generic Route for all users i.e. id is a variable.
func getUser(w http.ResponseWriter, r *http.Request) {
	// set header.
	w.Header().Set("Content-Type", "application/json")

	var user schema.User

	id := strings.TrimPrefix(r.URL.Path, "/users/") // Extracting user id from the URL by performing expression/pattern matching
	// OR
	// re := regexp.MustCompile("/users/([!-z]+)")
	// id := re.FindStringSubmatch(r.URL.Path)[1] 

	filter := bson.M{"id": id}
	
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

	infectionTimestamp, err := time.Parse(time.RFC3339, query.Get("infection_timestamp"))

	if err != nil {
		fmt.Println(err)
	}
	var fourteenDaysBeforeTimestamp = infectionTimestamp.AddDate(0, 0, -14)

	var contactsArray []schema.Contact
	var usersArray []string

	cur, err := contacts.Find(
		context.TODO(),
		bson.D{    // Filtering out the contacts. ( This filter is made by combining multiple small filters using "AND" and "OR" operations. )
			{"$or",           // filter for getting values with -> either "useridone == userID" or "useridtwo == userID"
				bson.A{
					bson.D{{"useridone", userID}},
					bson.D{{"useridtwo", userID}},
				},
			},
			{"timeofcontact", bson.M{"$gte": fourteenDaysBeforeTimestamp}}, // filter to get time >= fourteenDaysBeforeTimestamp
			{"timeofcontact", bson.M{"$lte": infectionTimestamp}}, // filter to get time <= finfectionTimestamp
		},
	)

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
		contactsArray = append(contactsArray, contact)
		if userID == contact.UserIDTwo {
			usersArray = append(usersArray, contact.UserIDOne)
		} else {
			usersArray = append(usersArray, contact.UserIDTwo)
		}

	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	json.NewEncoder(w).Encode(usersArray) // encode similar to serialize process.
}

func main() {
	// Arrange the routes
	http.HandleFunc("/users", createUser)
	http.HandleFunc("/", getUser) // Generic Route (Eg. URL: "/userid_1"), regular exp matching is done in getUser Handler to get the userID of the desired user
	http.HandleFunc("/contacts", createContact)
	http.HandleFunc("/contacts/", getContacts)

	// Set PORT address
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		panic(err)
	}
}
