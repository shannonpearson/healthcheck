# healthcheck

Sends a GET request to a specified URL, checks whether the correct response code is received, and sends a SMS report with the result.

## To run locally:
- specify env variables (target url, phone number, expected response code)
- `go run main.go`

## To schedule:
- deploy function to AWS Lambda
- define env variables
- subscribe to an AWS Cloudwatch event
