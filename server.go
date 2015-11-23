package main

import (
	// Standard library packages
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	// Third party packages
	"github.com/julienschmidt/httprouter"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//CoordinateResponse struct
type CoordinateResponse struct {
	Results []struct {
		Geometry struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
		} `json:"geometry"`
	} `json:"results"`
}

//UpdateLocation represents the structure of location which should be updated
type UpdateLocation struct {
	Address    string `json:"address" bson:"address"`
	City       string `json:"city" bson:"city"`
	State      string `json:"state" bson:"state"`
	Zip        string `json:"zip" bson:"zip"`
	Coordinate struct {
		Lat float32 `json:"latitude"`
		Lng float32 `json:"longitude"`
	} `json:"coordinate"`
}

type (
	// PlanRouteDetails represents the structure of our resource
	PlanRouteDetails struct {
		StartingFromLocationID string   `json:"starting_from_location_id" bson:"starting_from_location_id"`
		LocationIDs            []string `json:"locationIds" bson:"locationIds"`
	}

	// Location represents the structure of our resource
	Location struct {
		Name       string        `json:"name" bson:"name"`
		Address    string        `json:"address" bson:"address"`
		City       string        `json:"city" bson:"city"`
		State      string        `json:"state" bson:"state"`
		Zip        string        `json:"zip" bson:"zip"`
		ID         bson.ObjectId `json:"id" bson:"_id"`
		Coordinate struct {
			Lat float64 `json:"latitude"`
			Lng float64 `json:"longitude"`
		} `json:"coordinate"`
	}

	//Coordinates struct
	Coordinates struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	}
	CoordinateElements struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}
)

// PriceEstimates of price estimates
type PriceEstimates struct {
	Prices []PriceEstimate `json:"prices"`
}

// PriceEstimate price estimate
type PriceEstimate struct {
	ProductId       string  `json:"product_id"`
	CurrencyCode    string  `json:"currency_code"`
	DisplayName     string  `json:"display_name"`
	Estimate        string  `json:"estimate"`
	LowEstimate     int     `json:"low_estimate"`
	HighEstimate    int     `json:"high_estimate"`
	SurgeMultiplier float64 `json:"surge_multiplier"`
	Duration        int     `json:"duration"`
	Distance        float64 `json:"distance"`
}

type TripDetails struct {
	Id                           string
	Status                       string
	Starting_from_location_id    string
	Next_destination_location_id string
	Best_route_location_ids      []string
	Total_uber_costs             float64
	Total_uber_duration          float64
	Total_distance               float64
	Uber_wait_time_eta           int
}
type (
	// Location represents the structure of our resource
	BestRouteObject struct {
		Status                 string        `json:"status" bson:"status"`
		StartingFromLocationID string        `json:"starting_from_location_id" bson:"starting_from_location_id"`
		TotalUberCosts         int           `json:"total_uber_costs" bson:"total_uber_costs"`
		TotalUberDuration      int           `json:"total_uber_duration" bson:"total_uber_duration"`
		TotalDistance          float64       `json:"total_distance" bson:"total_distance"`
		ID                     bson.ObjectId `json:"id" bson:"_id"`
		BestRouteLocationIDs   []string      `json:"best_route_location_ids" bson:"best_route_location_ids"`
	}
)

type Products struct {
	Latitude  float64
	Longitude float64
	Products  []Product `json:"products"`
}
type Product struct {
	ProductId   string `json:"product_id"`
	Description string `json:"description"`
	DisplayName string `json:"display_name"`
	Capacity    int    `json:"capacity"`
	Image       string `json:"image"`
}

type ModifyTrip struct {
	Id                        string
	Status                    string
	Starting_from_location_id string
	Best_route_location_ids   []string
	Total_uber_costs          float64
	Total_uber_duration       float64
	Total_distance            float64
	Uber_wait_time_eta        int
	Next                      int
}

type EstimatedTime struct {
	Start_latitude  string `json:"start_latitude"`
	Start_longitude string `json:"start_longitude"`
	End_latitude    string `json:"end_latitude"`
	End_longitude   string `json:"end_longitude"`
	Product_id      string `json:"product_id"`
}

