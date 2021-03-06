package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jamiealquiza/envy"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Trainer struct {
	Name string
	Age  int
	City string
}

var serverName string

func init() {
	flag.StringVar(&serverName, "serverName", "localhost:27017", "mongoDB server to connect to")
}

func insertRecords(client *mongo.Client) {

	collection := client.Database("test").Collection("trainers")

	// Some dummy data to add to the Database
	ash := Trainer{"Ash", 10, "Pallet Town"}
	misty := Trainer{"Misty", 10, "Cerulean City"}
	brock := Trainer{"Brock", 15, "Pewter City"}

	// looping a lot
	for i := 0; i < 1000; i++ {
		// Insert a single document
		insertResult, err := collection.InsertOne(context.TODO(), ash)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Inserted a single document: ", insertResult.InsertedID)

		// Insert multiple documents
		trainers := []interface{}{misty, brock}

		insertManyResult, err := collection.InsertMany(context.TODO(), trainers)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Inserted multiple documents: ", insertManyResult.InsertedIDs)
	}
}

func deleteRecords(client *mongo.Client) {

	collection := client.Database("test").Collection("trainers")

	// Delete all the documents in the collection
	deleteResult, err := collection.DeleteMany(context.TODO(), bson.D{{}})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Deleted %v documents in the trainers collection\n", deleteResult.DeletedCount)

}

func queryRecords(client *mongo.Client) {

	collection := client.Database("test").Collection("trainers")

	// Update a document
	filter := bson.D{{"name", "Ash"}}

	// Find a single document
	var result Trainer

	err := collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Found a single document: %+v\n", result)

	findOptions := options.Find()

	var results []*Trainer

	// Finding multiple documents returns a cursor
	cur, err := collection.Find(context.TODO(), bson.D{{}}, findOptions)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Iterate through the cursor
	for cur.Next(context.TODO()) {
		var elem Trainer
		err := cur.Decode(&elem)
		if err != nil {
			fmt.Println(err)
			return
		}

		results = append(results, &elem)
	}

	if err := cur.Err(); err != nil {
		fmt.Println(err)
		return
	}
	// Close the cursor once finished
	cur.Close(context.TODO())

	fmt.Printf("Found multiple documents (array of pointers): %+v\n", results)
}

func updateRecords(client *mongo.Client) {

	collection := client.Database("test").Collection("trainers")

	// Update a document
	filter := bson.D{{"name", "Ash"}}

	update := bson.D{
		{"$inc", bson.D{
			{"age", 1},
		}},
	}

	updateResult, err := collection.UpdateMany(context.TODO(), filter, update)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Matched %v documents and updated %v documents.\n", updateResult.MatchedCount, updateResult.ModifiedCount)

}

func main() {

	envy.Parse("MGDEMO") // looks for MGDEMO_SERVERNAME
	flag.Parse()

	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://" + serverName)

	// Connect to MongoDB
	fmt.Printf("Connecting to server: %v\n", serverName)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		fmt.Println(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Connected to MongoDB!")

	// enable signal trapping to ensure clean shutdown
	// pass in the connection so we can close it properly
	go func(client *mongo.Client) {
		c := make(chan os.Signal, 1)
		signal.Notify(c,
			syscall.SIGINT,  // Ctrl+C
			syscall.SIGTERM, // Termination Request
			syscall.SIGSEGV, // Segmentation Fault
			syscall.SIGABRT, // Abnormal termination
			syscall.SIGILL,  // illegal instruction
			syscall.SIGFPE)  // floating point
		sig := <-c
		fmt.Printf("Signal (%v) Detected, Shutting Down.\n", sig)
		// Close the connection once no longer needed
		err = client.Disconnect(context.TODO())
		os.Exit(2)
	}(client)

	for {
		// preping for forking off to a go routine
		insertRecords(client)
		insertRecords(client)
		updateRecords(client)
		queryRecords(client)
		deleteRecords(client)
		mySleep := rand.Intn(10)
		fmt.Printf("Taking a %v second breather... \n", mySleep)
		time.Sleep(time.Duration(mySleep) * time.Second)
		fmt.Println("Lets GO do that again!")
	}

	// Close the connection once no longer needed
	err = client.Disconnect(context.TODO())

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Connection to MongoDB closed.")
	}

}
