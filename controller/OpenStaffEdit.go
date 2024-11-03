package controller

import (
  "context"
  "encoding/json"
  "fmt"
  "log"
  "net/http"
  "time"

  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo/options"
  // "go.mongodb.org/mongo-driver/bson/primitive"

  // webpush "github.com/SherClockHolmes/webpush-go"

  // "chat/collection"
  "chat/common"
)

// Define the structs for the parent structure
// type TimeEntry struct {
// 	BookTitle   string            `json:"bookTitle"`
// 	Date        string            `json:"date"`
// 	LimitStart  string            `json:"limitStart"`
// 	LimitEnd    string            `json:"limitEnd"`
// 	NeedRoles   [][2]interface{}  `json:"needRoles"` // Assuming [int, string] in a mixed slice
// }

// type Parent struct {
// 	AdminGroup    string          `json:"adminGroup"`
// 	NeedFacilities [][2]interface{} `json:"needFacilities"`
// 	MaxFacility   string          `json:"maxFacility"`
// 	Times         []TimeEntry     `json:"times"`
// 	BookPatternID string          `json:"bookPatternID"`
// 	OpenTimes     [][3]string     `json:"openTimes,omitempty"`
// }

func OpenStaffEdit(w http.ResponseWriter, r *http.Request) {
	session, err := common.Session(w,r)
	if err != nil {
  	http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
    return
	}
  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()
  c, err := mongo.Connect(ctx, options.Client().ApplyURI(common.Mongo1))
  if err != nil {
    log.Print(err)
  }
  defer c.Disconnect(ctx)
  db1 := c.Database(common.MongoDb1)

  verifiedName := ""
  for _, aliasArray := range session.AliasArray {
  	if (r.FormValue("aliasName") == aliasArray[0]) {
  		verifiedName = r.FormValue("aliasName")
  	}
  }
  if verifiedName == "" {
  	fmt.Printf("verifiedName err %s\n", r.FormValue("aliasName"))
  	return
  }

// Parse okStaffJSON to extract entries
	var okStaff []interface{}
	err = json.Unmarshal([]byte(r.FormValue("okStaff")), &okStaff)
	if err != nil {
		fmt.Printf("okStaff err %s\n", r.FormValue("okStaff"))
		return
	}

	// Get staff entries from parsed okStaff data
	staffEntries, ok := okStaff[2].([]interface{})
	if !ok {
		return
	}

	// Access the collection
	coll := db1.Collection("open_staff")

	// Loop over each staff entry in okStaff
	for _, entry := range staffEntries {
		staffEntry, ok := entry.([]interface{})
		if !ok || len(staffEntry) < 5 {
			log.Println("Invalid staff entry format, skipping:", entry)
			continue
		}

		// Parse individual fields from the entry
		dateStr, _ := staffEntry[0].(string)
		startStr, _ := staffEntry[1].(string)
		role, _ := staffEntry[3].(string)
		flag := int(staffEntry[4].(float64))
		endStr, _ := staffEntry[5].(string)

		// Convert date and time to time.Time format for OpenStart
		openStart, err := time.Parse("2006-01-02T15:04", fmt.Sprintf("%sT%s", dateStr, startStr))
		if err != nil {
			log.Println("Failed to parse date and time for OpenStart, skipping:", err)
			continue
		}

		openEnd, err := time.Parse("2006-01-02T15:04", fmt.Sprintf("%sT%s", dateStr, endStr))
		if err != nil {
			log.Println("Failed to parse date and time for OpenStart, skipping:", err)
			continue
		}
		id := r.FormValue("windowID") + openStart.Format("200601021504") + verifiedName
		// Define filter and update document
		filter := bson.M{
			"_id":  id,
		}

		updateFields := bson.M{
			"window_id":  r.FormValue("windowID"),
			"alias_name": verifiedName,
			"open_start": openStart,
			"open_end":   openEnd,
			"role":       role,
		}
		update := bson.D{
			{"$set", updateFields},
		}
		opts := options.Update().SetUpsert(true)
		if flag == 1 {
			_, err = coll.UpdateOne(context.TODO(), filter, update, opts)
			if err != nil {
				log.Println("Failed to upsert document:", err)
				continue
			}
		} else if flag == -1 {
			_, err = coll.DeleteOne(context.TODO(), filter)
			if err != nil {
				log.Println("Failed to delete document:", err)
				continue
			}
		}
	}

  fmt.Fprint(w, `{"Status":"1"}`)
}
