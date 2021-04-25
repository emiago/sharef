
all: proto build

proto:
	# protoc -I api/ api/api.proto --go_out=plugins=grpc:api

	# protoc -I streamer/ --gofast_out=streamer streamer/frame.proto
	protoc -I streamer/ --gogofaster_out=streamer streamer/frame.proto
	protoc -I streamer/ --js_out=import_style=commonjs,binary:streamer streamer/frame.proto

build:
	go build -o sharef

install:
	go install

test:
	go test ./streamer/...

integrationtests:
	# cd itests/
	# go test -timeout 60s --tags integration ./itests/... -v  
	cd itests/ && go test -timeout 60s --tags integration -v -cover 

release_linux:
	env GOOS=linux GOARCH=amd64 go build -o sharef
	strip sharef
	upx sharef
	#mv sharef release/linux/

release_windows:
	env GOOS=windows GOARCH=amd64 go build -o sharef
	strip sharef
	upx sharef
	# mv sharef release/windows/sharef.exe

release: release_linux