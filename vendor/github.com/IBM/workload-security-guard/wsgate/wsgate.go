package wsgate

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/IBM/go-security-plugs/iofilter"
	pi "github.com/IBM/go-security-plugs/pluginterfaces"
	"github.com/IBM/workload-security-guard/spec"
)

const version string = "0.0.1"
const name string = "wsgate"

type plug struct {
	name    string
	version string

	// Add here any other state the extension needs
	guardUrl   string
	id         string
	gateConfig spec.WsGate
	httpc      http.Client
	cycle      int
}

type minmaxFloat32 struct {
	L float32
	H float32
}

type minmaxUint16 struct {
	L uint16
	H uint16
}

//type wsgateConfig struct {
//	QsKeys         []minmaxUint16
//	ProcessingTime []minmaxUint16
//}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))[0:16]
}

func (p *plug) Shutdown() {
	pi.Log.Infof("%s: Shutdown", p.name)
}

func (p *plug) PlugName() string {
	return p.name
}

func (p *plug) PlugVersion() string {
	return p.version
}

func ReadUserIP(req *http.Request) string {
	IPAddress := req.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = req.Header.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress = req.RemoteAddr
	}
	return IPAddress
}

func (p *plug) screenRequest(req *http.Request) error {
	// Request client and server identities
	cip, cport, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return fmt.Errorf("illegal req.RemoteAddr %s", err.Error())
	}
	sip, sport, err := net.SplitHostPort(req.URL.Host)

	if err != nil {
		return fmt.Errorf("illegal req.URL.Host %s", err.Error())
	}
	pi.Log.Infof("Client: %s port %s", cip, cport)
	pi.Log.Infof("Server: %s port %s", sip, sport)

	// Request principles
	pi.Log.Infof("req.Method %s", req.Method)
	pi.Log.Infof("req.Proto %s", req.Proto)
	pi.Log.Infof("scheme: %s", req.URL.Scheme)
	pi.Log.Infof("opaque: %s", req.URL.Opaque)

	pi.Log.Infof("ContentLength: %d", req.ContentLength)
	pi.Log.Infof("Trailer: %#v", req.Trailer)

	// TBD req.Form

	rp := new(spec.ReqProfile)
	//rp.Profile(req, &p.gateConfig.Req)
	rp.Profile(req)
	decission := rp.Decide(&p.gateConfig.Req)
	if decission != "" {

		if !p.gateConfig.ConsultGuard {
			pi.Log.Infof("Gate Decission: %s", decission)
			p.reportBlock(rp, decission)
			return errors.New(decission)
		}
		decission = p.consult(rp)
		if decission != "" {
			pi.Log.Infof("Guard Decission:", decission)
			p.reportBlock(rp, decission)
			return errors.New(decission)
		}
	}
	p.aggregate(rp)

	/*
		//decoded path
		path := req.URL.Path
		pathProfile := p.gateConfig.ProfilePath(path)
		pi.Log.Infof("path profile %v", pathProfile)

		//decoded query string
		query := req.URL.Query()
		queryProfile := p.gateConfig.ProfileQueryString(query)
		pi.Log.Infof("query: %#v", queryProfile)

		//http headers
		hProfile := p.gateConfig.ProfileHttpHeaders(req.Header)

		pi.Log.Infof("Headers: %#v", hProfile)

		//http Trailers
		trailerProfile := spec.ProfileKeyVals(req.Trailer)
		pi.Log.Infof("Trailer: %#v", trailerProfile)
	*/
	/*
		// fingerprints representing the sender of the request and the event to be processed
		symbols := make([]string, 0, 12)
		symbols = append(symbols,
			req.Method,
			req.Proto,
			u.Scheme,
			opaque,

		// fingerprints representing the sender of the request and the event to be processed
		fingerprints := make([]string, 0, 12)
		fingerprints = append(fingerprints,
			pathSplits[0],
			GetMD5Hash(queryKeys),
			headers.Get("Transfer-Encoding"),
			headers.Get("Content-Encoding"),
			headers.Get("Keep-Alive"),
			headers.Get("Connection"),
			headers.Get("X-Forwarded-For"),
			headers.Get("Cache-Control"),
			headers.Get("Via"),
			acceptHeaderVals,
			contentHeaderVals,
			userAgentVals,
			allHeaderKeys,
			httpinfo["protocol"])
		pi.Log(fingerprints)
		for i, val := range fingerprints {
			fingerprints[i] = GetMD5Hash(val)
		}
	*/
	/*


		contentEncoding := r.Header.Get("content-encoding")
		transferEncoding := r.Header.Get("transfer-encoding")
		keepAlive := r.Header.Get("keep-alive")
		connection := r.Header.Get("Connection")
		xForwardedFor := r.Header.Get("x-forwarded-for")
		cacheControl := r.Header.Get("cache-control")
		via := r.Header.Get("via")

		log.Info("DH> userAgentVals ", userAgentVals)
		log.Info("DH> contentEncoding ", contentEncoding)
		log.Info("DH> transferEncoding ", transferEncoding)
		log.Info("DH> keepAlive ", keepAlive)
		log.Info("DH> connection ", connection)
		log.Info("DH> xForwardedFor ", xForwardedFor)
		log.Info("DH> cacheControl ", cacheControl)
		log.Info("DH> via ", via)
	*/
	//var d = new Date();
	//h := make(map[string]string)

	//markers := make([]float32, 0, 12)
	//integers := make([]int, 0, 12)
	//roundedMarkers := make([]float32, 0, 12)
	//histograms := make([][]int, 0, 12)

	// Create a sorted slice of all header keys

	// Create a sorted slice of all query leys

	/*
		roundedMarkers.append(fingerprints, d.getDay()/6)
		roundedMarkers.append(fingerprints, d.getHours()/23)
		console.log(roundedMarkers)

		console.log(httpreq.body)
		console.log(otherHeaderVals)
		console.log(queryContent)


		integers.append(integers, parseInt(httpreq.size)) // Content-Length  - size of body
		integers.append(integers, otherHeaderVals.length)
		integers.append(integers, queryContent.length)
		integers.append(integers, cookieVals.length)
		integers.append(integers, pathSplits[0].length)
		integers.append(integers, allHeaderVals.length)
		console.log(markers, integers)



		histograms.append(histograms, hist(httpreq.body))
		histograms.append(histograms, hist(otherHeaderVals))
		histograms.append(histograms, hist(queryContent))
		histograms.append(histograms, hist(cookieVals))
		histograms.append(histograms, hist(allHeaderVals))
		console.log(histograms)

		fingerprint_path= pathSplits[0]


		var triggerInstance = headers["x-request-id"]||uuid.v4()




		const dataout = JSON.stringify({
					gateId:   gate
				, serviceId: unit
				, triggerInstance: triggerInstance
				, data: {
						fingerprints: fingerprints
					, markers: markers
					, integers: integers
					, roundedMarkers: roundedMarkers
					, histograms: histograms
				}
			});

		console.log(unit, dataout);
		postRequest("Path: "+fingerprint_path, "/eval", dataout, callback)


	*/

	return nil
}

