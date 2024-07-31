package common

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"log"
	"os"
)

var TableName = os.Getenv("DYNAMODB_TABLE")
var ErrorLog = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
var InfoLog = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

var Dynamo *dynamodb.DynamoDB

func init() {
	Dynamo = getDynamoDb()
}

type Device struct {
	Id         string `json:"id,omitempty"`
	MacAddress string `json:"macAddress,omitempty"`
	CreateTime int64  `json:"createTime,omitempty"`
	DeviceName string `json:"deviceName,omitempty"`
	DeviceType string `json:"deviceType,omitempty"`
	HouseId    string `json:"houseId,omitempty"`
}

type SqsRecord struct {
	HouseId string `json:"houseId,omitempty"`
	Id      string `json:"id,omitempty"`
}

func getDynamoDb() *dynamodb.DynamoDB {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	return dynamodb.New(sess)
}
