#!/bin/sh

export GOOS="linux"
export GOARCH="amd64"

echo "start building binaries.."

mkdir -p build/lambda/add_device
go build -o build/lambda/add_device/bootstrap ./function/add/addDevice.go
echo "Finished building addDevice.go"

mkdir -p build/lambda/get_device/
go build -o build/lambda/get_device/bootstrap ./function/get/getDevice.go
echo "Finished building getDevice.go"

mkdir -p build/lambda/delete_device/
go build -o build/lambda/delete_device/bootstrap ./function/delete/deleteDevice.go
echo "Finished building deleteDevice.go"

mkdir -p build/lambda/get_device_by_mac/
go build -o build/lambda/get_device_by_mac/bootstrap ./function/get_by_mac/getDeviceByMac.go
echo "Finished building getDeviceByMac.go"

mkdir -p build/lambda/update_device/
go build -o build/lambda/update_device/bootstrap ./function/update/updateDevice.go
echo "Finished building updateDevice.go"

mkdir -p build/lambda/handle_queue_event/
go build -o build/lambda/handle_queue_event/bootstrap ./function/queue_handler/queueHandler.go
echo "Finished building queueHandler.go"

zip -j build/lambda/addDevice.zip build/lambda/add_device/bootstrap
zip -j build/lambda/getDevice.zip build/lambda/get_device/bootstrap
zip -j build/lambda/deleteDevice.zip build/lambda/delete_device/bootstrap
zip -j build/lambda/getDeviceByMac.zip build/lambda/get_device_by_mac/bootstrap
zip -j build/lambda/updateDevice.zip build/lambda/update_device/bootstrap
zip -j build/lambda/handleQueueEvent.zip build/lambda/handle_queue_event/bootstrap

echo "serverless deploy"

serverless deploy