var accessToken = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzY29wZXMiOlsicmVxdWVzdCJdLCJzdWIiOiI0MzM1NGQ0ZS03MzExLTRjNzgtOGUwZS1hMjRhYjZkYzQ3MTEiLCJpc3MiOiJ1YmVyLXVzMSIsImp0aSI6IjRhNzMxOGQwLTQ2NjEtNGE2Yi04YzQ5LTVkMWJmZGM3MjY0ZiIsImV4cCI6MTQ1MDc3NDU3MCwiaWF0IjoxNDQ4MTgyNTcwLCJ1YWN0IjoiVW1zQVVYaER6R0FFbXlkcWZvTlZGWVNSZ2RFQzF0IiwibmJmIjoxNDQ4MTgyNDgwLCJhdWQiOiJ1cklUVVJ3SmFNQm1zcE02QTBnSHNtWkRiQzBidnVlUCJ9.YhEIY2kJBxwF1kk8VAzyRkwCFSavzX1A_0E2R-l6G17fbluDQn1FnHQlaOSRgVsd7gAnEOF4wTi2W4bYhMMw17FML7JmdX8ws6Kb01ZzMMRbHkCwkSvZtOjGRLD5EBHBuTLBoaIY-zYfvsDO5MDeKj2Hi9d_WHMNbg2d-8nP_Bci1gShPg3U3FQZQFzlU9zRelA4hsFzkWNXkSSNf6wlwXRpEW4nsQJS3__D5alpWDS1becWtOQu_EgWCwpkQfNA1YquCgu-n1pvj7EGqSk8c0SIFft7D9DLinrVVkQ-_Rn9IEkR_iIz2KKELnuN8ZrWqDyBmYiMRC023AzJweeaxA"

func getSession() *mgo.Session {
	// Connect to our local mongo
	s, err := mgo.Dial("mongodb://assignment2:cmpe273Assignment2@ds037824.mongolab.com:37824/go_rest_api_assignment2")

	// Check if connection error, is mongo running?
	if err != nil {
		panic(err)
	}
	return s
}

type (
	// LocationController represents the controller for operating on the User resource
	LocationController struct {
		session *mgo.Session
	}
)

//NewLocationController creates a new location controller
func NewLocationController(s *mgo.Session) *LocationController {
	return &LocationController{s}
}

var initialCoordinates Coordinates
var initialLocationObject Location
var startingPoint string
var baseUrl = "https://sandbox-api.uber.com/v1/"
var totalDuration int
var totalPrice int
var totalDistance float64
var bestRouteStringArray []string
var startingCoordinates Coordinates

// GetLocation retrieves an individual user resource
func (lc LocationController) GetLocation(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Grab id
	id := p.ByName("id")

	// Verify id is ObjectId, otherwise bail
	if !bson.IsObjectIdHex(id) {
		w.WriteHeader(404)
		return
	}

	// Grab id
	oid := bson.ObjectIdHex(id)

	// Stub Location
	var locationObject Location

	// Fetch user
	if err := lc.session.DB("go_rest_api_assignment2").C("locations").FindId(oid).One(&locationObject); err != nil {
		w.WriteHeader(404)
		return
	}

	// Marshal provided interface into JSON structure
	uj, _ := json.Marshal(locationObject)

	// Write content-type, statuscode, payload
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	fmt.Fprintf(w, "%s", uj)
}

