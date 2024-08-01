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
	"strconv"
	"time"
)

var indexName = os.Getenv("SECONDARY_INDEX")

const ZERO = "0"

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	macAddress := request.PathParameters["macAddress"]
	start := request.QueryStringParameters["start"]
	if start == "" {
		start = ZERO
	}
	end := request.QueryStringParameters["end"]
	if end == "" {
		end = strconv.FormatInt(time.Now().UnixMilli(), 10)
	}
	err := validate(start, end)
	if err != nil {
		common.ErrorLog.Println("Incorrect timestamp params passed ", start, end, err)
		return events.APIGatewayProxyResponse{StatusCode: 400}, nil
	}
	input := &dynamodb.QueryInput{
		TableName:              aws.String(common.TableName),
		IndexName:              aws.String(indexName),
		KeyConditionExpression: aws.String("macAddress = :macAddress AND createTime BETWEEN :start AND :end"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":macAddress": {
				S: aws.String(macAddress),
			},
			":start": {
				N: aws.String(start),
			},
			":end": {
				N: aws.String(end),
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

func validate(start, end string) error {
	_, err := strconv.ParseInt(start, 10, 64)
	if err != nil {
		return err
	}
	_, err = strconv.ParseInt(end, 10, 64)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	lambda.Start(Handler)
}
