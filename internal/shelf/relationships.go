package shelf

import (
	"fmt"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/ardanlabs/kit/db"
	"github.com/ardanlabs/kit/log"
)

// AddRelationship adds a relationship to the relationship manager.
func AddRelationship(context interface{}, db *db.DB, rel Relationship) (string, error) {
	log.Dev(context, "AddRelationship", "Started")

	// Get the current relationship manager.
	rm, err := GetRelManager(context, db)
	if err != nil {
		log.Error(context, "AddRelationship", err, "Completed")
		return rel.ID, err
	}

	// Make sure the given predicate does not exist already.
	var predicates []string
	for _, prevRel := range rm.Relationships {
		predicates = append(predicates, prevRel.Predicate)
	}
	if stringContains(predicates, rel.Predicate) {
		log.Error(context, "AddRelationship", err, "Completed")
		return rel.ID, fmt.Errorf("Predicate already exists")
	}

	// Assign a relationship ID, and add the relationship to the relationship manager.
	if rel.ID == "" {
		relID, err := newUUID()
		if err != nil {
			log.Error(context, "AddRelationship", err, "Completed")
			return rel.ID, err
		}
		rel.ID = relID
	}

	// Upsert the relationship.
	f := func(c *mgo.Collection) error {
		q := bson.M{"id": rel.ID}
		_, err := c.Upsert(q, &rel)
		return err
	}
	if err := db.ExecuteMGO(context, RelCollection, f); err != nil {
		log.Error(context, "AddRelationship", err, "Completed")
		return rel.ID, err
	}

	log.Dev(context, "AddRelationship", "Completed")
	return rel.ID, nil
}

// RemoveRelationship removes a relationship from the relationship manager.
func RemoveRelationship(context interface{}, db *db.DB, relID string) error {
	log.Dev(context, "RemoveRelationship", "Started")

	// Get the current relationship manager.
	rm, err := GetRelManager(context, db)
	if err != nil {
		log.Error(context, "RemoveRelationship", err, "Completed")
		return err
	}

	// Make sure the given ID is not used in an active view.
	var relIDs []string
	for _, view := range rm.Views {
		for _, segment := range view.Path {
			relIDs = append(relIDs, segment.RelationshipID)
		}
	}
	if stringContains(relIDs, relID) {
		log.Error(context, "RemoveRelationship", err, "Completed")
		return fmt.Errorf("Active view is utilizing relationship %s", relID)
	}

	// Remove the relationship.
	f := func(c *mgo.Collection) error {
		q := bson.M{"id": relID}
		err := c.Remove(q)
		return err
	}
	if err := db.ExecuteMGO(context, RelCollection, f); err != nil {
		log.Error(context, "RemoveRelationship", err, "Completed")
		return err
	}

	log.Dev(context, "RemoveRelationship", "Completed")
	return nil
}

// UpdateRelationship updates a relationship in the relationship manager.
func UpdateRelationship(context interface{}, db *db.DB, rel Relationship) error {
	log.Dev(context, "UpdateRelationship", "Started")

	// Validate the relationship.
	if err := rel.Validate(); err != nil {
		log.Error(context, "UpdateRelationship", err, "Completed")
		return err
	}

	// Remove the relationship.
	f := func(c *mgo.Collection) error {
		q := bson.M{"id": rel.ID}
		err := c.Update(q, &rel)
		return err
	}
	if err := db.ExecuteMGO(context, RelCollection, f); err != nil {
		log.Error(context, "UpdateRelationship", err, "Completed")
		return err
	}

	log.Dev(context, "UpdateRelationship", "Completed")
	return nil
}
