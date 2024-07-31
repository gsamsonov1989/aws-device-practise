package main

import (
	common "com.glebsamsonov/go-devices/common"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	id := request.PathParameters["id"]
	common.InfoLog.Println("Deleting item: %s", id)

	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(common.TableName),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
			},
		},
	}

	_, err := common.Dynamo.DeleteItem(input)
	if err != nil {
		common.ErrorLog.Println("Got error calling DeleteItem: ", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}

	return events.APIGatewayProxyResponse{StatusCode: 200}, nil
}

func main() {
	lambda.Start(Handler)
}
