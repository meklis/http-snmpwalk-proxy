package pooller

import (
	"helpprovider_snmp/snmp"
	"time"
)

func backgroundWorker(w *Worker, numWorker uint) {
	for {
		request :=<- w.requestQueue
		w.Logger.DebugF("Received request %v in background worker %v", request.UUid, numWorker)
		if request.RequestBody.UseCache && request.Type != requestSet {
			exist, resp := w.getCacheResponseFromDevice(request)
			if exist  {
				w.Logger.DebugF("Find response for %v with device %v, with oid %v and type %v in cache", request.UUid, request.RequestBody.Ip, request.RequestBody.Oid, request.Type)
				resp.ResponseBody.FromCache = true
				resp.UUid = request.UUid
				w.responseQueue <- resp
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
			w.Logger.WarningF("Problem create snmp connection for request %v with err: %v", request.UUid, err.Error())
			request.ResponseBody = Response{
				Ip: request.RequestBody.Ip,
				Response: nil,
				Error: err.Error(),
				FromCache: false,
				Oid: request.RequestBody.Oid,
			}
			w.responseQueue <- request
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
			w.Logger.WarningF("Error in request %v, switch %v: %v", request.UUid, request.RequestBody.Ip, err.Error())
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
		w.responseQueue <- request
		conn.Close()
	}
}

func workerResponseCollector(w *Worker, numWorker uint) {
	for {
		response :=<- w.responseQueue
		w.Logger.DebugF("New response in collector with ip %v and requestId: %v",response.RequestBody.Ip, response.UUid)
		w.deleteCountFromRequest(response.UUid)
		w.deleteRequestForSwitch(response.RequestBody.Ip)
		w.addRequestData(response.UUid, response)
		w.Logger.DebugF("Sended response to RequestData with ip %v and requestId: %v ",response.RequestBody.Ip,response.UUid)
	}
}
