package pooller

import (
	"helpprovider_snmp/snmp"
	"log"
	"time"
)

func backgroundWorker(w *Worker) {
	for {
		request :=<- w.RequestQueue
		log.Printf("New request in worker for requestId:%v ",request.UUid)
		if request.RequestBody.UseCache && request.Type != requestSet {
			exist, resp := w.getCacheResponseFromDevice(request)
			if exist  {
				resp.ResponseBody.FromCache = true
				resp.UUid = request.UUid
				w.ResponseQueue <- resp
				continue
			}
		}
		//Try initialize snmp
		err, conn := snmp.Connect(snmp.InitStruct{
			Ip: request.RequestBody.Ip,
			Community: request.RequestBody.Community,
			Repeats: request.RequestBody.Repeats,
			TimeoutSec: time.Duration(request.RequestBody.Timeout) * time.Second,
			Version: snmp.Version2c,
		})

		//If error when connecting - continue this iteration
		if err != nil {
			request.ResponseBody = Response{
				Ip: request.RequestBody.Ip,
				Response: nil,
				Error: err.Error(),
				FromCache: false,
				Oid: request.RequestBody.Oid,
			}
			w.ResponseQueue <- request
			continue
		}


		var resp []snmp.SnmpResp
		switch request.Type {
		case requestGet:
			err, resp =  conn.Get(request.RequestBody.Oid)
		case requestWalk:
			err, resp =  conn.Walk(request.RequestBody.Oid)
		case requestBulkWalk:
			err, resp =  conn.WalkBulk(request.RequestBody.Oid)
		case requestSet:
			err, resp =  conn.Set(request.RequestBody.Oid, request.RequestBody.Type, request.RequestBody.Value)
		}

		if err != nil {
			request.ResponseBody = Response{
				Ip: request.RequestBody.Ip,
				Response: nil,
				Error: err.Error(),
				FromCache: false,
				Oid: request.RequestBody.Oid,
			}
		} else {
			request.ResponseBody = Response{
				Ip: request.RequestBody.Ip,
				Response: resp,
				Error: "",
				FromCache: false,
				Oid: request.RequestBody.Oid,
			}
			w.setCacheResponseFromDevice(request)
		}
		w.ResponseQueue <- request
		conn.Close()
	}
}

func workerResponseCollector(w *Worker) {
	for {
		response :=<- w.ResponseQueue
		log.Printf("New response in collector for requestId: %v ",response.UUid)
		w.deleteCountFromRequest(response.UUid)
		w.deleteRequestForSwitch(response.RequestBody.Ip)
		w.addRequestData(response.UUid, response)
		log.Printf("Sended ressponse to RequestData with id: %v ",response.UUid)
	}
}
