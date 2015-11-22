package main

import (
	// Standard library packages
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	// Third party packages
	"github.com/julienschmidt/httprouter"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

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

var bestRoute []Coordinates

// var tempPrice int
// var tempDistance float64
//
// var tempDuration int
var totalDuration int
var totalPrice int
var totalDistance float64

var startingCoordinates Coordinates

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
	fmt.Println("Initital --------------------------")
	fmt.Println(coordinateArray)
	loopingFunc(initialCoordinates, coordinateArray)
	var bestRouteObject BestRouteObject
	bestRouteObject.ID = bson.NewObjectId()
	bestRouteObject.Status = "planning"
	bestRouteObject.TotalDistance = totalDistance
	bestRouteObject.TotalUberCosts = totalPrice
	bestRouteObject.TotalUberDuration = totalDuration
	bestRouteObject.StartingFromLocationID = startingPoint

	lc.session.DB("go_rest_api_assignment2").C("priceEstimator").Insert(bestRouteObject)
	//
	// // Marshal provided interface into JSON structure
	uj, _ := json.Marshal(bestRouteObject)
	// Write content-type, statuscode, payload
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	fmt.Fprintf(w, "%s", uj)
}

func (lc LocationController) ModifyPlan(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")

	// Verify id is ObjectId, otherwise bail
	if !bson.IsObjectIdHex(id) {
		w.WriteHeader(404)
		return
	}

	// Grab id
	oid := bson.ObjectIdHex(id)

	// Stub Location
	var bestRouteObject BestRouteObject
	// //Fetch plan
	if err := lc.session.DB("go_rest_api_assignment2").C("priceEstimator").FindId(oid).One(&bestRouteObject); err != nil {
		w.WriteHeader(404)
		return
	}
	replica := [5]string{"56509a0e5da4711114bc9e74", "565099265da4711114bc9e70", "5650994d5da4711114bc9e71", "565099965da4711114bc9e72", "565099de5da4711114bc9e73"}
	for i := 0; i < len(replica); i++ {
		id := replica[i]
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
			}
			//			coordinateArray = append(coordinateArray, coordinates)
		}
	}
	//	fmt.Println(coordinateArray)
	fmt.Println(bestRouteObject.StartingFromLocationID)
}

func loopingFunc(initialCoordinates Coordinates, coordinateArray []Coordinates) {
	var bestRoute []Coordinates
	for len(coordinateArray) > 1 {
		var index int
		temp := 0
		var lowEstimate int
		var tempCoordinates Coordinates
		var tempPrice int
		var tempDistance float64
		var tempDuration int

		fmt.Println(len(coordinateArray))
		for j := 0; j < len(coordinateArray)-1; j++ {
			var prices PriceEstimates
			prices = getPriceEstimate(initialCoordinates, coordinateArray[j])
			if j == 0 {
				index = j
				temp = prices.Prices[0].LowEstimate
				tempCoordinates = coordinateArray[j]
				fmt.Println(tempCoordinates)
				tempPrice = prices.Prices[0].LowEstimate
				tempDuration = prices.Prices[0].Duration
				tempDistance = prices.Prices[0].Distance

			}
			lowEstimate = prices.Prices[0].LowEstimate
			if temp > lowEstimate {
				fmt.Println("index ------------")
				index = j
				temp = lowEstimate
				tempCoordinates = coordinateArray[j]
				fmt.Println(tempCoordinates)
				tempPrice = prices.Prices[0].LowEstimate
				tempDuration = prices.Prices[0].Duration
				tempDistance = prices.Prices[0].Distance
			}
		}
		fmt.Println("temp Coordinates ------------")
		fmt.Println(tempCoordinates)
		bestRoute = append(bestRoute, tempCoordinates)
		fmt.Println("bestRoute ---------------")
		fmt.Println(bestRoute)
		totalPrice = totalPrice + tempPrice
		totalDuration = totalDuration + tempDuration
		totalDistance = totalDistance + tempDistance
		fmt.Println(index)
		coordinateArray = append(coordinateArray[:index], coordinateArray[index+1:]...)
		fmt.Println("coordinatearay -----------------")
		fmt.Println(coordinateArray)
		initialCoordinates.Lat = tempCoordinates.Lat
		initialCoordinates.Lng = tempCoordinates.Lng
	}

	bestRoute = append(bestRoute, coordinateArray[0])
	fmt.Println(bestRoute)
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
	// Fire up the server
	http.ListenAndServe("localhost:3000", r)
}
