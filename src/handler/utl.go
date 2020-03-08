package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)


func respond(w http.ResponseWriter, v interface{}, statuscode int) {
	b, err := json.Marshal(v)

	if err != nil {
		respondError(w, fmt.Errorf("no se puede obtener la respuesta: %v", err))
		return
	}

	w.WriteHeader(statuscode)
	w.Write(b)

}

func respondError(w http.ResponseWriter, err error) {
	log.Println(err)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
