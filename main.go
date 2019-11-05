package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

// HandleRequest sends an HTTP request to the specified website and sends a text if all is not well
func HandleRequest(ctx context.Context) error {
	targetURL := os.Getenv("TARGET_URL")
	fmt.Printf("Starting health check for %v", targetURL)

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := sns.New(sess)

	httpResp, httpErr := http.Get(targetURL)

	targetPhoneNumber := os.Getenv("PHONE_NUM")

	// send error if page is down
	if httpErr != nil {
		fmt.Println(httpErr)

		// send text
		params := &sns.PublishInput{
			Message:     aws.String(httpErr.Error()),
			PhoneNumber: aws.String(targetPhoneNumber),
		}
		resp, err := svc.Publish(params)
		if err != nil {
			fmt.Println("Publish error.")
			if awsErr, ok := err.(awserr.Error); ok {
				// process SDK error
				fmt.Printf("AWS Publish error: %v", awsErr.Code())
			} else {
				fmt.Printf("Error processing AWS publish error: %v", err.Error())
			}
			return nil
		}

		// Pretty-print the response data.
		fmt.Println("Publish response: ", resp)
		return nil
	}
	defer httpResp.Body.Close()

	fmt.Println("HTTP response received: ", httpResp.StatusCode, http.StatusText(httpResp.StatusCode))

	expectedResponseCode, convErr := strconv.Atoi(os.Getenv("EXPECTED_RESPONSE_CODE"))
	if convErr != nil {
		expectedResponseCode = 200
	}

	if httpResp.StatusCode != expectedResponseCode {
		errorString := "Incorrect status code received from " + targetURL + ": " + strconv.Itoa(httpResp.StatusCode)
		fmt.Println("Error string: ", errorString)
		fmt.Println("Phone number", targetPhoneNumber)

		// send text
		params := &sns.PublishInput{
			Message:     aws.String(errorString),
			PhoneNumber: aws.String(targetPhoneNumber),
		}
		resp, err := svc.Publish(params)

		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				fmt.Printf("AWS Publish error: %v", awsErr.Code())
			} else {
				fmt.Printf("Error processing AWS publish error: %v", err.Error())
			}
			return nil
		}

		fmt.Println("Publish response: ", resp)
		return nil
	}

	params := &sns.PublishInput{
		Message:     aws.String("Health check successfully completed for " + targetURL),
		PhoneNumber: aws.String(targetPhoneNumber),
	}
	resp, err := svc.Publish(params)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			fmt.Printf("AWS Publish error: %v", awsErr.Code())
		} else {
			fmt.Printf("Error processing AWS publish error: %v", err.Error())
		}
		return nil
	}

	fmt.Println("Publish response: ", resp)

	body, readError := ioutil.ReadAll(httpResp.Body)
	if readError != nil {
		fmt.Printf("Response body read error: %v", readError)
		return nil
	}
	fmt.Println("HTTP response body: ", body)
	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