func (p *plug) screenResponse(resp *http.Response) error {
	return nil
}

func responseFilter(buf []byte) error {
	h := make([]int, 8)

	for _, c := range buf {
		switch {
		case (c >= 97 && c <= 122) || (c >= 48 && c <= 57) || (c == 32): //a..z, 0..9, <SPACE>
			h[0]++
		case c >= 127 || c <= 31: // All non ascii unicodes, ascii 0-31, <DEL>
			h[1]++
		case c == 34 || c == 96 || c == 39: // ascii quatations  - TBD IF NEED TO BE extended with other suspects
			h[2]++
		case c == 59: // ; - TBD IF NEED TO BE extended with other suspects
			h[3]++
		default: // anything else: !#$%&()*+,-./:<=>?@[\]^_{|}~
			h[7]++

		}
	}
	fmt.Printf("responseFilter Histogram: %v\n", h)

	return nil
}

func requestFilter(buf []byte) error {
	h := make([]int, 8)

	for _, c := range buf {
		switch {
		case (c >= 97 && c <= 122) || (c >= 48 && c <= 57) || (c == 32): //a..z, 0..9, <SPACE>
			h[0]++
		case c >= 127 || c <= 31: // All non ascii unicodes, ascii 0-31, <DEL>
			h[1]++
		case c == 34 || c == 96 || c == 39: // ascii quatations  - TBD IF NEED TO BE extended with other suspects
			h[2]++
		case c == 59: // ; - TBD IF NEED TO BE extended with other suspects
			h[3]++
		default: // anything else: !#$%&()*+,-./:<=>?@[\]^_{|}~
			h[7]++

		}
	}
	fmt.Printf("requestFilter Histogram: %v\n", h)

	return nil
}

func (p *plug) ApproveRequest(req *http.Request) (*http.Request, error) {
	testBodyHist := true

	pi.Log.Infof("%s: ApproveRequest started", p.name)

	if req.Header.Get("X-Block-Req") != "" {
		pi.Log.Infof("%s ........... Blocked During Request! returning an error!", p.name)
		return nil, errors.New("request blocked")
	}

	for name, values := range req.Header {
		// Loop over all values for the name.
		for _, value := range values {
			pi.Log.Infof("%s Request Header: %s: %s", p.name, name, value)
		}
	}

	if p.screenRequest(req) != nil {
		return nil, errors.New("secuirty blocked")
	}

	newCtx, cancelFunction := context.WithCancel(req.Context())
	req = req.WithContext(newCtx)

	timeoutStr := req.Header.Get("X-Block-Async")
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		timeoutStr = "5s"
		timeout, _ = time.ParseDuration(timeoutStr)
	}

	if testBodyHist && req.Body != nil {
		fmt.Printf("Analyze Start\n")

		// Asynchrniously stream bytes from the original resp.Body
		// to a new resp.Body
		req.Body = iofilter.New(req.Body, requestFilter)
	}

	pi.Log.Infof("%s ........... will asynchroniously block after %s", p.name, timeoutStr)
	go func(newCtx context.Context, cancelFunction context.CancelFunc, req *http.Request, timeout time.Duration) {
		select {
		case <-newCtx.Done():
			pi.Log.Infof("Done! %v", newCtx.Err())
		case <-time.After(timeout):
			pi.Log.Infof("Timeout!")
			cancelFunction()
		}
	}(newCtx, cancelFunction, req, timeout)

	return req, nil
}

