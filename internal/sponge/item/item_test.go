package item_test

import (
	"testing"

	"github.com/ardanlabs/kit/cfg"
	"github.com/ardanlabs/kit/db"
	"github.com/ardanlabs/kit/db/mongo"
	"github.com/ardanlabs/kit/tests"
	"github.com/coralproject/shelf/internal/sponge/item"
	"github.com/coralproject/shelf/internal/sponge/item/itemfix"
)

func init() {
	// Initialize the configuration and logging systems. Plus anything
	// else the web app layer needs.
	tests.Init("XENIA")

	// Initialize MongoDB using the `tests.TestSession` as the name of the
	// master session.
	cfg := mongo.Config{
		Host:     cfg.MustString("MONGO_HOST"),
		AuthDB:   cfg.MustString("MONGO_AUTHDB"),
		DB:       cfg.MustString("MONGO_DB"),
		User:     cfg.MustString("MONGO_USER"),
		Password: cfg.MustString("MONGO_PASS"),
	}
	tests.InitMongo(cfg)
}

// prefix is what we are looking to delete after the test.
const prefix = "ITEST_"

// TestUpsertDelete tests if we can add/remove an item to/from the db.
func TestUpsertDelete(t *testing.T) {
	tests.ResetLog()
	defer tests.DisplayLog()

	db, err := db.NewMGO(tests.Context, tests.TestSession)
	if err != nil {
		t.Fatalf("\t%s\tShould be able to get a Mongo session : %v", tests.Failed, err)
	}
	defer db.CloseMGO(tests.Context)

	defer func() {
		if err := itemfix.Remove(tests.Context, db, prefix); err != nil {
			t.Fatalf("\t%s\tShould be able to remove the items : %v", tests.Failed, err)
		}
		t.Logf("\t%s\tShould be able to remove the items.", tests.Success)
	}()

	t.Log("Given the need to upsert and delete items.")
	{
		t.Log("\tWhen starting from an empty items collection")
		{
			//----------------------------------------------------------------------
			// Get the fixture.

			items, err := itemfix.Get()
			if err != nil {
				t.Fatalf("\t%s\tShould be able retrieve item fixture : %s", tests.Failed, err)
			}

			//----------------------------------------------------------------------
			// Upsert the item.

			if err := item.Upsert(tests.Context, db, &items[0]); err != nil {
				t.Fatalf("\t%s\tShould be able to upsert a item : %s", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to upsert a item.", tests.Success)

			//----------------------------------------------------------------------
			// Get the item.

			itemsBack, err := item.GetByIDs(tests.Context, db, []string{items[0].ID})
			if err != nil {
				t.Fatalf("\t%s\tShould be able to get the item by ID : %s", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to get the item by ID.", tests.Success)

			//----------------------------------------------------------------------
			// Check that we got the item we expected.

			if items[0].ID != itemsBack[0].ID {
				t.Logf("\t%+v", items[0])
				t.Logf("\t%+v", itemsBack[0])
				t.Fatalf("\t%s\tShould be able to get back the same item.", tests.Failed)
			}
			t.Logf("\t%s\tShould be able to get back the same item.", tests.Success)

			//----------------------------------------------------------------------
			// Delete the item.

			if err := item.Delete(tests.Context, db, items[0].ID); err != nil {
				t.Fatalf("\t%s\tShould be able to delete the item : %s", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to delete the item.", tests.Success)

			//----------------------------------------------------------------------
			// Get the item.

			itemsBack, err = item.GetByIDs(tests.Context, db, []string{items[0].ID})
			if len(itemsBack) != 0 {
				t.Fatalf("\t%s\tShould generate an error when getting an item with the deleted ID : %s", tests.Failed, err)
			}
			t.Logf("\t%s\tShould generate an error when getting an item with the deleted ID.", tests.Success)
		}
	}
}

// TestGetByIDs tests if we can get items from the db.
func TestGetAll(t *testing.T) {
	tests.ResetLog()
	defer tests.DisplayLog()

	db, err := db.NewMGO(tests.Context, tests.TestSession)
	if err != nil {
		t.Fatalf("\t%s\tShould be able to get a Mongo session : %v", tests.Failed, err)
	}
	defer db.CloseMGO(tests.Context)

	defer func() {
		if err := itemfix.Remove(tests.Context, db, prefix); err != nil {
			t.Fatalf("\t%s\tShould be able to remove the items : %v", tests.Failed, err)
		}
		t.Logf("\t%s\tShould be able to remove the items.", tests.Success)
	}()

	t.Log("Given the need to get items in the database by IDs.")
	{
		t.Log("\tWhen starting from an empty items collection")
		{
			items1, err := itemfix.Get()
			if err != nil {
				t.Fatalf("\t%s\tShould be able retrieve item fixture : %s", tests.Failed, err)
			}

			var itemIDs []string
			for _, it := range items1 {
				if err := item.Upsert(tests.Context, db, &it); err != nil {
					t.Fatalf("\t%s\tShould be able to upsert items : %s", tests.Failed, err)
				}
				itemIDs = append(itemIDs, it.ID)
			}
			t.Logf("\t%s\tShould be able to upsert items.", tests.Success)

			items2, err := item.GetByIDs(tests.Context, db, itemIDs)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to get items by IDs : %s", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to get items by IDs.", tests.Success)

			if len(items1) != len(items2) {
				t.Logf("\t%+v", items1)
				t.Logf("\t%+v", items2)
				t.Fatalf("\t%s\tShould be able to get back the same items.", tests.Failed)
			}
			t.Logf("\t%s\tShould be able to get back the same items.", tests.Success)
		}
	}
}
