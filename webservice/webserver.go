package webservice

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/OmniCar/autobot/vehicle"
)

const (
	dateFmt = "2006-01-02"
	timeFmt = "2006-01-02T15:04:05"
)

const (
	errJSONEncoding = iota * 100
	errLogRetrieval
	errLookup
	errMarshalling
)

type status struct {
	Status string `json:"status"`
	Uptime string `json:"uptime"`
}

type storeStatus struct {
	HistorySize  int       `json:"historySize"`
	LastStatusAt time.Time `json:"lastStatusAt"`
	LastStatus   string    `json:"lastStatusMessage"`
}

// WebServer represents the REST-API part of autobot.
type WebServer struct {
	startTime time.Time
	store     *vehicle.Store
}

// APIError is the error returned to clients whenever an internal error has happened.
type APIError struct {
	HTTPCode int    `json:"-"`
	Code     int    `json:"code,omitempty"`
	Message  string `json:"message"`
}

// APIVehicle is the API representation of Vehicle. It has a JSON representation.
// Some fields that are only for internal use, are left out, and others are converted into something more readable.
type APIVehicle struct {
	Hash         string `json:"hash"`
	Country      string `json:"country"`
	RegNr        string `json:"regNr"`
	VIN          string `json:"vin"`
	Brand        string `json:"brand"`
	Model        string `json:"model"`
	FuelType     string `json:"fuelType"`
	FirstRegDate string `json:"firstRegDate"`
}

// vehicleToAPIType converts a vehicle.Vehicle into the local APIVehicle, which is used for the http request/response.
func vehicleToAPIType(veh vehicle.Vehicle) APIVehicle {
	return APIVehicle{strconv.FormatUint(veh.MetaData.Hash, 10), vehicle.RegCountryToString(veh.MetaData.Country), veh.RegNr, veh.VIN, veh.Brand, veh.Model, veh.FuelType, veh.FirstRegDate.Format(dateFmt)}
}

// New initialises a new webserver. You need to start it by calling Serve().
func New(store *vehicle.Store) *WebServer {
	return &WebServer{time.Now(), store}
}

// JSONError serves the given error as JSON.
func (srv *WebServer) JSONError(w http.ResponseWriter, handlerErr APIError) {
	data := struct {
		Err APIError `json:"error"`
	}{handlerErr}
	d, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(handlerErr.HTTPCode)
	fmt.Fprint(w, string(d))
}

// returnStatus returns a small JSON struct with the various information such as service uptime and status.
func (srv *WebServer) returnStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	uptime := time.Since(srv.startTime).Truncate(time.Second)
	s := status{"running", uptime.String()}
	bytes, err := json.Marshal(s)
	if err != nil {
		srv.JSONError(w, APIError{http.StatusInternalServerError, errJSONEncoding, err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}

// returnVehicleStoreStatus fetches and returns the current status of the vehicle store.
func (srv *WebServer) returnVehicleStoreStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	c, err := srv.store.CountLog()
	if err != nil {
		srv.JSONError(w, APIError{http.StatusInternalServerError, errLogRetrieval, err.Error()})
		return
	}
	entry, err := srv.store.LastLog()
	if err != nil {
		srv.JSONError(w, APIError{http.StatusInternalServerError, errLogRetrieval, err.Error()})
		return
	}
	s := storeStatus{c, entry.LoggedAt, entry.Message}
	bytes, err := json.Marshal(s)
	if err != nil {
		srv.JSONError(w, APIError{http.StatusInternalServerError, errJSONEncoding, err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}

// lookupVehicle allows vehicle lookups based on hash value, VIN or registration number. A country must always be
// provided.
func (srv *WebServer) lookupVehicle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	country := r.URL.Query().Get("country")
	hash := r.URL.Query().Get("hash")
	regNr := r.URL.Query().Get("regnr")
	vin := r.URL.Query().Get("vin")
	regCountry := vehicle.RegCountryFromString(country) // For now, we're forcing unknown countries into "DK".
	if regNr == "" && vin == "" && hash == "" {
		srv.JSONError(w, APIError{http.StatusBadRequest, errLookup, "Missing query parameter 'hash', 'regnr' or 'vin'"})
		return
	}
	// "country" is required for "regnr" and "vin" only.
	if country == "" && hash == "" {
		srv.JSONError(w, APIError{http.StatusBadRequest, errLookup, "Missing query parameter 'country'"})
		return
	}
	var (
		veh vehicle.Vehicle
		err error
	)
	if hash != "" {
		veh, err = srv.store.LookupByHash(hash)
	} else if regNr != "" {
		veh, err = srv.store.LookupByRegNr(regCountry, regNr, false)
	} else {
		veh, err = srv.store.LookupByVIN(regCountry, vin, false)
	}
	if err != nil {
		srv.JSONError(w, APIError{http.StatusInternalServerError, errLookup, err.Error()})
		return
	}
	if veh == (vehicle.Vehicle{}) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	bytes, err := json.Marshal(vehicleToAPIType(veh))
	if err != nil {
		srv.JSONError(w, APIError{http.StatusInternalServerError, errMarshalling, err.Error()})
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}

// Serve starts the web server.
// Currently, there is no config for which port to listen on.
func (srv *WebServer) Serve(port uint) error {
	http.HandleFunc("/", srv.returnStatus)                                // GET.
	http.HandleFunc("/vehiclestore/status", srv.returnVehicleStoreStatus) // GET.
	http.HandleFunc("/lookup", srv.lookupVehicle)                         // GET.
	srv.startTime = time.Now()
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
