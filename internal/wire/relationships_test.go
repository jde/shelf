package wire_test

import (
	"testing"

	"github.com/ardanlabs/kit/db"
	"github.com/ardanlabs/kit/tests"
	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/cayley/quad"
	"github.com/coralproject/shelf/internal/wire"
	"github.com/coralproject/shelf/internal/wire/pattern/patternfix"
)

// setupGraph initializes an in-memory Cayley graph and logging for an individual test.
func setupGraph(t *testing.T) (*db.DB, *cayley.Handle, []map[string]interface{}) {
	tests.ResetLog()

	_, items, err := patternfix.Get()
	if err != nil {
		t.Fatalf("%s\tShould load item records from the fixture file : %v", tests.Failed, err)
	}
	t.Logf("%s\tShould load item records from the fixture file.", tests.Success)

	db, err := db.NewMGO(tests.Context, tests.TestSession)
	if err != nil {
		t.Fatalf("%s\tShould be able to get a Mongo session : %v", tests.Failed, err)
	}

	store, err := cayley.NewMemoryGraph()
	if err != nil {
		t.Fatalf("\t%s\tShould be able to create a new Cayley graph : %v", tests.Failed, err)
	}
	t.Logf("\t%s\tShould be able to create a new Cayley graph.", tests.Success)

	return db, store, items
}

// TestAddRemoveGraph tests if we can add/remove relationship quads to/from cayley.
func TestAddRemoveGraph(t *testing.T) {
	db, store, items := setupGraph(t)
	defer tests.DisplayLog()

	t.Log("Given the need to add/remove relationship quads from the Cayley graph.")
	{
		t.Log("\tWhen starting from an empty graph")
		{

			//----------------------------------------------------------------------
			// Infer and add the relationships to the graph.

			if err := wire.AddToGraph(tests.Context, db, store, items[0]); err != nil {
				t.Fatalf("\t%s\tShould be able to add relationships to the graph : %s", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to add relationships to the graph.", tests.Success)

			//----------------------------------------------------------------------
			// Get the relationship quads from the graph.

			p := cayley.StartPath(store, quad.String("80aa936a-f618-4234-a7be-df59a14cf8de")).Out(quad.String("authored"))
			it, _ := p.BuildIterator().Optimize()
			defer it.Close()
			for it.Next() {
				token := it.Result()
				value := store.NameOf(token)
				if quad.NativeOf(value) != "d1dfa366-d2f7-4a4a-a64f-af89d4c97d82" {
					t.Fatalf("\t%s\tShould be able to get the relationships from the graph", tests.Failed)
				}
			}
			if err := it.Err(); err != nil {
				t.Fatalf("\t%s\tShould be able to get the relationships from the graph : %s", tests.Failed, err)
			}
			it.Close()

			p = cayley.StartPath(store, quad.String("d1dfa366-d2f7-4a4a-a64f-af89d4c97d82")).Out(quad.String("on"))
			it, _ = p.BuildIterator().Optimize()
			defer it.Close()
			for it.Next() {
				token := it.Result()
				value := store.NameOf(token)
				if quad.NativeOf(value) != "c1b2bbfe-af9f-4903-8777-bd47c4d5b20a" {
					t.Fatalf("\t%s\tShould be able to get the relationships from the graph", tests.Failed)
				}
			}
			if err := it.Err(); err != nil {
				t.Fatalf("\t%s\tShould be able to get the relationships from the graph : %s", tests.Failed, err)
			}
			it.Close()
			t.Logf("\t%s\tShould be able to get relationships from the graph.", tests.Success)

			//----------------------------------------------------------------------
			// Remove the relationships from the graph.

			params1 := wire.QuadParam{
				Subject:   "80aa936a-f618-4234-a7be-df59a14cf8de",
				Predicate: "authored",
				Object:    "d1dfa366-d2f7-4a4a-a64f-af89d4c97d82",
			}
			params2 := wire.QuadParam{
				Subject:   "d1dfa366-d2f7-4a4a-a64f-af89d4c97d82",
				Predicate: "on",
				Object:    "c1b2bbfe-af9f-4903-8777-bd47c4d5b20a",
			}
			params := []wire.QuadParam{params1, params2}

			if err := wire.RemoveFromGraph(tests.Context, store, params); err != nil {
				t.Fatalf("\t%s\tShould be able to remove relationships from the graph : %s", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to remove relationships from the graph.", tests.Success)

			//----------------------------------------------------------------------
			// Try to get the relationships.

			var count int
			p = cayley.StartPath(store, quad.String("80aa936a-f618-4234-a7be-df59a14cf8de")).Out(quad.String("authored"))
			it, _ = p.BuildIterator().Optimize()
			defer it.Close()
			for it.Next() {
				count++
			}
			if err := it.Err(); err != nil {
				t.Fatalf("\t%s\tShould be able to verify the empty graph : %s", tests.Failed, err)
			}
			it.Close()

			p = cayley.StartPath(store, quad.String("d1dfa366-d2f7-4a4a-a64f-af89d4c97d82")).Out(quad.String("on"))
			it, _ = p.BuildIterator().Optimize()
			defer it.Close()
			for it.Next() {
				count++
			}
			if err := it.Err(); err != nil {
				t.Fatalf("\t%s\tShould be able to verify the empty graph : %s", tests.Failed, err)
			}
			it.Close()

			if count != 0 {
				t.Fatalf("\t%s\tShould be able to verify the empty graph", tests.Failed)
			}
			t.Logf("\t%s\tShould be able to verify the empty graph.", tests.Success)
		}
	}
}

// TestGraphParamFail tests if we can handle invalid quad parameters.
func TestGraphParamFail(t *testing.T) {
	_, store, _ := setupGraph(t)
	defer tests.DisplayLog()

	t.Log("Given the need to add/remove relationship quads from the Cayley graph.")
	{
		t.Log("\tWhen starting from an empty graph")
		{
			//----------------------------------------------------------------------
			// Create some example parameters to import into the graph.

			params1 := wire.QuadParam{
				Subject:   "",
				Predicate: "",
				Object:    "the ring",
			}
			params2 := wire.QuadParam{
				Subject:   "orcs",
				Predicate: "chase",
				Object:    "frodo",
			}
			params := []wire.QuadParam{params1, params2}

			//----------------------------------------------------------------------
			// Try to remove the invalid relationship to the graph.

			if err := wire.RemoveFromGraph(tests.Context, store, params); err == nil {
				t.Fatalf("\t%s\tShould be able to catch invalid quad parameters on remove : %s", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to catch invalid quad parameters on remove.", tests.Success)
		}
	}
}
