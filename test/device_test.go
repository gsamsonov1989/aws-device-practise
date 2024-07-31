package test

import (
	"bytes"
	"com.glebsamsonov/go-devices/common"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"testing"
	"time"
)

var assumeRole = os.Getenv("ASSUME_ROLE")
var baseURL = os.Getenv("BASE_URL")

var region = os.Getenv("REGION")

var profile = os.Getenv("PROFILE")

var queueUrl = os.Getenv("QUEUE_URL")

var addedIds []string

var signer *v4.Signer

var dynamo *dynamodb.DynamoDB

var queue *sqs.SQS

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	shutdown()
	os.Exit(code)
}

func setup() {
	if assumeRole == "" || baseURL == "" || region == "" || profile == "" || queueUrl == "" {
		panic(any("some env variables are not set (ASSUME_ROLE, BASE_URL, REGION, " +
			"PROFILE, DYNAMODB_TABLE, QUEUE_URL must be set)"))
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile:           profile,
	}))
	creds := stscreds.NewCredentials(sess, assumeRole)
	_, err := creds.Get()
	if err != nil {
		panic(any(err))
	}
	signer = v4.NewSigner(creds)
	dynamo = dynamodb.New(sess)
	queue = sqs.New(sess)
	addedIds = []string{}
}

func shutdown() {
	for i := 0; i < len(addedIds); i++ {
		input := &dynamodb.DeleteItemInput{
			TableName: aws.String(common.TableName),
			Key: map[string]*dynamodb.AttributeValue{
				"id": {
					S: aws.String(addedIds[i]),
				},
			},
		}

		_, err := dynamo.DeleteItem(input)

		if err != nil {
			fmt.Printf("Not able to remove id %s, error = %v \n", addedIds[i], err)
		}
	}
}

func TestAddGetCycle(t *testing.T) {
	d := common.Device{DeviceName: "Sony545", DeviceType: "Clock", MacAddress: "0000000000FF"}
	assertion := func(t *testing.T, received common.Device) {
		assert.Equal(t, "Sony545", received.DeviceName)
		assert.Equal(t, "Clock", received.DeviceType)
		assert.Equal(t, "0000000000FF", received.MacAddress)
	}

	id := createDeviceAndAssert(t, d, assertion)
	getDeviceAndAssert(t, id, assertion)
}

func TestAddGetByMacCycle(t *testing.T) {
	var mac = "00000000001F"
	d1 := common.Device{DeviceName: "LG2024", DeviceType: "tv", MacAddress: mac}
	d2 := common.Device{DeviceName: "wWatch", DeviceType: "watch", MacAddress: mac}

	assertion := func(t *testing.T, received common.Device) {
		assert.Equal(t, mac, received.MacAddress)
	}

	createDeviceAndAssert(t, d1, assertion)
	createDeviceAndAssert(t, d2, assertion)

	getDevicesByMacAddressAndAssert(t, mac, func(t *testing.T, received []common.Device) {
		assert.Len(t, received, 2)
		assert.Equal(t, mac, received[0].MacAddress)
		assert.Equal(t, mac, received[1].MacAddress)
		assert.ElementsMatch(t, []string{"LG2024", "wWatch"}, []string{received[0].DeviceName, received[1].DeviceName})
		assert.ElementsMatch(t, []string{"tv", "watch"}, []string{received[0].DeviceType, received[1].DeviceType})
	})
}

func TestAddAndUpdateCycle(t *testing.T) {
	var mac = "00000000001C"
	d := common.Device{DeviceName: "SGV524", DeviceType: "Lamp", MacAddress: mac}
	id := createDeviceAndAssert(t, d, func(t *testing.T, received common.Device) {
		assert.Equal(t, mac, received.MacAddress)
		assert.Equal(t, "Lamp", received.DeviceType)
		assert.Equal(t, "SGV524", received.DeviceName)
	})
	newDevice := common.Device{DeviceName: "SGV_NEW", DeviceType: "LampSmart"}

	updateDeviceAndAssert(t, newDevice, id, func(t *testing.T, received *common.Device) {
		assert.Equal(t, mac, received.MacAddress)
		assert.Equal(t, "LampSmart", received.DeviceType)
		assert.Equal(t, "SGV_NEW", received.DeviceName)
	}, 200, &common.Device{})
}

func TestAddAndDeleteCycle(t *testing.T) {
	d := common.Device{DeviceName: "SmartChargerS4", DeviceType: "Charger", MacAddress: "00000000002C"}
	id := createDeviceAndAssert(t, d, func(t *testing.T, received common.Device) {
		assert.Equal(t, "00000000002C", received.MacAddress)
		assert.Equal(t, "Charger", received.DeviceType)
		assert.Equal(t, "SmartChargerS4", received.DeviceName)
	})
	deleteDeviceAndAssert(t, id)
}

func TestWhenUpdateWithWrongIdErrorReturned(t *testing.T) {
	d := common.Device{DeviceName: "Ebook23", DeviceType: "Book", MacAddress: "00000000003C"}
	createDeviceAndAssert(t, d, func(t *testing.T, received common.Device) {
		assert.Equal(t, "00000000003C", received.MacAddress)
		assert.Equal(t, "Book", received.DeviceType)
		assert.Equal(t, "Ebook23", received.DeviceName)
	})
	newDevice := common.Device{DeviceName: "Ebook231", DeviceType: "EBOOK"}
	updateDeviceAndAssert(t, newDevice, "11111111-1111-1111-1111-111111111111", func(t *testing.T, received *common.Device) {
		assert.Equal(t, "", received.Id)
	}, 400, &common.Device{})
}

