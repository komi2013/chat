package collection

type MenuStruct struct {
	ID              int           `bson:"id" json:"id"`
	Name            string        `bson:"name" json:"name"`
	Price           int           `bson:"price" json:"price"`
	NeedFacility  string        `bson:"needFacility" json:"needFacility"` // Array with [int, string]
	NeedRole       string        `bson:"needRole" json:"needRole"`
	SpecifyNameFlag int           `bson:"specifyNameFlag" json:"specifyNameFlag"`
}
