// Package handlers contains the handler logic for processing requests.
// USE THIS AS A MODEL FOR NOW.
package handlers

import (
	"net/http"
	"net/url"

	"github.com/coralproject/shelf/pkg/query"

	"github.com/ardanlabs/kit/web/app"
)

// queryHandle maintains the set of handlers for the query api.
type queryHandle struct{}

// Query fronts the access to the query service functionality.
var Query queryHandle

//==============================================================================

// List returns all the existing query names in the system.
// 200 Success, 404 Not Found, 500 Internal
func (queryHandle) List(c *app.Context) error {
	names, err := query.GetSetNames(c.SessionID, c.DB)
	if err != nil {
		return err
	}

	c.Respond(names, http.StatusOK)
	return nil
}

// Retrieve returns the specified user from the system.
// 200 Success, 400 Bad Request, 404 Not Found, 500 Internal
func (queryHandle) Retrieve(c *app.Context) error {
	set, err := query.GetSetByName(c.SessionID, c.DB, c.Params["name"])
	if err != nil {
		return err
	}

	c.Respond(set, http.StatusOK)
	return nil
}

// Retrieve returns the specified user from the system.
// 200 Success, 400 Bad Request, 404 Not Found, 500 Internal
func (queryHandle) Execute(c *app.Context) error {
	set, err := query.GetSetByName(c.SessionID, c.DB, c.Params["name"])
	if err != nil {
		return err
	}

	var vars map[string]string
	if c.Request.URL.RawQuery != "" {
		if m, err := url.ParseQuery(c.Request.URL.RawQuery); err == nil {
			vars = make(map[string]string)
			for k, v := range m {
				vars[k] = v[0]
			}
		}
	}

	result := query.ExecuteSet(c.SessionID, c.DB, set, vars)

	c.Respond(result, http.StatusOK)
	return nil
}