// CreateLocation creates a new user resource
func (lc LocationController) CreateLocation(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Stub a location to be populated from the body

	var locationObject Location
	//String to store address
	var queryParamBuilder string
	// Populate the user data
	json.NewDecoder(r.Body).Decode(&locationObject)

	addressKeys := strings.Fields(locationObject.Address)
	cityKeys := strings.Fields(locationObject.City)
	stateKeys := strings.Fields(locationObject.State)
	keys := append(addressKeys, cityKeys...)
	locationKeys := append(keys, stateKeys...)
	for i := 0; i < len(locationKeys); i++ {
		if i == len(locationKeys)-1 {
			queryParamBuilder += locationKeys[i]
		} else {
			queryParamBuilder += locationKeys[i] + "+"
		}
	}
	url := fmt.Sprintf("http://maps.google.com/maps/api/geocode/json?address=%s&sensor=false", queryParamBuilder)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	// read json http response
	jsonDataFromHTTP, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}
	var coordinates CoordinateResponse

	err = json.Unmarshal(jsonDataFromHTTP, &coordinates) // here!
	if err != nil {
		panic(err)
	}
	if len(coordinates.Results) == 0 {
		w.WriteHeader(400)
		return
	}
	locationObject.Coordinate.Lat = coordinates.Results[0].Geometry.Location.Lat
	locationObject.Coordinate.Lng = coordinates.Results[0].Geometry.Location.Lng

	// Add an Id
	locationObject.ID = bson.NewObjectId()

	// Write the user to mongo
	lc.session.DB("go_rest_api_assignment2").C("locations").Insert(locationObject)

	// Marshal provided interface into JSON structure
	uj, _ := json.Marshal(locationObject)

	// Write content-type, statuscode, payload
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	fmt.Fprintf(w, "%s", uj)
}

// ModifyLocation modifies a created resource
func (lc LocationController) ModifyLocation(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Grab id
	id := p.ByName("id")

	// Verify id is ObjectId, otherwise bail
	if !bson.IsObjectIdHex(id) {
		w.WriteHeader(404)
		return
	}

	// Grab id
	oid := bson.ObjectIdHex(id)
	// Stub a location to be populated from the body

	var locationObject Location
	var retrivedObject Location

	//String to store address
	var queryParamBuilder string
	// Populate the user data
	json.NewDecoder(r.Body).Decode(&locationObject)

	addressKeys := strings.Fields(locationObject.Address)
	cityKeys := strings.Fields(locationObject.City)
	stateKeys := strings.Fields(locationObject.State)
	keys := append(addressKeys, cityKeys...)
	locationKeys := append(keys, stateKeys...)
	for i := 0; i < len(locationKeys); i++ {
		if i == len(locationKeys)-1 {
			queryParamBuilder += locationKeys[i]
		} else {
			queryParamBuilder += locationKeys[i] + "+"
		}
	}
	url := fmt.Sprintf("http://maps.google.com/maps/api/geocode/json?address=%s&sensor=false", queryParamBuilder)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	// read json http response
	jsonDataFromHTTP, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}
	var coordinates CoordinateResponse

	err = json.Unmarshal(jsonDataFromHTTP, &coordinates) // here!
	if err != nil {
		panic(err)
	}
	if len(coordinates.Results) == 0 {
		w.WriteHeader(400)
		return
	}
	locationObject.Coordinate.Lat = coordinates.Results[0].Geometry.Location.Lat
	locationObject.Coordinate.Lng = coordinates.Results[0].Geometry.Location.Lng

	//Fetch user
	if err := lc.session.DB("go_rest_api_assignment2").C("locations").FindId(oid).One(&retrivedObject); err != nil {
		w.WriteHeader(404)
		return
	}
	locationObject.Name = retrivedObject.Name
	locationObject.ID = retrivedObject.ID
	if locationObject.City == "" {
		locationObject.City = retrivedObject.City
	}
	if locationObject.Address == "" {
		locationObject.Address = retrivedObject.Address
	}
	if locationObject.State == "" {
		locationObject.State = retrivedObject.State
	}
	if locationObject.Zip == "" {
		locationObject.Zip = retrivedObject.Zip
	}
	if err := lc.session.DB("go_rest_api_assignment2").C("locations").RemoveId(oid); err != nil {
		w.WriteHeader(404)
		return
	}
	// Write the user to mongo
	lc.session.DB("go_rest_api_assignment2").C("locations").Insert(locationObject)

	// Marshal provided interface into JSON structure

	uj, _ := json.Marshal(locationObject)

	// Write content-type, statuscode, payload
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	fmt.Fprintf(w, "%s", uj)

}

