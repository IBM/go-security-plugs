package wsgate

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/IBM/go-security-plugs/iofilter"
	pi "github.com/IBM/go-security-plugs/pluginterfaces"
)

const version string = "0.0.1"
const name string = "wsgate"

type plug struct {
	name    string
	version string

	// Add here any other state the extension needs
	config           map[string]string
	guardUrl         string
	servingNamespace string
	servingService   string
	gateConfig       wsgateConfig
}

type minmaxFloat32 struct {
	L float32
	H float32
}

type minmaxUint16 struct {
	L uint16
	H uint16
}

type wsgateConfig struct {
	QsKeys         []minmaxUint16
	ProcessingTime []minmaxUint16
}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))[0:16]
}

func (p *plug) getSortedKeys(m map[string][]string) (sKeys []string) {
	sKeys = make([]string, len(m))
	i := 0
	for k := range m {
		sKeys[i] = k
		i++
	}
	sort.Strings(sKeys)
	pi.Log.Infof("WSGate: sorted keys %s", sKeys)
	return
}

func (p *plug) getSortedVals(m map[string][]string, sKeys []string) (sVals []string) {
	sVals = make([]string, len(m))
	for i, k := range sKeys {
		sVals[i] = strings.Join(m[k], " ")
	}
	pi.Log.Infof("WSGate: sorted vals %s", sVals)
	return
}

