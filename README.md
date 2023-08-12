# Flight Calculator

This service calculates the primary source and final destination from a series of flight paths. It is implemented in Go and listens on port 8080.

## Overview

Given an input of flight paths as pairs of source and destination airports, the service identifies the primary source and the final destination.

## Requirements

- Go 1.16 or higher

## Installation

1. Clone the repository:

```
git clone https://github.com/jsridharma/flight_calculator.git
```

2. Navigate to the project directory:

```
cd flight_calculator
```


3. Build the project:

```
go build
```


## Running the Application

To start the server, run the following command:

```
./flight_calculator
```

The server will start listening on port 8080.

## API Endpoint

The service exposes the following endpoint:

### `/calculate`

- **Method:** GET
- **Content-Type:** `application/json`
- **Body:** An array of flight paths (pairs of source and destination airports)
- **Response:** A JSON array containing the primary source and final destination

#### Example Request

```
curl -X GET -H "Content-Type: application/json" -d '[["IND", "EWR"], ["SFO", "ATL"], ["GSO", "IND"], ["ATL", "GSO"]]' http://localhost:8080/calculate
```


## Testing

Run the unit tests by executing the following command:

```
go test
```