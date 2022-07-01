package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/wafer-bw/whatsmyip/spec"
	"google.golang.org/protobuf/proto"
)

// GetRouter returns the router for the API
func GetRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", Handler).Methods(http.MethodGet)
	return r
}

func resolver(request *http.Request) *spec.IPReply {
	// Try to get Cf-Connecting-Ip header first
	ip := request.Header.Get("Cf-Connecting-Ip")
	if ip == "" {
		// If not found, try to get X-Forwarded-For header
		ip = request.Header.Get("X-Forwarded-For")
		if ip == "" {
			// If not found, try to get RemoteAddr
			ip = request.RemoteAddr
			if ip == "" {
				// If not found, return an empty IPReply
				return &spec.IPReply{}
			}
		}
	}
	return &spec.IPReply{Ip: ip}
}

func respond(w http.ResponseWriter, r *http.Request, body []byte, err error) {
	switch err {
	case nil:
		w.Write(body)
	default:
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	return
}

func getAcceptHeader(r *http.Request) string {
	return strings.ToLower(r.Header.Get("Accept"))
}

func marshal(w http.ResponseWriter, r *http.Request, reply *spec.IPReply) (body []byte, err error) {
	accept := getAcceptHeader(r)
	w.Header().Set("Content-Type", accept)
	switch accept {
	case "application/protobuf":
		return proto.Marshal(reply)
	case "application/json":
		return json.Marshal(reply)
	default:
		w.Header().Set("Content-Type", "text/plain")
		return []byte(reply.Ip), nil
	}
}

// Handler responds with the IP address of the request
func Handler(w http.ResponseWriter, r *http.Request) {
	log.Println(*r)
	body, err := marshal(w, r, resolver(r))
	respond(w, r, body, err)
}