// DeleteLocation removes an existing user resource
func (lc LocationController) DeleteLocation(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Grab id
	id := p.ByName("id")

	// Verify id is ObjectId, otherwise bail
	if !bson.IsObjectIdHex(id) {
		w.WriteHeader(404)
		return
	}

	// Grab id
	oid := bson.ObjectIdHex(id)

	// Remove user
	if err := lc.session.DB("go_rest_api_assignment2").C("locations").RemoveId(oid); err != nil {
		w.WriteHeader(404)
		return
	}

	// Write status
	w.WriteHeader(200)
}

func (lc LocationController) GetTrip(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Grab id
	id := p.ByName("id")

	// Verify id is ObjectId, otherwise bail
	if !bson.IsObjectIdHex(id) {
		w.WriteHeader(404)
		return
	}

	// Grab id
	oid := bson.ObjectIdHex(id)

	// Stub Location
	var bestRouteProject BestRouteObject

	// Fetch user
	if err := lc.session.DB("go_rest_api_assignment2").C("priceEstimator").FindId(oid).One(&bestRouteProject); err != nil {
		w.WriteHeader(404)
		return
	}

	// Marshal provided interface into JSON structure
	uj, _ := json.Marshal(bestRouteProject)

	// Write content-type, statuscode, payload
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	fmt.Fprintf(w, "%s", uj)
}

// CreatePlan creates a new user resource
func (lc LocationController) CreatePlan(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Stub  PlanRouteDetails  to be populated from the body
	var planRouteDetails PlanRouteDetails
	//var coordinateArray []Coordinates

	coordinateArray := make([]Coordinates, 0)
	var m map[string]Coordinates
	m = make(map[string]Coordinates)

	// Populate the user data
	json.NewDecoder(r.Body).Decode(&planRouteDetails)
	if !bson.IsObjectIdHex(planRouteDetails.StartingFromLocationID) {
		w.WriteHeader(404)
		return
	}
	startingPoint = planRouteDetails.StartingFromLocationID
	startID := bson.ObjectIdHex(planRouteDetails.StartingFromLocationID)
	if err := lc.session.DB("go_rest_api_assignment2").C("locations").FindId(startID).One(&initialLocationObject); err != nil {
		w.WriteHeader(404)
		return
	}
	initialCoordinates.Lat = initialLocationObject.Coordinate.Lat
	initialCoordinates.Lng = initialLocationObject.Coordinate.Lng
	for i := 0; i < len(planRouteDetails.LocationIDs); i++ {
		id := planRouteDetails.LocationIDs[i]
		if !bson.IsObjectIdHex(id) {
			w.WriteHeader(404)
			return
		} else {
			oid := bson.ObjectIdHex(id)
			var locationObject Location
			// Fetch user
			if err := lc.session.DB("go_rest_api_assignment2").C("locations").FindId(oid).One(&locationObject); err != nil {
				w.WriteHeader(404)
				return
			} else {
				var coordinates Coordinates
				coordinates.Lat = locationObject.Coordinate.Lat
				coordinates.Lng = locationObject.Coordinate.Lng
				m[planRouteDetails.LocationIDs[i]] = coordinates
				coordinateArray = append(coordinateArray, coordinates)
			}
		}
	}
	loopingFunc(initialCoordinates, coordinateArray, m)
	var bestRouteObject BestRouteObject
	bestRouteObject.ID = bson.NewObjectId()
	bestRouteObject.Status = "planning"
	bestRouteObject.TotalDistance = totalDistance
	bestRouteObject.TotalUberCosts = totalPrice
	bestRouteObject.TotalUberDuration = totalDuration
	bestRouteObject.StartingFromLocationID = startingPoint
	bestRouteObject.BestRouteLocationIDs = bestRouteStringArray
	lc.session.DB("go_rest_api_assignment2").C("priceEstimator").Insert(bestRouteObject)
	//
	// // Marshal provided interface into JSON structure
	uj, _ := json.Marshal(bestRouteObject)
	// Write content-type, statuscode, payload
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	fmt.Fprintf(w, "%s", uj)
}

