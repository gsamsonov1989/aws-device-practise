package main

import (
	common "com.glebsamsonov/go-devices/common"
	"encoding/json"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"strings"
)

type ErrorMsg struct {
	Message string `json:"message"`
}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	id := request.PathParameters["id"]

	bodyString := request.Body
	device := common.Device{}
	err := json.Unmarshal([]byte(bodyString), &device)
	if err != nil {
		common.ErrorLog.Println("JSON parse error ", err)
		return events.APIGatewayProxyResponse{StatusCode: 400}, nil
	}
	//prevent from updates
	device.Id = ""
	device.CreateTime = 0

	av, err := dynamodbattribute.MarshalMap(device)
	if err != nil {
		common.ErrorLog.Println("Error marshaling into a map ", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}

	exprAttrValue := map[string]*dynamodb.AttributeValue{
		":id": {
			S: aws.String(id),
		},
	}
	updateExpr := strings.Builder{}
	updateExpr.WriteString("SET ")

	var cnt = 0
	for k, v := range av {
		cnt++
		var newKey = ":" + k
		exprAttrValue[newKey] = v
		updateExpr.WriteString(k)
		updateExpr.WriteString(" = ")
		updateExpr.WriteString(newKey)
		if cnt < len(av) {
			updateExpr.WriteString(", ")
		}
	}

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(common.TableName),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
			},
		},
		ConditionExpression:       aws.String("id = :id"),
		ExpressionAttributeValues: exprAttrValue,
		UpdateExpression:          aws.String(updateExpr.String()),
		ReturnValues:              aws.String("ALL_NEW"),
	}

	out, err := common.Dynamo.UpdateItem(input)
	if err != nil {
		var ccf *dynamodb.ConditionalCheckFailedException
		common.ErrorLog.Println("Got error calling UpdateItem: ", err)

		res := errors.As(err, &ccf)
		if res {
			bts, _ := json.Marshal(ErrorMsg{Message: "Device with such id is not found"})
			return events.APIGatewayProxyResponse{StatusCode: 400, Body: string(bts)}, nil
		}
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}

	var newDevice common.Device
	err = dynamodbattribute.UnmarshalMap(out.Attributes, &newDevice)
	if err != nil {
		common.ErrorLog.Println("Error unMarshaling into object ", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}

	deviceMarshalled, err := json.Marshal(newDevice)
	if err != nil {
		common.ErrorLog.Println("Error marshaling into JSON string ", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}

	return events.APIGatewayProxyResponse{Body: string(deviceMarshalled), StatusCode: 200}, nil
}

func main() {
	lambda.Start(Handler)
}
