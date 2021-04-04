package datebase

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type Client struct {
	Collection *mongo.Collection
}

type comments struct {
	Title           string    `bson:"title"`
	TorrentHash     string    `bson:"torrent_hash"`
	TorrentAnnounce string    `bson:"torrent_announce"`
	CreateTime      time.Time `bson:"create_time"`
	Finished        bool      `bson:"finished"`
}

func NewClient(connURL string) Client {
	client, err := mongo.NewClient(options.Client().ApplyURI(connURL))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	db := client.Database("torrent_demo")
	collection := db.Collection("comments")

	return Client{
		Collection: collection,
	}
}

func (c *Client) Get(hashId string) bool {
	var dataset comments
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	if err := c.Collection.FindOne(ctx, bson.M{"torrent_hash": hashId}).Decode(&dataset); err == nil {
		return true
	}

	return false
}

func (c *Client) Insert(Title string, TorrentHash string, TorrentAnnounce string) bool {
	dataset := comments{
		Title:           Title,
		TorrentHash:     TorrentHash,
		TorrentAnnounce: TorrentAnnounce,
		CreateTime:      time.Now(),
		Finished:        false,
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	if _, err := c.Collection.InsertOne(ctx, dataset); err == nil {
		return true
	} else {
		fmt.Println(err)
	}

	return false
}

func (c *Client) MarkFinished(hashId string) bool {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	if _, err := c.Collection.UpdateOne(ctx, bson.M{"torrent_hash": hashId}, bson.M{"$set": bson.M{"finished": true}}); err == nil {
		return true
	}

	return false
}
