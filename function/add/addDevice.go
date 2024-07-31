package main

import (
	common "com.glebsamsonov/go-devices/common"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"
	"time"
)

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	deviceUuid := uuid.New().String()
	createTime := time.Now().UnixMilli()
	common.InfoLog.Println("Generated new item uuid:", deviceUuid)

	bodyString := request.Body
	device := common.Device{}
	err := json.Unmarshal([]byte(bodyString), &device)
	if err != nil {
		common.ErrorLog.Println("JSON parse error ", err)
		return events.APIGatewayProxyResponse{StatusCode: 400}, nil
	}

	device.Id = deviceUuid
	device.CreateTime = createTime

	av, err := dynamodbattribute.MarshalMap(device)
	if err != nil {
		common.ErrorLog.Println("Error marshalling item: ", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}

	common.InfoLog.Println("Putting item: %v", av)
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(common.TableName),
	}

	_, err = common.Dynamo.PutItem(input)
	if err != nil {
		common.ErrorLog.Println("Got error calling PutItem: ", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}

	deviceMarshalled, err := json.Marshal(device)
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
