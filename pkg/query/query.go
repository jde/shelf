package query

import (
	"errors"
	"strings"

	"github.com/ardanlabs/kit/db"
	"github.com/ardanlabs/kit/db/mongo"
	"github.com/ardanlabs/kit/log"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Contains the name of Mongo collections.
const (
	Collection         = "query_sets"
	CollectionHistory  = "query_sets_history"
	CollectionExecTest = "test_query"
)

// =============================================================================

// UpsertSet is used to create or update an existing Set document.
func UpsertSet(context interface{}, db *db.DB, qs *Set) error {
	log.Dev(context, "UpsertSet", "Started : Name[%s]", qs.Name)

	// We need to know if this is a new set.
	var new bool
	if _, err := GetSetByName(context, db, qs.Name); err != nil {
		if err != mgo.ErrNotFound {
			log.Error(context, "UpsertSet", err, "Completed")
			return err
		}

		new = true
	}

	// Insert or update the query set.
	f := func(c *mgo.Collection) error {
		q := bson.M{"name": qs.Name}
		log.Dev(context, "UpsertSet", "MGO : db.%s.upsert(%s, %s)", c.Name, mongo.Query(q), mongo.Query(qs))
		_, err := c.Upsert(q, qs)
		return err
	}

	if err := db.ExecuteMGO(context, Collection, f); err != nil {
		log.Error(context, "UpsertSet", err, "Completed")
		return err
	}

	// Add a history record if this query set is new.
	if new {
		f = func(c *mgo.Collection) error {
			qh := bson.M{
				"name": qs.Name,
				"sets": []bson.M{},
			}

			log.Dev(context, "UpsertSet", "MGO : db.%s.insert(%s)", c.Name, mongo.Query(qh))
			return c.Insert(qh)
		}

		if err := db.ExecuteMGO(context, CollectionHistory, f); err != nil {
			log.Error(context, "UpsertSet", err, "Completed")
			return err
		}
	}

	// Add this query set to the beginning of the history.
	f = func(c *mgo.Collection) error {
		q := bson.M{"name": qs.Name}
		qu := bson.M{
			"$push": bson.M{
				"sets": bson.M{
					"$each":     []*Set{qs},
					"$position": 0,
				},
			},
		}

		log.Dev(context, "UpsertSet", "MGO : db.%s.update(%s, %s)", c.Name, mongo.Query(q), mongo.Query(qu))
		_, err := c.Upsert(q, qu)
		return err
	}

	if err := db.ExecuteMGO(context, CollectionHistory, f); err != nil {
		log.Error(context, "UpsertSet", err, "Completed")
		return err
	}

	log.Dev(context, "UpsertSet", "Completed")
	return nil
}

// =============================================================================

// GetSetNames retrieves a list of rule names.
func GetSetNames(context interface{}, db *db.DB) ([]string, error) {
	log.Dev(context, "GetSetNames", "Started")

	var names []bson.M
	f := func(c *mgo.Collection) error {
		q := bson.M{"name": 1}
		log.Dev(context, "GetSetNames", "MGO : db.%s.find({}, %s).sort([\"name\"])", c.Name, mongo.Query(q))
		return c.Find(nil).Select(q).Sort("name").All(&names)
	}

	if err := db.ExecuteMGO(context, Collection, f); err != nil {
		log.Error(context, "GetSetNames", err, "Completed")
		return nil, err
	}

	var qsn []string
	for _, doc := range names {
		name := doc["name"].(string)
		if strings.HasPrefix(name, "test") {
			continue
		}

		qsn = append(qsn, name)
	}

	log.Dev(context, "GetSetNames", "Completed : QSN[%+v]", qsn)
	return qsn, nil
}

// GetSetByName retrieves the configuration for the specified Set.
func GetSetByName(context interface{}, db *db.DB, name string) (*Set, error) {
	log.Dev(context, "GetSetByName", "Started : Name[%s]", name)

	var qs Set
	f := func(c *mgo.Collection) error {
		q := bson.M{"name": name}
		log.Dev(context, "GetSetByName", "MGO : db.%s.findOne(%s)", c.Name, mongo.Query(q))
		return c.Find(q).One(&qs)
	}

	if err := db.ExecuteMGO(context, Collection, f); err != nil {
		log.Error(context, "GetSetByName", err, "Completed")
		return nil, err
	}

	log.Dev(context, "GetSetByName", "Completed : QS[%+v]", &qs)
	return &qs, nil
}

// GetLastSetHistoryByName gets the last written Set within the query_history
// collection and returns the last one else returns a non-nil error if it fails.
func GetLastSetHistoryByName(context interface{}, db *db.DB, name string) (*Set, error) {
	log.Dev(context, "GetLastSetHistoryByName", "Started : Name[%s]", name)

	var result struct {
		Name string `bson:"name"`
		Sets []Set  `bson:"sets"`
	}

	f := func(c *mgo.Collection) error {
		q := bson.M{"name": name}
		proj := bson.M{"sets": bson.M{"$slice": 1}}

		log.Dev(context, "GetLastSetHistoryByName", "MGO : db.%s.find(%s,%s)", c.Name, mongo.Query(q), mongo.Query(proj))
		return c.Find(q).Select(proj).One(&result)
	}

	err := db.ExecuteMGO(context, CollectionHistory, f)
	if err != nil {
		log.Error(context, "GetLastSetHistoryByName", err, "Complete")
		return nil, err
	}

	if result.Sets == nil {
		err := errors.New("History not found")
		log.Error(context, "GetLastSetHistoryByName", err, "Complete")
		return nil, err
	}

	log.Dev(context, "GetLastSetHistoryByName", "Completed : QS[%+v]", &result.Sets[0])
	return &result.Sets[0], nil
}

// =============================================================================

// DeleteSet is used to remove an existing Set document.
func DeleteSet(context interface{}, db *db.DB, name string) error {
	log.Dev(context, "DeleteSet", "Started : Name[%s]", name)

	qs, err := GetSetByName(context, db, name)
	if err != nil {
		return err
	}

	f := func(c *mgo.Collection) error {
		q := bson.M{"name": qs.Name}
		log.Dev(context, "DeleteSet", "MGO : db.%s.remove(%s)", c.Name, mongo.Query(q))
		return c.Remove(q)
	}

	if err := db.ExecuteMGO(context, Collection, f); err != nil {
		log.Error(context, "DeleteSet", err, "Completed")
		return err
	}

	log.Dev(context, "DeleteSet", "Completed")
	return nil
}
