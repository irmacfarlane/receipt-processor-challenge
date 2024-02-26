# Fetch Rewards Processor

A small webservice designed to process JSON receipts, returning JSON objects that can be used to get points awarded based on certain criteria found on the receipt. This is based on the Receipt Processor Challenge from Fetch; see https://github.com/fetch-rewards/receipt-processor-challenge for more details.

# How to Use

This webservice may be run either natively or in a Docker container. This guide assumes you have Go and Docker installed.

## Running in Docker

1. Build the Docker container:

``` docker build --tag fetch-rewards-processor . ```

2. Start the Docker image. Make sure to specify port 8080/TCP as the host port, as this is what the webservice is configured for.

## Running Natively

1. Build the executable:

``` go build -o fetch-rewards-processor ```

2. Run the executable.

## Using the Webservice

As described in the Challenge, this webservice has two endpoints: POST /receipts/process, and GET /receipts/{uuid}/points. The former sends a JSON object representing a receipt to the webservice, which upon receipt processes it, determines the proper point score based on certain criteria, and delivers a JSON object with a UUID belonging to that processed receipt. The latter returns another JSON object with a point value corresponding to that receipt's UUID.

An example curl command to send a receipt:

``` curl -v POST http://localhost:8080/receipts/process -H "Content-Type:application/json" -d @target-receipt.json ```

Another example curl command to get the points for that receipt:

``` curl -v GET http://localhost:8080/receipts/[UUID]/points -H "Content-Type:application/json" ```

(Note that `[UUID]` will be replaced with the UUID you got from the POST.)

This repo contains several example .json receipts for use.

When running the executable natively, you can also pass a .json file with receipt info as an argument, and the executable will just process the json and spit out the point value instead of running as a webservice.