func (p *plug) hist(str string) []int {
	h := make([]int, 8)

	str = strings.ToLower(str)
	for _, c := range str {
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

	h[4] = strings.Count(str, "/*") + strings.Count(str, "*/") + strings.Count(str, "--") + strings.Count(str, "[]") //why -- and []

	h[5] = strings.Count(str, "0x")
	h[6] = strings.Count(str, "select") + strings.Count(str, "delete") + strings.Count(str, "drop") + strings.Count(str, "from") + strings.Count(str, "where")
	fmt.Printf("Histogram: %v", h)
	return h
}

func (p *plug) Shutdown() {
	pi.Log.Infof("%s: Shutdown", p.name)
	if p.config["panicShutdown"] == "true" {
		panic("it is fun to panic everywhere! also in Shutdown")
	}
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
	var acceptHeaderVals, contentHeaderVals, userAgentVals strings.Builder
	var allHeaderKeys, otherHeaderKeys, otherHeaderVals, allHeaderVals, cookieVals strings.Builder

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

	//url path
	path := req.URL.Path
	pathhist := p.hist(path)
	pathSplits := strings.Split(path, "/")
	pi.Log.Infof("path aplits %v hist %v", pathSplits, pathhist)

	//url quesy string
	query := req.URL.Query()
	qkeys := p.getSortedKeys(query)
	qvals := p.getSortedVals(query, qkeys)
	qvalstr := strings.Join(qvals, " ")
	qvalhist := p.hist(qvalstr)
	pi.Log.Infof("query: %#v", query)
	pi.Log.Infof("queryKeys: %s", strings.Join(qkeys, " "))
	pi.Log.Infof("queryVals: %s %v", qvalstr, qvalhist)

	//http headers
	hkeys := p.getSortedKeys(req.Header)

	// Construct data about header keys and header vals
	for _, k := range hkeys {
		val := ""
		if vals, ok := req.Header[k]; ok {
			val = strings.Join(vals, " ")
		}

		allHeaderKeys.WriteString(k)
		allHeaderVals.WriteString(val)
		allHeaderKeys.WriteString(" ")
		allHeaderVals.WriteString(" ")
		switch {
		case strings.HasPrefix(k, "Accept"):
			acceptHeaderVals.WriteString(val)
		case strings.HasPrefix(k, "Content"):
			contentHeaderVals.WriteString(val)
		case k == "User-Agent":
			userAgentVals.WriteString(val)
		case strings.HasPrefix(k, "Cookie"):
			cookieVals.WriteString(val)
		default:
			otherHeaderVals.WriteString(val)
			otherHeaderKeys.WriteString(k)
		}
	}
	pi.Log.Infof("Headers: %#v", req.Header)
	pi.Log.Infof("WSGate: allHeaderKeys: %s", allHeaderKeys.String())
	pi.Log.Infof("WSGate: allHeaderVals: %s", allHeaderVals.String())
	pi.Log.Infof("WSGate: acceptHeaderVals: %s", acceptHeaderVals.String())
	pi.Log.Infof("WSGate: contentHeaderVals: %s", contentHeaderVals.String())
	pi.Log.Infof("WSGate: userAgentVals: %s", userAgentVals.String())
	pi.Log.Infof("WSGate: cookieVals: %s", cookieVals.String())
	pi.Log.Infof("WSGate: otherHeaderVals: %s", otherHeaderVals.String())
	pi.Log.Infof("WSGate: otherHeaderKeys: %s", otherHeaderKeys.String())

	//http Trailers
	tkeys := p.getSortedKeys(req.Trailer)
	tvals := p.getSortedVals(req.Trailer, tkeys)
	pi.Log.Infof("query: %#v", query)
	pi.Log.Infof("Trailer Keys: %s", strings.Join(tkeys, " "))
	pi.Log.Infof("Trailer Vals: %s", strings.Join(tvals, " "))

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

func responseFilter(buf []byte, state *interface{}) {
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

}

func requestFilter(buf []byte, state *interface{}) {
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
}

func (p *plug) ApproveRequest(req *http.Request) (*http.Request, error) {
	testBodyHist := true

	pi.Log.Infof("%s: ApproveRequest started", p.name)
	if p.config["panicReq"] == "true" {
		panic("it is fun to panic everywhere! also in ApproveRequest")
	}

	if p.config["errorReq"] != "" {
		return nil, errors.New(p.config["errorReq"])
	}

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
	if p.config["panicResp"] == "true" {
		panic("it is fun to panic everywhere! also in ApproveResponse")
	}

	if p.config["errorResp"] != "" {
		return nil, errors.New(p.config["errorResp"])
	}

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
	httpc := http.Client{}
	req, err := http.NewRequest(http.MethodGet, p.guardUrl+"/fetchConfig", nil)
	if err != nil {
		pi.Log.Infof("wsgate getConfig: http.NewRequest error %v", err)
	}
	query := req.URL.Query()
	query.Add("ns", p.servingNamespace)
	query.Add("srv", p.servingService)
	req.URL.RawQuery = query.Encode()
	res, getErr := httpc.Do(req)
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

	pi.Log.Infof("wsgate getConfig: body will be unmarsheled")
	jsonErr := json.Unmarshal(body, &p.gateConfig)
	if jsonErr != nil {
		pi.Log.Infof("wsgate getConfig: unmarshel error %v", jsonErr)
	}
	pi.Log.Infof("wsgate getConfig: ended %v ", p.gateConfig)
}

func (p *plug) Init() {
	pi.Log.Infof("plug %s: Initializing - version %v", p.name, p.version)

	p.config = make(map[string]string)
	p.config["panicInitialize"] = os.Getenv("WS_GATE_PANIC_INIT")
	p.config["panicShutdown"] = os.Getenv("WS_GATE_PANIC_SHUTDOWN")
	p.config["panicReq"] = os.Getenv("WS_GATE_PANIC_REQ")
	p.config["panicResp"] = os.Getenv("WS_GATE_PANIC_RESP")
	p.config["errorReq"] = os.Getenv("WS_GATE_ERROR_REQ")
	p.config["errorResp"] = os.Getenv("WS_GATE_ERROR_RESP")

	pi.Log.Infof("Plug.config %v", p.config)

	if p.config["panicInitialize"] == "true" {
		panic("it is fun to panic everywhere! also in Initialize")
	}
	p.guardUrl = os.Getenv("WSGATE_GUARD_URL")
	if p.guardUrl == "" {
		p.guardUrl = "http://ws.knative-guard"
	}

	p.servingNamespace = os.Getenv("SERVING_NAMESPACE")
	if p.servingNamespace == "" {
		panic("Cant find SERVING_NAMESPACE")
	}
	p.servingService = os.Getenv("SERVING_SERVICE")
	if p.servingService == "" {
		panic("Cant find SERVING_SERVICE")
	}

	p.fetchConfig()
}

func init() {
	p := new(plug)
	p.version = version
	p.name = name
	pi.RegisterPlug(p)
}
