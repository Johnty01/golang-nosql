package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Person struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Firstname string             `json:"firstname,omitempty" bson:"firstname,omitempty"`
	Lastname  string             `json:"lastname,omitempty" bson:"lastname,omitempty"`
}

var client *mongo.Client

func CreatePerson(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	var person Person
	err := json.NewDecoder(request.Body).Decode(&person)
	if err != nil {
		fmt.Errorf("error happend", err)
	}
	collection := client.Database("nosqlgotest").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	result, _ := collection.InsertOne(ctx, person)
	err = json.NewEncoder(response).Encode(result)
	if err != nil {
		fmt.Errorf("error happend", err)
	}
}

func GetPeople(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	var people []Person
	collection := client.Database("nosqlgotest").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	result, err := collection.Find(ctx, bson.M{})
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message"` + err.Error() + `"}`))
		return
	}
	defer result.Close(ctx)
	for result.Next(ctx) {
		var person Person
		result.Decode(&person)
		people = append(people, person)
	}
	if err := result.Err(); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message"` + err.Error() + `"}`))
		return
	}
	err = json.NewEncoder(response).Encode(people)
	if err != nil {
		fmt.Errorf("encoding response error", err)
	}
}
func GetPerson(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	var person Person
	params := mux.Vars(request)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	collection := client.Database("nosqlgotest").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	errDB := collection.FindOne(ctx, Person{ID: id}).Decode(&person)
	if errDB != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":` + errDB.Error() + `}`))
		return
	}
	err := json.NewEncoder(response).Encode(person)
	if err != nil {
		fmt.Errorf("error happend", err)
	}
}
func main() {
	fmt.Println("starting the application now...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, _ = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	err := client.Ping(ctx, readpref.Primary())
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
	router := mux.NewRouter()
	router.HandleFunc("/person", CreatePerson).Methods("POST")
	router.HandleFunc("/people", GetPeople).Methods("GET")
	router.HandleFunc("/person/{id}", GetPerson).Methods("GET")
	http.ListenAndServe(":12345", router)
}
