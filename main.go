package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "regexp"

    "github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-lambda-go/lambda"
)

var isbnRegexp = regexp.MustCompile(`[0-9]{3}\-[0-9]{10}`)
var errorLogger = log.New(os.Stderr, "ERROR ", log.Llongfile)

type book struct {
    ISBN   string `json:"isbn"`
    Title  string `json:"title"`
    Author string `json:"author"`
}

func router(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
  switch req.HTTPMethod {
    case "GET":
      if(req.QueryStringParameters == nil) {
        return index(req)
      } else {
        return show(req)
      }
    case "POST":
        return create(req)
    case "DELETE":
        return delete(req)
    default:
        return clientError(http.StatusMethodNotAllowed)
    }
}

func index(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
  bks, err := getItems()
  if err != nil {
      return serverError(err)
  }
  js, err := json.Marshal(bks)
  if err != nil {
      return serverError(err)
  }

  return events.APIGatewayProxyResponse{
      StatusCode: http.StatusOK,
      Body:       string(js),
  }, nil
}

func show(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
    isbn := req.QueryStringParameters["isbn"]
    if !isbnRegexp.MatchString(isbn) {
        return clientError(http.StatusBadRequest)
    }

    bk, err := getItem(isbn)
    if err != nil {
        return serverError(err)
    }
    if bk == nil {
        return clientError(http.StatusNotFound)
    }

    js, err := json.Marshal(bk)
    if err != nil {
        return serverError(err)
    }

    return events.APIGatewayProxyResponse{
        StatusCode: http.StatusOK,
        Body:       string(js),
    }, nil
}

func create(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
    /* ↓があると、headerにaplication/jsonを指定しても通らない。
     if req.Headers["Content-Type"] != "application/json" {
       return clientError(http.StatusNotAcceptable)
     }
    */

    bk := new(book)
    err := json.Unmarshal([]byte(req.Body), bk)
    if err != nil {
        return clientError(http.StatusUnprocessableEntity)
    }

    if !isbnRegexp.MatchString(bk.ISBN) {
        return clientError(http.StatusBadRequest)
    }
    if bk.Title == "" || bk.Author == "" {
        return clientError(http.StatusBadRequest)
    }

    err = putItem(bk)
    if err != nil {
        return serverError(err)
    }

    return events.APIGatewayProxyResponse{
        StatusCode: 201,
        Headers:    map[string]string{"Location": fmt.Sprintf("/books?isbn=%s", bk.ISBN)},
    }, nil
}

func delete(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
  isbn := req.QueryStringParameters["isbn"]
  if !isbnRegexp.MatchString(isbn) {
      return clientError(http.StatusBadRequest)
  }

  err := deleteItem(isbn)
  if err != nil {
      return serverError(err)
  }

  return events.APIGatewayProxyResponse{
      StatusCode: http.StatusOK,
  }, nil
}

func serverError(err error) (events.APIGatewayProxyResponse, error) {
    errorLogger.Println(err.Error())

    return events.APIGatewayProxyResponse{
        StatusCode: http.StatusInternalServerError,
        Body:       http.StatusText(http.StatusInternalServerError),
    }, nil
}

func clientError(status int) (events.APIGatewayProxyResponse, error) {
    return events.APIGatewayProxyResponse{
        StatusCode: status,
        Body:       http.StatusText(status),
    }, nil
}

func main() {
    lambda.Start(router)
}
