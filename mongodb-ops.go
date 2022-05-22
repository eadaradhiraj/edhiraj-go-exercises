package main

import (
	"context"
	"time"

	"github.com/leekchan/timeutil"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type Likes struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	liker      string             `bson:"liker,omitempty"`
	room_id    string             `bson:"room_id,omitempty"`
	message_id string             `bson:"message_id,omitempty"`
}

var ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
var client, err = mongo.Connect(ctx, options.Client().ApplyURI("mongodb+srv://eadaradhiraj:le701TTuXwAPraqM@chatappcluster.fonar.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"))
var chat_db = client.Database("ChatDB")
var likes_collection = chat_db.Collection("likes")
var users_collection = chat_db.Collection("users")
var rooms_collection = chat_db.Collection("rooms")
var room_members_collection = chat_db.Collection("room_members")

func GetLikesForRoom(room_id string) []bson.M {
	if err != nil {
		panic(err)
	}
	var likes []bson.M

	defer client.Disconnect(ctx)

	cursor, err := likes_collection.Find(context.Background(), bson.M{"room_id": room_id})
	if err != nil {
		panic(err)
	}
	if err = cursor.All(ctx, &likes); err != nil {
		panic(err)
	}
	return (likes)
}

func GetUser(username string) bson.M {
	if err != nil {
		panic(err)
	}

	defer client.Disconnect(ctx)

	var user bson.M
	if err = users_collection.FindOne(ctx, bson.M{}).Decode(&user); err != nil {
		panic(err)
	}
	return (user)
}

func save_user(username string, email string, password string) {
	password_hash := HashPassword(password)
	doc := bson.D{{"_id", username}, {"password", password_hash}, {"email", email}}
	_, err := users_collection.InsertOne(ctx, doc)
	if err != nil {
		panic(err)
	}
}

func get_current_time() string {
	n := time.Now()
	return timeutil.Strftime(&n, "%d %b, %H:%M")
}

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

func save_room(room_name string, created_by string, email string) interface{} {
	doc := bson.D{{"name", room_name}, {"created_at", get_current_time()}, {"created_by", created_by}, {"email", email}}
	result, err := rooms_collection.InsertOne(context.TODO(), doc)
	if err != nil {
		panic(err)
	}
	return result.InsertedID
}
func update_room(room_id string, room_name string) {
	objID, err := primitive.ObjectIDFromHex(room_id)
	if err != nil {
		panic(err)
	}
	rooms_collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.D{{"$set", bson.D{{"name", room_name}}}})
	room_members_collection.UpdateMany(ctx, bson.M{"_id": objID}, bson.D{{"$set", bson.D{{"room_name", room_name}}}})
	defer client.Disconnect(ctx)
}

func get_room(room_id string) bson.M {
	objID, err := primitive.ObjectIDFromHex(room_id)
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(ctx)
	var room_coll bson.M
	if err = rooms_collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&room_coll); err != nil {
		panic(err)
	}
	return (room_coll)
}

func add_room_member(room_id string, room_name string, username string, added_by string, is_room_admin bool) {
	objID, err := primitive.ObjectIDFromHex(room_id)
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(ctx)
	doc := bson.D{{"_id", bson.M{"room_id": objID, "username": username}}, {"added_at", get_current_time()}, {"room_name", room_name}, {"added_by", added_by}, {"is_room_admin", is_room_admin}}
	room_members_collection.InsertOne(ctx, doc)
}

func add_room_members(room_id string, room_name string, usernames []string, added_by string) {
	objID, err := primitive.ObjectIDFromHex(room_id)
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(ctx)
	for _, username := range usernames {
		doc := bson.D{{"_id", bson.M{"room_id": objID, "username": username}}, {"added_at", get_current_time()}, {"room_name", room_name}, {"added_by", added_by}, {"is_room_admin", is_room_admin}}
		room_members_collection.InsertOne(ctx, doc)
	}
}

func remove_room_members(room_id string, usernames string) {
	objID, err := primitive.ObjectIDFromHex(room_id)
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(ctx)
	for _, username := range usernames {
		doc := bson.D{{"_id", bson.M{"room_id": objID, "username": username}}}
		room_members_collection.InsertOne(ctx, doc)
	}
}
