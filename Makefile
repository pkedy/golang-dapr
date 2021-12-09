PHONY: run-custom-http run-custom-grpc run-sdk-http run-sdk-grpc send-widget send-gadget send-thingamajig

run-custom-http:
	dapr run --app-id inventory --config ./config.yaml --components-path ./components --app-protocol http --app-port 3001 --dapr-http-port 3500 -- go run cmd/inventory/main.go

run-custom-grpc:
	dapr run --app-id inventory --config ./config.yaml --components-path ./components --app-protocol grpc --app-port 4001 --dapr-http-port 3500 -- go run cmd/inventory/main.go

run-sdk-http:
	dapr run --app-id inventory --config ./config.yaml --components-path ./components --app-protocol http --app-port 3002 --dapr-http-port 3500 -- go run cmd/inventory/main.go

run-sdk-grpc:
	dapr run --app-id inventory --config ./config.yaml --components-path ./components --app-protocol grpc --app-port 4002 --dapr-http-port 3500 -- go run cmd/inventory/main.go

run-products:
	dapr run --app-id products --config ./config.yaml --components-path ./components --app-protocol grpc --app-port 50151 -- go run cmd/products/main.go

send-widget:
	curl -s http://localhost:3500/v1.0/publish/pubsub/inventory -H Content-Type:application/cloudevents+json --data @messages/widget.json

send-gadget:
	curl -s http://localhost:3500/v1.0/publish/pubsub/inventory -H Content-Type:application/cloudevents+json --data @messages/gadget.json

send-thingamajig:
	curl -s http://localhost:3500/v1.0/publish/pubsub/inventory -H Content-Type:application/cloudevents+json --data @messages/thingamajig.json

get-widget:
	curl -s http://localhost:3000/v1/widgets/widget | jq

get-gadget:
	curl -s http://localhost:3000/v1/gadgets/gadget | jq

get-product:
	curl -s http://localhost:3000/v1/products/thingamajig | jq
