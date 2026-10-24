package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var connectionString = "mongodb://localhost:27017"
var dbName = "inventory"
var colName = "supplies"

var collection *mongo.Collection

type Item struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name     string             `json:"name,omitempty" bson:"name,omitempty"`
	Quantity int                `json:"quantity,omitempty" bson:"quantity,omitempty"`
	Price    int                `json:"price,omitempty" bson:"price,omitempty"`
}

func init() {
	clientOption := options.Client().
		ApplyURI(connectionString)
	client, err := mongo.Connect(
		context.TODO(),
		clientOption,
	)
	if err != nil {
		log.Fatal(err)
	}
	collection = client.
		Database(dbName).
		Collection(colName)

	fmt.Println("Collection instance is ready")
}
func pushInventoryItem(collection *mongo.Collection, item Item) {
	res, err := collection.InsertOne(context.TODO(), item)
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Print("Item inserted successfully: ", res.InsertedID)
}
func popInventoryItem(collection *mongo.Collection, itemName string) {
	res, err := collection.DeleteOne(context.TODO(), bson.M{"name": itemName})
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Print(res.DeletedCount, "Item deleted successfully")
}
func modInventoryItem(collection *mongo.Collection, itemName string, newValue int, field string) {
	res, err := collection.UpdateOne(context.TODO(), bson.M{"name": itemName}, bson.M{"$set": bson.M{field: newValue}})
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Print(res.ModifiedCount, "Item updated successfully")
}
func fetchInventory(collection *mongo.Collection) []Item {
	arr := []Item{}
	res, _ := collection.Find(context.TODO(), bson.D{})
	res.All(context.TODO(), &arr)
	return arr
}
func getAllItems(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(fetchInventory(collection))
}
func popItem(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	popInventoryItem(collection, name)
}
func pushItem(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	qty, _ := strconv.Atoi(r.PathValue("qty"))
	amt, _ := strconv.Atoi(r.PathValue("amt"))
	newItem := Item{
		Name:     name,
		Quantity: qty,
		Price:    amt,
	}
	pushInventoryItem(collection, newItem)
}
func modCost(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	amt, _ := strconv.Atoi(r.PathValue("amt"))

	modInventoryItem(collection, name, amt, "price")
}
func modQty(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	qty, _ := strconv.Atoi(r.PathValue("qty"))
	modInventoryItem(collection, name, qty, "quantity")
}
func main() {
	fmt.Print("Server is up and running")
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", getAllItems)
	mux.HandleFunc("POST /new/{name}/{qty}/{amt}", pushItem)
	mux.HandleFunc("POST /erase/{name}", popItem)
	mux.HandleFunc("POST /update_cost/{name}/{amt}", modCost)
	mux.HandleFunc("POST /update_qty/{name}/{qty}", modQty)
	log.Fatal(http.ListenAndServe(":3000", mux))

}
