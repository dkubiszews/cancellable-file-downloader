# How to test

## In one terminal start the test server

```
cd test-server
go run main.go resources/set1/
```

## In second terminal use client
```
cd client
go get

go run main.go 'http://localhost:8080' 'A' out; # NOTICE: make sure that out dir is empty.
go run main.go 'http://localhost:8080' 'C' out; # NOTICE: make sure that out dir is empty.
go run main.go 'http://localhost:8080' 'AB' out; # NOTICE: make sure that out dir is empty.
```
