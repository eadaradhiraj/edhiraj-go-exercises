package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var valid_headers []string = []string{"firstname", "pincode", "lastname", "street", "city"}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	// Maximum upload of 10 MB files
	r.ParseMultipartForm(10 << 20)

	// Get handler for filename, size and headers
	file, handler, err := r.FormFile("myFile")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}

	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	// Create file
	dst, err := os.Create(handler.Filename)
	defer dst.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Copy the uploaded file to the created file on the filesystem
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Successfully Uploaded File\n")
}

func csvtogo() {

	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb+srv://eadaradhiraj:le701TTuXwAPraqM@chatappcluster.fonar.mongodb.net/myFirstDatabase?retryWrites=true&w=majority")

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")
	// Get a handle for your collection
	collection := client.Database("test").Collection("addresses")
	if err = collection.Drop(context.TODO()); err != nil {
		log.Fatal(err)
	}
	headers, records, err := readData("addresses.csv")

	if err != nil {
		log.Fatal(err)
	}
	// maps := make([]map[string]string, 0)
	docs := make([]interface{}, 0)
	for _, v := range records {
		var cur_doc bson.D
		for j, h := range headers {
			// cmap[h] = v[j]
			cur_doc = append(cur_doc, bson.E{h, v[j]})
		}
		docs = append(docs, cur_doc)
	}

	insertManyResult, err := collection.InsertMany(context.TODO(), docs)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted multiple documents: ", insertManyResult.InsertedIDs)
	// Close the connection once no longer needed
	err = client.Disconnect(context.TODO())

	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Connection to MongoDB closed.")
	}
}

func contains(header string) bool {
	true_counts := 0
	for _, vheader := range valid_headers {
		if header == vheader {
			true_counts += 1
		}
	}
	return true_counts == 1
}

func check_valid_headers(headers []string) bool {
	for _, header := range headers {
		if !(contains(header)) {
			return false
		}
	}
	return true
}

func readData(fileName string) ([]string, [][]string, error) {

	f, err := os.Open(fileName)

	if err != nil {
		return []string{}, [][]string{}, err
	}

	defer f.Close()

	r := csv.NewReader(f)

	r.Comma = ';'

	// skip first line
	headers, err := r.Read()
	if err != nil {
		return []string{}, [][]string{}, err
	}

	if check_valid_headers(headers) == false {
		return []string{}, [][]string{}, err
	}

	records, err := r.ReadAll()

	if err != nil {

		return []string{}, [][]string{}, err
	}

	return headers, records, nil
}

func setupRoutes() {
	http.HandleFunc("/form", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/form.html")
	})
	http.HandleFunc("/upload", uploadFile)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func main() {
	setupRoutes()
}
