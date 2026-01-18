package main

//
//import (
//	"bytes"
//	"encoding/json"
//	"fmt"
//	"github.com/prometheus/client_golang/prometheus"
//	"github.com/prometheus/client_golang/prometheus/collectors"
//	"github.com/prometheus/client_golang/prometheus/promhttp"
//	"log"
//	"math/rand"
//	"net/http"
//	"time"
//)
//
//const (
//	testAddr = "192.168.77.38:9123"
//)
//
//var httpStatusCodeCounter = prometheus.NewCounterVec(
//	prometheus.CounterOpts{
//		Name: "http_status",
//		Help: "Count http status code",
//	},
//	[]string{"status_code"})
//
//type request struct {
//	StatusCode string
//}
//
//func main() {
//	go produceData()
//	reg := prometheus.NewRegistry()
//	prometheus.WrapRegistererWith(prometheus.Labels{
//		"serviceName": "demo-service",
//	}, reg).MustRegister(
//		collectors.NewGoCollector(),
//		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
//		httpStatusCodeCounter,
//	)
//
//	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
//	http.HandleFunc("/", sendMetricsHandle)
//	log.Fatal(http.ListenAndServe(testAddr, nil))
//}
//
//func sendMetricsHandle(writer http.ResponseWriter, r *http.Request) {
//	var req request
//	defer func() {
//		httpStatusCodeCounter.WithLabelValues(req.StatusCode).Inc()
//		log.Printf("add 1 to %s", req.StatusCode)
//	}()
//	_ = json.NewDecoder(r.Body).Decode(&req)
//	log.Printf("receive req:%+v", req)
//	_, _ = writer.Write([]byte(req.StatusCode))
//}
//
//func produceData() {
//	codes := []string{"503", "404", "400", "200", "304", "500"}
//	for {
//		body, _ := json.Marshal(request{
//			StatusCode: codes[rand.Intn(len(codes))],
//		})
//		requestBody := bytes.NewBuffer(body)
//		http.Post(fmt.Sprintf("http://%s/", testAddr), "application/json", requestBody)
//		log.Printf("send request=%s to %s", requestBody.String(), testAddr)
//		time.Sleep(2 * time.Second)
//	}
//}