func getProductID(Coordinates CoordinateElements) {
	var url = baseUrl + "products?latitude=" + strconv.FormatFloat(float64(Coordinates.Latitude), 'f', 2, 64) + "&longitude=" +
		strconv.FormatFloat(float64(Coordinates.Longitude), 'f', 2, 64)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	client := &http.Client{}
	getResp, err := client.Do(req)
	if err != nil {
		return
	}
	var products Products
	err = json.NewDecoder(getResp.Body).Decode(&products)
	if err != nil {
		return
	}
	fmt.Println(products)
}

func (lc *LocationController) ModifyPlan(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
	id := p.ByName("id")

	c := lc.session.DB("go_rest_api_assignment2").C("priceEstimator")
	savedTripObjet := ModifyTrip{}
	err := c.Find(bson.M{"id": id}).One(&savedTripObjet)
	if err != nil {
		fmt.Fprint(rw, "Data not found")
		return
	}
	if savedTripObjet.Next == len(savedTripObjet.Best_route_location_ids) {
		rw.WriteHeader(http.StatusCreated)
		fmt.Fprint(rw, "You have completed the trip")
		return
	}
	MakeCoordinates := make([]CoordinateElements, 2)
	var NewLocationId string
	for i := 0; i < 2; i++ {
		NewLocation := new(CoordinateElements)
		if i == 0 {
			if savedTripObjet.Next == 0 {
				NewLocationId = savedTripObjet.Starting_from_location_id
			} else {
				NewLocationId = savedTripObjet.Best_route_location_ids[savedTripObjet.Next-1]
			}
		} else {
			NewLocationId = savedTripObjet.Best_route_location_ids[savedTripObjet.Next]
		}
		resp, err := http.Get("http://localhost:3000/locations/" + NewLocationId)
		if err == nil {
			body, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				var data interface{}
				json.Unmarshal(body, &data)
				var m = data.(map[string]interface{})
				var coordinates = m["UserAddress"].(map[string]interface{})["Coordinates"]
				NewLocation.Latitude = coordinates.(map[string]interface{})["Latitude"].(float64)
				NewLocation.Longitude = coordinates.(map[string]interface{})["Longitude"].(float64)
				MakeCoordinates[i] = *NewLocation
			} else {
				fmt.Println(err)
			}
		} else {
			fmt.Println(err)
		}
	}
	url := "https://sandbox-api.uber.com/v1/requests"
	estimatedTime := EstimatedTime{}
	estimatedTime.Start_latitude = strconv.FormatFloat(MakeCoordinates[0].Latitude, 'E', -1, 64)
	estimatedTime.Start_longitude = strconv.FormatFloat(MakeCoordinates[0].Longitude, 'E', -1, 64)
	estimatedTime.End_latitude = strconv.FormatFloat(MakeCoordinates[1].Latitude, 'E', -1, 64)
	estimatedTime.End_longitude = strconv.FormatFloat(MakeCoordinates[1].Longitude, 'E', -1, 64)
	estimatedTime.Product_id = "04a497f5-380d-47f2-bf1b-ad4cfdcb51f2"
	estimatedTimeReq, err := json.Marshal(estimatedTime)
	req, err = http.NewRequest("POST", url, bytes.NewBuffer(estimatedTimeReq))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var data interface{}
	json.Unmarshal(body, &data)
	var m = data.(map[string]interface{})
	eta := (m["eta"].(float64))

	c = lc.session.DB("go_rest_api_assignment2").C("priceEstimator")

	trip := ModifyTrip{}
	trip.Id = id
	var status string
	if savedTripObjet.Next == len(savedTripObjet.Best_route_location_ids)-1 {
		status = "finished"
	} else {
		status = "requesting"
	}
	counter := savedTripObjet.Next + 1
	colQuerier := bson.M{"id": id}
	change := bson.M{"$set": bson.M{"uber_wait_time_eta": 0, "next": counter, "status": status}}
	err = c.Update(colQuerier, change)
	if err != nil {
		panic(err)
	}

	tripDetails := TripDetails{}
	tripDetails.Id = id
	tripDetails.Status = status
	tripDetails.Starting_from_location_id = savedTripObjet.Starting_from_location_id
	tripDetails.Next_destination_location_id = savedTripObjet.Best_route_location_ids[counter-1]
	tripDetails.Best_route_location_ids = savedTripObjet.Best_route_location_ids
	tripDetails.Total_uber_costs = savedTripObjet.Total_uber_costs
	tripDetails.Total_uber_duration = savedTripObjet.Total_uber_duration
	tripDetails.Total_distance = savedTripObjet.Total_distance
	tripDetails.Uber_wait_time_eta = int(eta)

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusCreated)
	updateTripJson, err := json.Marshal(tripDetails)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprint(rw, string(updateTripJson))
}

