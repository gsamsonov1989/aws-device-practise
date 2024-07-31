package main

import (
	common "com.glebsamsonov/go-devices/common"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	id := request.PathParameters["id"]

	if id == "" {
		common.ErrorLog.Println("device ID is not specified ")
		return events.APIGatewayProxyResponse{StatusCode: 400}, nil
	}

	input := &dynamodb.GetItemInput{
		TableName:      aws.String(common.TableName),
		ConsistentRead: aws.Bool(true),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
			},
		},
	}
	out, err := common.Dynamo.GetItem(input)
	if err != nil {
		common.InfoLog.Println("Not able to find deviceId ", id)
		return events.APIGatewayProxyResponse{StatusCode: 400}, nil
	}

	d := common.Device{}
	err = dynamodbattribute.UnmarshalMap(out.Item, &d)
	if err != nil {
		common.ErrorLog.Println("Error unmarshalling attrs into object ", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}
	if d.Id == "" {
		return events.APIGatewayProxyResponse{StatusCode: 404, Body: "Device not found"}, nil
	}

	deviceMarshalled, err := json.Marshal(d)
	if err != nil {
		common.ErrorLog.Println("Error marshaling into JSON string ", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}

	common.InfoLog.Println("Returning item: ", string(deviceMarshalled))

	return events.APIGatewayProxyResponse{Body: string(deviceMarshalled), StatusCode: 200}, nil
}

func main() {
	lambda.Start(Handler)
}
