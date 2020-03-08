package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/hako/branca"

	"github.com/Mynor2397/social-network/src/handler"
	"github.com/Mynor2397/social-network/src/mysql"
	"github.com/Mynor2397/social-network/src/service"
)

var port = ":4545"

func main() {
	//Configuracion del archivo log
	logfile, err := os.OpenFile("test.log", os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer logfile.Close()
	log.SetOutput(logfile)

	//Configuración del token
	codec := branca.NewBranca("supersecretkeyyoushouldnotcommit")
	codec.SetTTL(uint32(service.TokenLifespan.Seconds()))

	//Configuración de las instancias del servicio
	db := mysql.Connect()
	s := service.New(db, codec)
	h := handler.New(s)

	fmt.Printf("Starting server on port %s", port)
	//Configuracion de los encabezados para peticiones cruzadas
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	if err := http.ListenAndServe(port, handlers.CORS(headersOk, methodsOk, originsOk)(h)); err != nil {
		log.Fatalf("No se pudo iniciar el servidor: %v", err)
	}
}
