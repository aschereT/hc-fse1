package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type empty struct{}

type message struct {
	//the standard POST by frontend to create message
	Username string `json:"username"`
	Content  string `json:"content"`
}

type stateResp struct {
	//the response by stateManager
	status  int //status. we use http statuses here to avoid having to translate between internal/external status codes
	mes     message
	id      int             //id for the message, if successful
	bulkMes map[int]message //for bulk request
}

type backendResp struct {
	//the standard response by the backend API to the frontend
	Status  int
	ID      int             `json:",omitempty"`
	Mes     message         `json:",omitempty"`
	BulkMes map[int]message `json:",omitempty"`
}
type stateReq struct {
	//the internal request to stateManager
	task     int              //0=get, 1=post
	response chan<- stateResp //the channel where stateManager will return results. Must be buffered w/ len 1. int for errors, 0=success
	id       int              //used during get
	mes      message          //used during post
}

type globalState struct {
	//for stateManager's exclusive use, preventing race conditions
	badwords map[string]empty
	messages map[int]message
	banlist  map[string]int
}

var requestChannel chan stateReq //channel to send job requests to stateManager
const censorThreshold = 3        //reject message if containing more than this number of bad words
const banThreshold = 10          //ban user if posted more than this number of offensive messages

func censor(s string, bw map[string]empty) (string, int) {
	//this doesn't catch "bliar" or "l i a r"
	//but I presume that to be out of scope
	count := 0
	fields := strings.Fields(s)
	for ind, word := range fields {
		_, ex := bw[strings.ToLower(word)]
		if ex {
			fields[ind] = strings.Repeat("*", len(word))
			count++
		}
	}
	log.Println("censor: bad words count:", count, "sanitised to:", strings.Join(fields, " "))
	return strings.Join(fields, " "), count
}

func stateManager(glob globalState, jobs chan stateReq) {
	//stateManager effectively serializes all requests to the API
	//this is bad for multi-user, but allows us to avoid
	//multiplexing. If parallel requests are needed,
	//should use an external data store anyways
	availID := 1
	for {
		job := <-jobs
		// fmt.Println(job)
		switch job.task {
		case 0:
			//get
			mes, ex := glob.messages[job.id]
			if !ex {
				job.response <- stateResp{status: http.StatusNotFound}
			} else {
				count, ex := glob.banlist[mes.Username]
				if ex && count > banThreshold {
					//post is by a banned user
					job.response <- stateResp{status: http.StatusNotFound}
					delete(glob.messages, job.id) //delete the message because it is by a banned user
					//incidentally, messages by a banned user are deleted as they are accessed.
					//however, the moment a user is banned, their messages are inaccessible
					//does this count at least for the spirit of the BONUS, if not the words?
				} else {
					job.response <- stateResp{status: http.StatusOK, mes: mes}
				}
			}
		case 1:
			//post
			count, ex := glob.banlist[job.mes.Username]
			if ex && count > banThreshold {
				//user is banned for a post
				job.response <- stateResp{status: http.StatusForbidden}
			} else {
				//make like china and censor that
				cleanMes, badCount := censor(job.mes.Content, glob.badwords)
				if badCount > censorThreshold {
					//this message is not ok
					job.response <- stateResp{status: http.StatusNotAcceptable}
					glob.banlist[job.mes.Username]++
					continue
				} else if badCount > 0 {
					//there goes their social score
					glob.banlist[job.mes.Username]++
				}
				glob.messages[availID] = message{Username: job.mes.Username, Content: cleanMes}
				resp := stateResp{status: http.StatusOK, mes: glob.messages[availID], id: availID}
				log.Printf("stateManager: post request processed %+v\n", resp)
				job.response <- resp
				availID++
			}
		case 2:
			//grab all current message
			tempCopy := make(map[int]message, len(glob.messages))
			for ind, mes := range glob.messages {
				//why does golang not have deep copy built in yet? ik this doesn't scale
				count, ex := glob.banlist[mes.Username]
				if ex && count > banThreshold {
					delete(glob.messages, ind)
				} else {
					tempCopy[ind] = mes
				}
			}
			job.response <- stateResp{status: http.StatusOK, bulkMes: tempCopy, id: availID}
		}
	}
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil || id < 0 {
		//id is not numeric, or negative
		log.Println("getHandler: new get request but id is not numeric")
		enc.Encode(backendResp{Status: http.StatusBadRequest})
	} else {
		log.Println("getHandler: new get request, id", id)
		respChan := make(chan stateResp, 1)
		requestChannel <- stateReq{task: 0, id: id, response: respChan}
		log.Println("getHandler: get request sent to stateManager")

		select {
		case resp := <-respChan:
			log.Println("getHandler: response from stateManager for id", id, "with status", resp)
			if resp.status != http.StatusOK {
				//not found and/or banned, act as if not found
				log.Println("getHandler: ", id, "is not found. or user is banned")
			} else {
				log.Println("getHandler: message id", id, "returned", resp.mes)
			}
			enc.Encode(backendResp{ID: id, Status: resp.status, Mes: resp.mes})

		case <-time.After(10 * time.Second):
			log.Println("getHandler: timed out")
			enc.Encode(backendResp{Status: http.StatusGatewayTimeout})
		}
	}
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(r.Body)
	enc := json.NewEncoder(w)
	var req message
	log.Println("postHandler: received new POST")
	if dec.Decode(&req) != nil {
		//failed to decode
		log.Println("postHandler: failed to decode POST body")
		enc.Encode(backendResp{Status: http.StatusBadRequest})
	} else {
		respChan := make(chan stateResp, 1)
		requestChannel <- stateReq{task: 1, response: respChan, mes: req}
		log.Println("postHandler: create request sent to stateManager")

		select {
		case resp := <-respChan:
			log.Println("postHandler: response from stateManager for message posting")
			switch resp.status {
			case http.StatusOK:
				log.Println("postHandler: message id is", resp.id, ", message stored is", resp.mes)
				enc.Encode(backendResp{Status: http.StatusOK, ID: resp.id, Mes: resp.mes})
			case http.StatusForbidden:
				log.Println("postHandler: user is banned, rejecting POST")
				enc.Encode(backendResp{Status: resp.status})
			case http.StatusNotAcceptable:
				log.Println("postHandler: message too profane, rejecting POST")
				enc.Encode(backendResp{Status: resp.status})
			}

		case <-time.After(10 * time.Second):
			log.Println("postHandler: timed out")
			enc.Encode(backendResp{Status: http.StatusGatewayTimeout})
		}
	}

}

