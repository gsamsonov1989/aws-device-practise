package main

import (
	"com.glebsamsonov/go-devices/common"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func handler(sqsEvent events.SQSEvent) error {
	for _, message := range sqsEvent.Records {

		var record common.SqsRecord
		err := json.Unmarshal([]byte(message.Body), &record)
		if err != nil {
			common.ErrorLog.Println("Incorrect JSON format ", message.Body, err)
			continue
		}

		input := &dynamodb.UpdateItemInput{
			TableName: aws.String(common.TableName),
			Key: map[string]*dynamodb.AttributeValue{
				"id": {
					S: aws.String(record.Id),
				},
			},
			ConditionExpression: aws.String("id = :id"),
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":id": {
					S: aws.String(record.Id),
				},
				":houseId": {
					S: aws.String(record.HouseId),
				},
			},
			UpdateExpression: aws.String("SET houseId = :houseId"),
			ReturnValues:     aws.String("ALL_NEW"),
		}

		_, err = common.Dynamo.UpdateItem(input)
		if err != nil {
			common.ErrorLog.Println("Unable to update device ", err)
		}
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
