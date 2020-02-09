package main

/*
This is a hackish backend service to store email sign-ups in SendGrid.

You can run it on a VPS or via a Lambda function via API Gateway Proxy

There are no tests or anything, so use at your own risk.
*/

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"strings"

	"net/http"
	"net/mail"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

const errorPage = "https://digitalmaneuver.com/error.html"
const successPage = "https://digitalmaneuver.com/success.html"

type searchResponse struct {
	Result []searchResult `json:"result"`
}

type searchResult struct {
	Id    string `json:"id"`
	Email string `json:"email"`
}

func sendGridApiRequest(method, endpoint, body string) (string, error) {
	apiKey := os.Getenv("SENDGRID_API_KEY")
	host := "https://api.sendgrid.com"

	url := host + endpoint
	payload := strings.NewReader("")
	if len(body) > 0 {
		payload = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, url, payload)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	if method == "PUT" || method == "POST" {
		req.Header.Add("content-type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bdy, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(bdy), nil
}

func addContactToSendGridContacts(email string) error {
	listId := os.Getenv("SENDGRID_LIST_ID")
	endpoint := "/v3/marketing/contacts"
	payload := `{"contacts": 
		[
			{"email":"` + email + `"}
		],
		"list_ids": ["` + listId + `"]}`
	res, err := sendGridApiRequest("PUT", endpoint, payload)
	if err != nil {
		log.Println(res)
		return errors.New("could not send address to sendgrid")
	}
	log.Println(res)
	return nil

}

func searchSendGridContacts(email string) (string, error) {
	endpoint := "/v3/marketing/contacts/search"
	payload := `{"query": "email = '` + email + `'"}`

	resp, err := sendGridApiRequest("POST", endpoint, payload)
	if err != nil {
		return "", err
	}
	data := searchResponse{}
	err = json.Unmarshal([]byte(resp), &data)
	if len(data.Result) < 1 {
		return "", errors.New("email address not in sendgrid")
	}
	return data.Result[0].Id, nil
}

func deleteContactFromSendGrid(email string) error {
	listId := os.Getenv("SENDGRID_LIST_ID")
	id, err := searchSendGridContacts(email)
	if err != nil {
		log.Println(err)
		return err
	}

	endpoint := "/v3/marketing/lists/" + listId + "/contacts?contact_ids=" + id
	res, err := sendGridApiRequest("DELETE", endpoint, "")
	log.Println(res)
	if err != nil {
		log.Println(err, res)
		return err
	}
	log.Println("deleted: " + email)
	return nil
}

func lambdaHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	resp := events.APIGatewayProxyResponse{Headers: make(map[string]string)}
	resp.Headers["Access-Control-Allow-Origin"] = "*"
	resp.StatusCode = http.StatusSeeOther

	r := http.Request{}
	r.Header = make(map[string][]string)
	for k, v := range request.Headers {
		if k == "content-type" || k == "Content-Type" {
			r.Header.Set(k, v)
		}
	}
	if request.Path == "/subscribe" {
		email, err := mail.ParseAddress(request.QueryStringParameters["email"])
		if err != nil || len(email.Address) < 5 {
			log.Println(err, email.Address)
			resp.Headers["Location"] = errorPage
			return resp, nil
		}

		err = addContactToSendGridContacts(email.Address)
		if err != nil {
			log.Println(err)
			resp.Headers["Location"] = errorPage
			return resp, nil
		}
		resp.Headers["Location"] = successPage
		return resp, nil

	}

	if request.Path == "/unsubscribe" {
		email := request.QueryStringParameters["email"]
		if email == "" {
			log.Println("email parameter not supplied")
			resp.Headers["Location"] = errorPage
			return resp, nil
		}
		deleteContactFromSendGrid(email) // NOTE: If this is done in separate goroutine on Lambda then the function will exit and return before finishing.  Ideally it would probably pass off to another Lambda function via SQS, but this will work for now.
		resp.Headers["Location"] = successPage
		return resp, nil
	}

	resp.Headers["Location"] = errorPage
	return resp, nil
}

func subscribeHandler(w http.ResponseWriter, r *http.Request) {
	email, err := mail.ParseAddress(r.URL.Query()["email"][0])
	if err != nil || len(email.Address) < 5 {
		http.Redirect(w, r, errorPage, http.StatusSeeOther)
	}

	err = addContactToSendGridContacts(email.Address)
	if err != nil || email == nil {
		log.Println(err)
		http.Redirect(w, r, errorPage, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, successPage, http.StatusSeeOther)
}

func unSubscribeHandler(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query()["email"][0]
	if email == "" {
		log.Println("email parameter not supplied")
		return
	}
	go deleteContactFromSendGrid(email)
	http.Redirect(w, r, successPage, http.StatusSeeOther)
}

func main() {
	apiKey := os.Getenv("SENDGRID_API_KEY")
	listId := os.Getenv("SENDGRID_LIST_ID")

	if apiKey == "" || listId == "" {
		panic("no SendGrid API key or listId available")
	}

	runAsLambda := os.Getenv("RUN_AS_LAMBDA")

	if runAsLambda == "TRUE" {
		lambda.Start(lambdaHandler)
	}

	http.HandleFunc("/subscribe", subscribeHandler)
	http.HandleFunc("/unsubscribe", unSubscribeHandler)
	http.ListenAndServe(":5001", nil)

}