func loopingFunc(initialCoordinates Coordinates, coordinateArray []Coordinates, m map[string]Coordinates) {
	var bestRoute []Coordinates
	for len(coordinateArray) > 1 {
		var index int
		temp := 0
		var lowEstimate int
		var tempCoordinates Coordinates
		var tempPrice int
		var tempDistance float64
		var tempDuration int

		for j := 0; j < len(coordinateArray)-1; j++ {
			var prices PriceEstimates
			prices = getPriceEstimate(initialCoordinates, coordinateArray[j])
			if j == 0 {
				index = j
				temp = prices.Prices[0].LowEstimate
				tempCoordinates = coordinateArray[j]
				tempPrice = prices.Prices[0].LowEstimate
				tempDuration = prices.Prices[0].Duration
				tempDistance = prices.Prices[0].Distance

			}
			lowEstimate = prices.Prices[0].LowEstimate
			if temp > lowEstimate {
				index = j
				temp = lowEstimate
				tempCoordinates = coordinateArray[j]
				tempPrice = prices.Prices[0].LowEstimate
				tempDuration = prices.Prices[0].Duration
				tempDistance = prices.Prices[0].Distance
			}
		}
		bestRoute = append(bestRoute, tempCoordinates)
		totalPrice = totalPrice + tempPrice
		totalDuration = totalDuration + tempDuration
		totalDistance = totalDistance + tempDistance
		coordinateArray = append(coordinateArray[:index], coordinateArray[index+1:]...)
		initialCoordinates.Lat = tempCoordinates.Lat
		initialCoordinates.Lng = tempCoordinates.Lng
	}

	bestRoute = append(bestRoute, coordinateArray[0])

	for i := 0; i < len(bestRoute); i++ {
		for k := range m {
			if m[k].Lat == bestRoute[i].Lat && m[k].Lng == bestRoute[i].Lng {
				bestRouteStringArray = append(bestRouteStringArray, k)
			}
		}
	}
}

func getPriceEstimate(initialCoordinates Coordinates, endCoordinates Coordinates) PriceEstimates {
	server_token := "gH8jKD3zwrrx5x2xAxihUwsVGKj7H5Yhu3ffdYL0"
	url := "https://sandbox-api.uber.com/v1/estimates/price?start_latitude=" + strconv.FormatFloat(initialCoordinates.Lat, 'f', 6, 64) + "&start_longitude=" + strconv.FormatFloat(initialCoordinates.Lng, 'f', 6, 64) +
		"&end_latitude=" + strconv.FormatFloat(endCoordinates.Lat, 'f', 6, 64) + "&end_longitude=" + strconv.FormatFloat(endCoordinates.Lng, 'f', 6, 64) + "&server_token=" + server_token
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	jsonDataFromHttp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	var prices PriceEstimates
	err = json.Unmarshal(jsonDataFromHttp, &prices) // here!
	if err != nil {
		panic(err)
	}
	return prices
}

func main() {
	// Instantiate a new router
	r := httprouter.New()

	// Get a LocationController instance
	lc := NewLocationController(getSession())
	//Get a plan
	r.GET("/trips/:id", lc.GetTrip)
	//Save a plan
	r.POST("/trips", lc.CreatePlan)

	//Update a plan
	r.PUT("/trips/:id/request", lc.ModifyPlan)
	// Get a location resource
	r.GET("/location/:id", lc.GetLocation)

	//Save a location
	r.POST("/location", lc.CreateLocation)

	//Update a Location
	r.PUT("/location/:id", lc.ModifyLocation)
	//Delete a location
	r.DELETE("/location/:id", lc.DeleteLocation)

	// Fire up the server
	http.ListenAndServe("localhost:3000", r)
}
