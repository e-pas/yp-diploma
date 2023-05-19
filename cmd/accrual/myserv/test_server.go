package main

import (
	"encoding/json"
	"flag"
	"log"
	"math/rand"
	"net/http"
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
	log.Println("Listen on: ", listen)
	http.ListenAndServe(listen, r)
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