func TestWhenSqsEventComesHouseIdGetsUpdated(t *testing.T) {
	d := common.Device{DeviceName: "Ebook23", DeviceType: "Book", MacAddress: "0000000000CC"}
	id := createDeviceAndAssert(t, d, func(t *testing.T, received common.Device) {
		assert.Equal(t, "0000000000CC", received.MacAddress)
		assert.Equal(t, "Book", received.DeviceType)
		assert.Equal(t, "Ebook23", received.DeviceName)
		assert.Equal(t, "", received.HouseId)
	})

	var rec = common.SqsRecord{Id: id, HouseId: "11111111-1111-1111-1111-111111111111"}
	bts, err := json.Marshal(rec)
	failOnError(err, t)

	var input = sqs.SendMessageInput{
		MessageGroupId:         aws.String("1"),
		MessageDeduplicationId: aws.String("1"),
		MessageBody:            aws.String(string(bts)),
		QueueUrl:               aws.String(queueUrl),
	}
	_, err = queue.SendMessage(&input)
	failOnError(err, t)

	var purgeInput = sqs.PurgeQueueInput{
		QueueUrl: aws.String(queueUrl),
	}
	defer queue.PurgeQueue(&purgeInput)

	assert.Eventually(t, func() bool {
		var res = false
		getDeviceAndAssert(t, id, func(t *testing.T, received common.Device) {
			res = "11111111-1111-1111-1111-111111111111" == received.HouseId
		})
		return res
	}, 3*time.Second, 400*time.Millisecond)
}

func createDeviceAndAssert(t *testing.T, d common.Device, assertion func(t *testing.T, received common.Device)) string {
	bts, err := json.Marshal(d)
	failOnError(err, t)

	rq, err := http.NewRequest("POST", baseURL+"/dev/device/", bytes.NewReader(bts))
	failOnError(err, t)

	rq.Header.Add("Content-Type", "application/json")
	_, err = signer.Sign(rq, bytes.NewReader(bts), "execute-api", region, time.Now())
	failOnError(err, t)

	var received common.Device
	doRequestAndAssertStatus(t, rq, &received, 200)
	assertion(t, received)

	addedIds = append(addedIds, received.Id)
	return received.Id
}

func getDeviceAndAssert(t *testing.T, id string, assertion func(t *testing.T, received common.Device)) {
	rq, err := http.NewRequest("GET", baseURL+"/dev/device/"+id, nil)
	failOnError(err, t)
	_, err = signer.Sign(rq, nil, "execute-api", region, time.Now())
	failOnError(err, t)

	var received common.Device
	doRequestAndAssertStatus(t, rq, &received, 200)
	assertion(t, received)
}

func getDevicesByMacAddressAndAssert(t *testing.T, mac string, assertion func(t *testing.T, received []common.Device)) {
	rq, err := http.NewRequest("GET", baseURL+"/dev/deviceByMac/"+mac, nil)
	failOnError(err, t)
	_, err = signer.Sign(rq, nil, "execute-api", region, time.Now())
	failOnError(err, t)

	var received []common.Device
	doRequestAndAssertStatus(t, rq, &received, 200)
	assertion(t, received)
}

func updateDeviceAndAssert(t *testing.T, newDevice common.Device, id string,
	assertion func(t *testing.T, received *common.Device), status int, rsv *common.Device) {

	bts, err := json.Marshal(newDevice)
	failOnError(err, t)

	rq, err := http.NewRequest("PUT", baseURL+"/dev/device/"+id, bytes.NewReader(bts))
	failOnError(err, t)
	rq.Header.Add("Content-Type", "application/json")
	_, err = signer.Sign(rq, bytes.NewReader(bts), "execute-api", region, time.Now())
	failOnError(err, t)

	doRequestAndAssertStatus(t, rq, rsv, status)
	assertion(t, rsv)
}

func deleteDeviceAndAssert(t *testing.T, id string) {
	rq, err := http.NewRequest("DELETE", baseURL+"/dev/device/"+id, nil)
	failOnError(err, t)
	_, err = signer.Sign(rq, nil, "execute-api", region, time.Now())
	failOnError(err, t)

	doRequestAndAssertStatus(t, rq, nil, 200)

	rq, err = http.NewRequest("GET", baseURL+"/dev/device/"+id, nil)
	failOnError(err, t)
	_, err = signer.Sign(rq, nil, "execute-api", region, time.Now())
	failOnError(err, t)

	doRequestAndAssertStatus(t, rq, nil, 404)
}

func doRequestAndAssertStatus(t *testing.T, rq *http.Request, received interface{}, status int) {
	res, err := http.DefaultClient.Do(rq)
	failOnError(err, t)
	assert.Equal(t, status, res.StatusCode)
	if received != nil {
		err = json.NewDecoder(res.Body).Decode(&received)
	}
	failOnError(err, t)
	defer res.Body.Close()
}

func failOnError(err error, t *testing.T) {
	if err != nil {
		t.Error(err)
	}
}