func bulkHandler(w http.ResponseWriter, r *http.Request) {
	//this handler is used to send the bulk of the current messages,
	//to bring a new frontend client up to speed
	respChan := make(chan stateResp, 1)
	requestChannel <- stateReq{task: 2, response: respChan}
	log.Println("bulkHandler: bulk request sent to stateManager")
	enc := json.NewEncoder(w)
	select {
	case resp := <-respChan:
		log.Println("bulkHandler: response from stateManager for message posting")
		enc.Encode(backendResp{Status: resp.status, BulkMes: resp.bulkMes})

	case <-time.After(10 * time.Second):
		log.Println("bulkHandler: timed out")
		enc.Encode(backendResp{Status: http.StatusGatewayTimeout})
	}
}

func main() {
	//we start with an array and then translate to map,
	//as changing the array is easier than a bunch
	//of "badword": empty{}
	dictionary := []string{"bad", "horrible", "liar", "waterfall", "javascript"}
	badWordsMap := map[string]empty{}
	for _, bw := range dictionary {
		badWordsMap[bw] = empty{}
	}
	//initialize stateManager, he who controls access to the in-memory structure
	glob := globalState{badWordsMap, map[int]message{}, map[string]int{"satan": 11}}
	requestChannel = make(chan stateReq, 640) //640 ought to be enough for anybody
	go stateManager(glob, requestChannel)

	//set up routes, enforce methods
	r := mux.NewRouter().PathPrefix("/hcfse/").Subrouter()
	r.HandleFunc("/get/{id}", getHandler).Methods("GET")
	r.HandleFunc("/post", postHandler).Methods("POST")
	r.HandleFunc("/bulk", bulkHandler).Methods("GET")
	handler := cors.Default().Handler(r)

	srv := &http.Server{
		Handler: r,
		Addr:    "127.0.0.1:8000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Println("Server is listening at", srv.Addr+"/hcfse/")
	log.Fatal(http.ListenAndServe(srv.Addr, handler))
}
