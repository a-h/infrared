build:
	GOOS=linux GOARCH=arm GOARM=5 go build

deploy: build
	scp ./infrared pi@192.168.0.49:/home/pi/infrared

.PHONY:
	build deploy