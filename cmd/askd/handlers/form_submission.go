package handlers

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/ardanlabs/kit/db"
	"github.com/ardanlabs/kit/log"
	"github.com/ardanlabs/kit/web/app"
	"github.com/coralproject/shelf/internal/ask"
	"github.com/coralproject/shelf/internal/ask/form"
	"github.com/coralproject/shelf/internal/ask/form/submission"
)

// ErrInvalidCaptcha is returned when a captcha is required for a form but it
// is not valid on the request.
var ErrInvalidCaptcha = errors.New("captcha invalid")

// formSubmissionHandle maintains the set of handlers for the form submission api.
type formSubmissionHandle struct{}

// FormSubmission fronts the access to the form service functionality.
var FormSubmission formSubmissionHandle

// Create creates a new FormSubmission based on the payload of replies and the
// formID that is being submitted.
// 200 Success, 400 Bad Request, 404 Not Found, 500 Internal
func (formSubmissionHandle) Create(c *app.Context) error {
	var payload struct {
		Recaptcha string                   `json:"recaptcha"`
		Answers   []submission.AnswerInput `json:"replies"`
	}
	if err := json.NewDecoder(c.Request.Body).Decode(&payload); err != nil {
		return err
	}

	// FIXME: remove "comma ok" check after API migration. This is requried
	// because there is a wildcard collision inbetween the form_id and the id
	// for the specific endpoints old/new.
	formID, ok := c.Params["form_id"]
	if !ok {
		formID = c.Params["id"]
	}

	{
		// We should check to see if the form has a recaptcha property.
		f, err := form.Retrieve(c.SessionID, c.Ctx["DB"].(*db.DB), formID)
		if err != nil {
			return err
		}

		// If the recaptcha is enabled on the form, then we should check that the
		// response contains the data we need and if it's valid.
		if f.Settings["recaptcha"].(bool) {
			if len(payload.Recaptcha) <= 0 {
				log.Error(c.SessionID, "FormSubmission : Create", ErrInvalidCaptcha, "Payload empty")
				return ErrInvalidCaptcha
			}

			body := url.Values{
				"secret":   []string{c.App.Ctx["recaptcha"].(string)},
				"response": []string{payload.Recaptcha},
			}

			// FIXME: ensure that we can trust the proxy?

			// If there is a header on the incoming request of X-Real-IP then we know
			// that there is a proxy in place providing the real ip. Otherwise, just
			// grab the up address from the request payload.
			if ip := c.Request.Header.Get("X-Real-IP"); ip != "" {
				body["remoteip"] = []string{ip}
			} else if ip, _, err := net.SplitHostPort(c.Request.RemoteAddr); err == nil {
				body["remoteip"] = []string{ip}
			}

			log.Dev(c.SessionID, "FormSubmission : Create", "Checking Recaptcha : %s", body.Encode())

			resp, err := http.PostForm("https://www.google.com/recaptcha/api/siteverify", body)
			if err != nil {
				log.Error(c.SessionID, "FormSubmission : Create", err, "Error posting to the google service")
				return err
			}
			defer resp.Body.Close()

			var rr struct {
				Success    bool          `json:"success"`
				Hostname   string        `json:"hostname"`
				ErrorCodes []interface{} `json:"error-codes"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&rr); err != nil {
				log.Error(c.SessionID, "FormSubmission : Create", err, "Error decoding the google service response")
				return err
			}

			log.Dev(c.SessionID, "FormSubmission : Create", "Recaptcha Response : %#v", rr)

			if !rr.Success {
				log.Error(c.SessionID, "FormSubmission : Create", ErrInvalidCaptcha, "Recaptcha is invalid as per the google service")
				return ErrInvalidCaptcha
			}
		}
	}

	s, err := ask.CreateSubmission(c.SessionID, c.Ctx["DB"].(*db.DB), formID, payload.Answers)
	if err != nil {
		return err
	}

	c.Respond(s, http.StatusOK)

	return nil
}

// UpdateStatus updates the status of a FormSubmission based on the route
// params.
// 200 Success, 400 Bad Request, 404 Not Found, 500 Internal
func (formSubmissionHandle) UpdateStatus(c *app.Context) error {
	id := c.Params["id"]
	status := c.Params["status"]

	s, err := submission.UpdateStatus(c.SessionID, c.Ctx["DB"].(*db.DB), id, status)
	if err != nil {
		return err
	}

	c.Respond(s, http.StatusOK)

	return nil
}

// UpdateAnswer updates an answer based on the payload submitted to the
// endpoint.
// 200 Success, 400 Bad Request, 404 Not Found, 500 Internal
func (formSubmissionHandle) UpdateAnswer(c *app.Context) error {
	var editedAnswer struct {
		Edited string
	}

	if err := json.NewDecoder(c.Request.Body).Decode(&editedAnswer); err != nil {
		return err
	}

	id := c.Params["id"]
	answerID := c.Params["answer_id"]

	s, err := submission.UpdateAnswer(c.SessionID, c.Ctx["DB"].(*db.DB), id, submission.AnswerInput{
		WidgetID: answerID,
		Answer:   editedAnswer.Edited,
	})
	if err != nil {
		return err
	}

	c.Respond(s, http.StatusOK)

	return nil
}

// Search retrieves a set of FormSubmission's based on the search params
// provided in the query string.
// 200 Success, 400 Bad Request, 404 Not Found, 500 Internal
func (formSubmissionHandle) Search(c *app.Context) error {
	formID := c.Params["form_id"]

	limit, err := strconv.Atoi(c.Request.URL.Query().Get("limit"))
	if err != nil {
		limit = 0
	}

	skip, err := strconv.Atoi(c.Request.URL.Query().Get("skip"))
	if err != nil {
		skip = 0
	}

	opts := submission.SearchOpts{
		Query:    c.Request.URL.Query().Get("search"),
		FilterBy: c.Request.URL.Query().Get("filterby"),
	}

	if c.Request.URL.Query().Get("orderby") == "dsc" {
		opts.DscOrder = true
	}

	results, err := submission.Search(c.SessionID, c.Ctx["DB"].(*db.DB), formID, limit, skip, opts)
	if err != nil {
		return err
	}

	if csv := c.Request.URL.Query().Get("csv"); csv != "" {
		results.CSVURL = fmt.Sprintf("http://%v%v/csv", c.Request.Host, c.Request.URL.Path)
	}

	c.Respond(results, http.StatusOK)

	return nil
}

// SearchCSV retrieves a set of FormSubmission's based on the search params
// provided in the query string. It also generates a CSV with the replies.
// 200 Success, 400 Bad Request, 404 Not Found, 500 Internal
func (formSubmissionHandle) SearchCSV(c *app.Context) error {
	formID := c.Params["form_id"]

	limit := 0
	skip := 0

	opts := submission.SearchOpts{
		Query:    c.Request.URL.Query().Get("search"),
		FilterBy: c.Request.URL.Query().Get("filterby"),
	}

	if c.Request.URL.Query().Get("orderby") == "dsc" {
		opts.DscOrder = true
	}

	results, err := submission.Search(c.SessionID, c.Ctx["DB"].(*db.DB), formID, limit, skip, opts)
	if err != nil {
		return err
	}

	// Convert into [][]string to encode the CSV.
	records := getReplies(results.Submissions)

	// Marshal the data into a CSV string.
	csvData, err := encodeCSV(records)
	if err != nil {
		return err
	}

	// Set the content type.
	c.Header().Set("Content-Type", "text/csv")
	c.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"ask_%s_%s.csv\"", formID, time.Now().String()))

	c.WriteHeader(http.StatusOK)
	c.Status = http.StatusOK

	c.Write(csvData)

	return nil
}

// Retrieves a given FormSubmission based on the route params.
// 200 Success, 400 Bad Request, 404 Not Found, 500 Internal
func (formSubmissionHandle) Retrieve(c *app.Context) error {
	id := c.Params["id"]

	s, err := submission.Retrieve(c.SessionID, c.Ctx["DB"].(*db.DB), id)
	if err != nil {
		return err
	}

	c.Respond(s, http.StatusOK)

	return nil
}

// Removes a FormSubmission based on the route params.
// 200 Success, 400 Bad Request, 404 Not Found, 500 Internal
func (formSubmissionHandle) Delete(c *app.Context) error {
	id := c.Params["id"]

	// FIXME: pull out ", ok" check once old API has been deprecated.
	formID, ok := c.Params["form_id"]
	if !ok {

		// If in the event that the url does not contain the form id, we should get
		// it from the database by looking up the requested submission.

		sub, err := submission.Retrieve(c.SessionID, c.Ctx["DB"].(*db.DB), id)
		if err != nil {
			return err
		}

		formID = sub.FormID.Hex()
	}

	err := ask.DeleteSubmission(c.SessionID, c.Ctx["DB"].(*db.DB), id, formID)
	if err != nil {
		return err
	}

	c.Respond(nil, http.StatusOK)

	return nil
}

// AddFlag adds a new flag to a given FormSubmission based on the provided route
// params.
// 200 Success, 400 Bad Request, 404 Not Found, 500 Internal
func (formSubmissionHandle) AddFlag(c *app.Context) error {
	id := c.Params["id"]
	flag := c.Params["flag"]

	s, err := submission.AddFlag(c.SessionID, c.Ctx["DB"].(*db.DB), id, flag)
	if err != nil {
		return err
	}

	c.Respond(s, http.StatusOK)

	return nil
}

// RemoveFlag removes a given flag from a FormSubmission based on the provided
// route params.
// 200 Success, 400 Bad Request, 404 Not Found, 500 Internal
func (formSubmissionHandle) RemoveFlag(c *app.Context) error {
	id := c.Params["id"]
	flag := c.Params["flag"]

	s, err := submission.RemoveFlag(c.SessionID, c.Ctx["DB"].(*db.DB), id, flag)
	if err != nil {
		return err
	}

	c.Respond(s, http.StatusOK)

	return nil
}

// encodeCSV gets all the submissions and encode them into a CSV.
func encodeCSV(records [][]string) ([]byte, error) {

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	w.WriteAll(records)

	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// convert a bson.M into string.
func convertToString(m bson.M) string {
	var s string
	for _, v := range m {
		s = s + fmt.Sprintf("%v ", v)
	}
	return s
}

// get the keys of a map.
func keys(m map[string]bool) []string {

	k := make([]string, 0)
	for i := range m {
		k = append(k, i)
	}

	sort.Strings(k)

	return k
}

// get all the questions for the submissions to the form.
func getAllQuestions(submissions []submission.Submission) []string {

	q := make(map[string]bool, 0)

	for _, s := range submissions {
		for _, r := range s.Answers {
			question := r.Question.(string)
			if question != "" && !q[question] {
				q[question] = true
			}
		}
	}
	return keys(q)
}

// find the Answer for the specific question in the submission. It returns an empty string if it does not find it.
func findAnswer(s submission.Submission, q string) string {
	for _, r := range s.Answers {
		if r.Question.(string) == q {
			switch t := r.Answer.(type) {
			case bson.M:
				return convertToString(t)
			default:
				return fmt.Sprintf("%v", t)
			}
		}
	}
	return ""
}

// getReplies retrieves the questions and answers on a specific submission.
// we are returning a [][]string to use the resolut in returing the data in CSV.
func getReplies(submissions []submission.Submission) [][]string {

	questions := getAllQuestions(submissions)

	var rows [][]string

	rows = append(rows, questions)

	for _, s := range submissions {
		row := make([]string, 0)
		for _, q := range questions {
			row = append(row, findAnswer(s, q))
		}

		rows = append(rows, row)
	}

	return rows
}
