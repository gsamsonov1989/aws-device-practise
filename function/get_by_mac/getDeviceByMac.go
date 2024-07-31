package main

import (
	common "com.glebsamsonov/go-devices/common"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"os"
)

var indexName = os.Getenv("SECONDARY_INDEX")

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	macAddress := request.PathParameters["macAddress"]

	input := &dynamodb.QueryInput{
		TableName:              aws.String(common.TableName),
		IndexName:              aws.String(indexName),
		KeyConditionExpression: aws.String("macAddress = :macAddress"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":macAddress": {
				S: aws.String(macAddress),
			},
		},
	}
	out, err := common.Dynamo.Query(input)
	if err != nil {
		common.ErrorLog.Println("Got error Querying dynamodb: ", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}
	var devices []common.Device
	err = dynamodbattribute.UnmarshalListOfMaps(out.Items, &devices)
	if err != nil {
		common.ErrorLog.Println("Error unMarshaling into device list ", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}

	devicesJson, err := json.Marshal(devices)
	if err != nil {
		common.ErrorLog.Println("Error marshaling into JSON string ", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}

	return events.APIGatewayProxyResponse{Body: string(devicesJson), StatusCode: 200}, nil
}

func main() {
	lambda.Start(Handler)
}