func (p *plug) ApproveResponse(req *http.Request, resp *http.Response) (*http.Response, error) {
	testBodyHist := true

	pi.Log.Infof("%s: ApproveResponse started", p.name)

	if req.Header.Get("X-Block-Resp") != "" {
		pi.Log.Infof("%s ........... Blocked During Response! returning an error!", p.name)
		return nil, errors.New("response blocked")
	}

	for name, values := range resp.Header {
		// Loop over all values for the name.
		for _, value := range values {
			pi.Log.Infof("%s Response Header: %s: %s", p.name, name, value)
		}
	}

	if p.screenResponse(resp) != nil {
		return nil, errors.New("secuirty blocked")
	}

	if testBodyHist && resp.Body != nil {
		fmt.Printf("Analyze Start\n")

		// Asynchrniously stream bytes from the original resp.Body
		// to a new resp.Body
		resp.Body = iofilter.New(resp.Body, responseFilter)
	}

	return resp, nil
}

func (p *plug) fetchConfig() {

	req, err := http.NewRequest(http.MethodGet, p.guardUrl+"/fetchConfig", nil)
	if err != nil {
		pi.Log.Infof("wsgate getConfig: http.NewRequest error %v", err)
	}
	query := req.URL.Query()
	query.Add("sid", p.id)
	req.URL.RawQuery = query.Encode()
	res, getErr := p.httpc.Do(req)
	if getErr != nil {
		pi.Log.Infof("wsgate getConfig: httpc.Do error %v", getErr)
		return
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		pi.Log.Infof("wsgate getConfig: http.NewRequest error %v", readErr)
	}

	//pi.Log.Infof("wsgate getConfig: unmarshal %s", string(body))
	jsonErr := json.Unmarshal(body, &p.gateConfig)
	if jsonErr != nil {
		pi.Log.Infof("wsgate getConfig: unmarshel error %v", jsonErr)
	}

	//pi.Log.Infof("wsgate getConfig: ended %v ", p.gateConfig)
}

func (p *plug) consult(req *spec.ReqProfile) string {
	postBody, marshalErr := json.Marshal(req)
	if marshalErr != nil {
		log.Fatalf("An Error Occured %v", marshalErr)
		return fmt.Sprintf("Ilegalg consult %v", marshalErr)
	}
	reqBody := bytes.NewBuffer(postBody)
	res, postErr := p.httpc.Post(p.guardUrl+"/consult", "application/json", reqBody)
	if postErr != nil {
		pi.Log.Infof("wsgate getConfig: httpc.Do error %v", postErr)
		return fmt.Sprintf("Guard unavaliable during consult %v", postErr)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		pi.Log.Infof("wsgate consult: cant read result error %v", readErr)
		return fmt.Sprintf("Guard ilegal response during consult %v", readErr)
	}
	if len(body) == 0 {
		pi.Log.Infof("wsgate consult: response is %s", string(body))
		return fmt.Sprintf("Guard: %s", string(body))
	}
	pi.Log.Infof("wsgate consult: approved!")
	return ""
}

func (p *plug) reportBlock(req *spec.ReqProfile, decission string) {

}

func (p *plug) aggregate(req *spec.ReqProfile) {
	p.cycle--
	if p.cycle <= 0 {
		//p.ReportToGuard()
		p.cycle = 100
	}
}

func (p *plug) Init() {
	pi.Log.Infof("plug %s: Initializing - version %v", p.name, p.version)

	p.guardUrl = os.Getenv("WSGATE_GUARD_URL")
	if p.guardUrl == "" {
		p.guardUrl = "http://ws.knative-guard"
	}

	servingNamespace := os.Getenv("SERVING_NAMESPACE")
	if servingNamespace == "" {
		panic("Cant find SERVING_NAMESPACE")
	}
	servingService := os.Getenv("SERVING_SERVICE")
	if servingService == "" {
		panic("Cant find SERVING_SERVICE")
	}
	p.id = servingService + "." + servingNamespace

	p.fetchConfig()
}

func init() {
	p := new(plug)
	p.version = version
	p.name = name
	pi.RegisterPlug(p)
}
