package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi"
)

type res struct {
	Order   string `json:"order"`
	Status  string `json:"status"`
	Accrual int    `json:"accrual,omitempty"`
}

func main() {
	var listen string
	flag.StringVar(&listen, "a", ":9090", "HTTP listen addr")
	flag.Parse()
	r := chi.NewRouter()
	r.Get("/api/orders/{id}", genAcc)
	server := &http.Server{
		Addr:    listen,
		Handler: r,
	}
	go func() {
		log.Println("Listen on: ", listen)
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
		log.Println("server gracefully shut down")
	}()
	waitForShutDown(server)
}

func waitForShutDown(server *http.Server) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := server.Shutdown(ctx)
	if err != nil {
		log.Fatal("failed shut down server")
	}
}

func genAcc(w http.ResponseWriter, r *http.Request) {
	var resStatus int
	ID := chi.URLParam(r, "id")
	res := res{}

	i := rand.Intn(10)

	time.Sleep(time.Duration(i / 2 * int(time.Second)))
	switch i {
	case 7:
		resStatus = http.StatusNoContent
		log.Printf("order: %s, error generated", ID)
	case 8:
		res.Order = ID
		res.Status = "INVALID"
		res.Accrual = 0
		resStatus = http.StatusOK
		log.Printf("order: %s, status: %s", ID, res.Status)
	case 9:
		res.Order = ID
		res.Status = "REGISTERED"
		res.Accrual = 0
		resStatus = http.StatusOK
		log.Printf("order: %s, status: %s", ID, res.Status)
	case 10:
		res.Order = ID
		res.Status = "PROCESSING"
		res.Accrual = 0
		resStatus = http.StatusOK
		log.Printf("order: %s, status: %s", ID, res.Status)
	default:
		res.Order = ID
		res.Status = "PROCESSED"
		res.Accrual = i * 100
		resStatus = http.StatusOK
		log.Printf("order: %s, status: %s, accr:%d", ID, res.Status, res.Accrual)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resStatus)
	if resStatus == http.StatusOK {
		buf, err := json.MarshalIndent(res, "", "  ")
		if err != nil {
			log.Fatal("error marshal json")
		}
		w.Write(buf)
	}
}
