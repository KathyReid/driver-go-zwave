#Ninja Sphere Go zwave Driver

##Building
Run `make` in the directory of the driver

or to develop on mac and run on the sphere
`GOOS=linux GOARCH=arm go build -o driver-go-zwave main.go driver.go version.go && scp driver-go-zwave ninja@ninjasphere.local:~/`

##Running
Run `./bin/driver-go-zwave` from the `bin` directory after building
