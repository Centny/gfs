package gfsdb

// import (
// 	"gopkg.in/mgo.v2/bson"
// )

// func FindFolder(id string) (*Floder, error) {
// 	var folder = &Floder{}
// 	var err = C(CN_FOLDER).Find(bson.M{
// 		"_id":    id,
// 		"status": "N",
// 	}).One(folder)
// 	return folder, err
// }

// func AddFolder(folder *Floder) error {
// 	folder.Id = bson.NewObjectId().Hex()
// 	return C(CN_FOLDER).Insert(folder)
// }
