package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

type FlightPath []string

// calculate finds the primary source and final destination pairs
func calculate(flightPaths []FlightPath) ([][]string, error) {
	srcToDestMap := make(map[string]string)
	sourceSet := make(map[string]bool)
	destinationSet := make(map[string]bool)

	// iterate through the flightPaths and populate maps accordingly
	for _, fp := range flightPaths {
		src := fp[0]
		dest := fp[1]

		srcToDestMap[src] = dest

		sourceSet[src] = true
		destinationSet[dest] = true
	}

	// identify the primary source and final destination using the maps
	var primarySources, finalDestinations []string
	for key := range sourceSet {
		if !destinationSet[key] {
			primarySources = append(primarySources, key)
		}
	}
	if len(primarySources) < 1 {
		return nil, fmt.Errorf("No primary sources could be identified")
	}

	for key := range destinationSet {
		if !sourceSet[key] {
			finalDestinations = append(finalDestinations, key)
		}
	}
	if len(destinationSet) < 1 {
		return nil, fmt.Errorf("No final destinations could be identified")
	} else if len(primarySources) != len(finalDestinations) {
		return nil, fmt.Errorf("Mismatch in primary sources and final destinations")
	}

	var result [][]string

	for _, src := range primarySources {
		currentSrc := src
		visitedSource := make(map[string]bool)

		for {
			if visitedSource[currentSrc] {
				return nil, errors.New("Cycle detected in flight paths")
			}
			visitedSource[currentSrc] = true
			dest, exists := srcToDestMap[currentSrc]
			if !exists {
				break
			}
			currentSrc = dest
		}
		result = append(result, []string{src, currentSrc})
	}

	return result, nil
}

func calculateHandler(w http.ResponseWriter, r *http.Request) {

	var flightPaths []FlightPath
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&flightPaths)

	// handle errors when unmarshalling json
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		// syntax error in request body
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			http.Error(w, msg, http.StatusBadRequest)

		// unmarshalling error
		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			http.Error(w, msg, http.StatusBadRequest)

		// request body is empty
		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			http.Error(w, msg, http.StatusBadRequest)

		// default server error response: 500 Internal
		default:
			log.Print(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	// validate the flightPaths
	for _, fp := range flightPaths {
		if len(fp) != 2 {
			msg := "Request body contains invalid structure"
			http.Error(w, msg, http.StatusBadRequest)
			return
		}
	}

	// calculate the source and destination
	resultFlightPath, err := calculate(flightPaths)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// marshal the response
	result, err := json.Marshal(resultFlightPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}

func main() {
	http.HandleFunc("/calculate", calculateHandler)

	fmt.Println("Starting server...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
