package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCalculate(t *testing.T) {
	tests := []struct {
		name          string
		flightPaths   []FlightPath
		want          [][]string
		alternateWant [][]string
		wantErr       bool
	}{
		{
			name:        "Empty Flight Paths",
			flightPaths: []FlightPath{},
			wantErr:     true,
		},
		{
			name: "Simple Case",
			flightPaths: []FlightPath{
				{"SFO", "EWR"},
			},
			want:    [][]string{{"SFO", "EWR"}},
			wantErr: false,
		},
		{
			name: "Connected Flights",
			flightPaths: []FlightPath{
				{"ATL", "EWR"},
				{"SFO", "ATL"},
			},
			want:    [][]string{{"SFO", "EWR"}},
			wantErr: false,
		},
		{
			name: "Multiple Sources Destinations",
			flightPaths: []FlightPath{
				{"IND", "EWR"},
				{"SFO", "ATL"},
				{"GSO", "IND"},
				{"ATL", "GSO"},
				{"IAD", "JFK"},
			},
			want:          [][]string{{"SFO", "EWR"}, {"IAD", "JFK"}},
			alternateWant: [][]string{{"IAD", "JFK"}, {"SFO", "EWR"}},
			wantErr:       false,
		},
		{
			name: "Cycle in Flight Paths",
			flightPaths: []FlightPath{
				{"SFO", "IAD"},
				{"IAD", "JFK"},
				{"JFK", "SFO"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculate(tt.flightPaths)
			if (err != nil) != tt.wantErr {
				t.Errorf("calculate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if tt.name == "Multiple Sources Destinations" {
					if !equal(got, tt.want) && !equal(got, tt.alternateWant) {
						t.Errorf("calculate() = %v, want either %v or %v", got, tt.want, tt.alternateWant)
					}
				} else if !equal(got, tt.want) {
					t.Errorf("calculate() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestCalculateHandler(t *testing.T) {
	tests := []struct {
		name       string
		input      []FlightPath
		wantStatus int
		wantBody   [][]string
	}{
		{
			name:       "Empty Flight Paths",
			input:      []FlightPath{},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "Cycle in Flight Paths",
			input: []FlightPath{
				{"SFO", "IAD"},
				{"IAD", "JFK"},
				{"JFK", "SFO"},
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "Mismatch in Sources and Destinations",
			input: []FlightPath{
				{"SFO", "IAD"},
				{"JFK", "D"},
				{"JFK", "F"},
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "No Primary Sources",
			input: []FlightPath{
				{"SFO", "IAD"},
				{"IAD", "SFO"},
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "No Final Destinations",
			input: []FlightPath{
				{"SFO", "IAD"},
				{"JFK", "SFO"},
				{"IAD", "SFO"},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("Failed to encode input: %v", err)
			}

			req, err := http.NewRequest("POST", "/calculate", bytes.NewBuffer(body))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(calculateHandler)

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.wantStatus {
				t.Errorf("Handler returned wrong status code: got %v want %v", status, tt.wantStatus)
			}

			if tt.wantBody != nil {
				var response [][]string
				err = json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if !equal(response, tt.wantBody) {
					t.Errorf("Handler returned unexpected body: got %v want %v", response, tt.wantBody)
				}
			}
		})
	}
}

func equal(a, b [][]string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v[0] != b[i][0] || v[1] != b[i][1] {
			return false
		}
	}
	return true
}
