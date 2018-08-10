package main

import (
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/dynamodb"
    "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

var db = dynamodb.New(session.New(), aws.NewConfig().WithRegion("ap-northeast-1"))

func getItem(isbn string) (*book, error) {
    input := &dynamodb.GetItemInput{
        TableName: aws.String("Books"),
        Key: map[string]*dynamodb.AttributeValue{
            "ISBN": {
                S: aws.String(isbn),
            },
        },
    }

    result, err := db.GetItem(input)
    if err != nil {
        return nil, err
    }
    if result.Item == nil {
        return nil, nil
    }

    bk := new(book)
    err = dynamodbattribute.UnmarshalMap(result.Item, bk)
    if err != nil {
        return nil, err
    }

    return bk, nil
}

// Add a book record to DynamoDB.
func putItem(bk *book) error {
    input := &dynamodb.PutItemInput{
        TableName: aws.String("Books"),
        Item: map[string]*dynamodb.AttributeValue{
            "ISBN": {
                S: aws.String(bk.ISBN),
            },
            "Title": {
                S: aws.String(bk.Title),
            },
            "Author": {
                S: aws.String(bk.Author),
            },
        },
    }

    _, err := db.PutItem(input)
    return err
}
