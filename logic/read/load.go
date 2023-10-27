package read

import (
  "context"
  "fmt"
  "log"
  "time"


  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/mongo/options"

  "chat/common"
  "chat/collection"
)

func Load() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
	if err != nil {
		log.Print(err)
	}
	defer c.Disconnect(ctx)
	db1 := c.Database(common.MongoDb1)

  coll := db1.Collection("channel_group")
  filter := bson.D{{"alias_name", bson.D{{"$in", [2]string{"ali1", "Python"}}}}}
  // filter := bson.D{}
  project := bson.D{
    {"channel_id", 1},
    {"unread_flg", 1}}
  opts := options.Find().SetProjection(project)
  cursor, err := coll.Find(context.TODO(), filter, opts)
  if err != nil {
    return err
  }
  var results []collection.ChannelGroupStruct
  if err = cursor.All(context.TODO(), &results); err != nil {
    return err
  }
  fmt.Printf(" results %s\n", results)
  for _, r := range results {
    cursor.Decode(&r)
    fmt.Printf(" r %s\n", r.ChannelID)
  }
  return nil
}
