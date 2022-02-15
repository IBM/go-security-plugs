package wsgate

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	spec "github.com/IBM/workload-security-guard/pkg/apis/wsecurity/v1"
	guardianclientset "github.com/IBM/workload-security-guard/pkg/generated/clientset/guardians"
	v1 "github.com/IBM/workload-security-guard/pkg/generated/clientset/guardians/typed/wsecurity/v1"

	"github.com/IBM/go-security-plugs/iofilter"
	pi "github.com/IBM/go-security-plugs/pluginterfaces"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const version string = "0.0.1"
const name string = "wsgate"

type plug struct {
	name    string
	version string

	// Add here any other state the extension needs
	guardUrl  string
	namespace string
	serviceId string
	//kClient   corev1.ConfigMapInterface
	gClient v1.WsecurityV1Interface
	wsGate  *spec.WsGate
	//blocked             []string
	//numOk               uint32
	lastConsultReported time.Time
	numConsultsCount    uint16
	httpc               http.Client
	cycle               int
}

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
	rp.Profile(req)
	decission := p.wsGate.Req.Decide(rp)
	if decission != "" {
		// potentially consult guard before rejecting
		pi.Log.Infof("Guardian refused to allow: %s", decission)
		if p.wsGate.ConsultGuard.Active {
			minuete := time.Now().Truncate(time.Minute)
			if p.lastConsultReported != minuete {
				p.lastConsultReported = minuete
				p.numConsultsCount = 0
			}
			if p.numConsultsCount < p.wsGate.ConsultGuard.RequestsPerMinuete {
				p.numConsultsCount = p.numConsultsCount + 1
				pi.Log.Infof("Consulting Guard %d/%d", p.numConsultsCount, p.wsGate.ConsultGuard.RequestsPerMinuete)
				decission = p.consultOnRequest(rp)
				//pi.Log.Infof("Guard said: %s", decission)
			}
		}
	}
	if decission == "" {
		p.reportAllow(rp)
	} else {
		pi.Log.Infof("Decission: %s", decission)
		p.reportBlock(rp, decission)
		if !p.wsGate.ForceAllow {
			return errors.New(decission)
		}
	}

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
		//fmt.Printf("Analyze Start\n")

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
		//fmt.Printf("Analyze Start\n")

		// Asynchrniously stream bytes from the original resp.Body
		// to a new resp.Body
		resp.Body = iofilter.New(resp.Body, responseFilter)
	}

	return resp, nil
}

func (p *plug) consultOnRequest(reqProfile *spec.ReqProfile) string {
	postBody, marshalErr := json.Marshal(reqProfile)
	if marshalErr != nil {
		log.Printf("consultOnRequest error during marshal: %v", marshalErr)
		return fmt.Sprintf("Cant marshal in consultOnRequest %v", marshalErr)
	}
	reqBody := bytes.NewBuffer(postBody)
	req, err := http.NewRequest(http.MethodPost, p.guardUrl+"/req", reqBody)
	if err != nil {
		pi.Log.Infof("wsgate consultOnRequest: http.NewRequest error %v", err)
	}
	query := req.URL.Query()
	query.Add("sid", p.serviceId)
	query.Add("ns", p.namespace)
	req.URL.RawQuery = query.Encode()

	res, postErr := p.httpc.Do(req)
	//res, postErr := p.httpc.Post(p.guardUrl+"/req", "application/json", reqBody)
	if postErr != nil {
		pi.Log.Infof("wsgate consultOnRequest: httpc.Do error %v", postErr)
		return fmt.Sprintf("Guard unavaliable during consult %v", postErr)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		pi.Log.Infof("wsgate consultOnRequest: response error %v", readErr)
		return fmt.Sprintf("Guard ilegal response during consult %v", readErr)
	}
	if len(body) != 0 {
		pi.Log.Infof("wsgate consultOnRequest: response is %s", string(body))
		return fmt.Sprintf("Guard: %s", string(body))
	}
	pi.Log.Infof("wsgate consultOnRequest: approved!")
	return ""
}

func (p *plug) reportBlock(req *spec.ReqProfile, decission string) {
	// build statistics on blocked requests
	p.cycle--
	if p.cycle <= 0 {
		//p.ReportToGuard()
		p.cycle = 100
	}
}

func (p *plug) reportAllow(req *spec.ReqProfile) {
	// build statistics on allowed requests
	p.cycle--
	if p.cycle <= 0 {
		//p.ReportToGuard()
		p.cycle = 100
	}
}

func (p *plug) initCrd() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	cfg, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	guardianClient, err := guardianclientset.NewForConfig(cfg)
	if err != nil {
		panic(err.Error())
	}

	p.gClient = guardianClient.WsecurityV1()

	//clientset, err := kubernetes.NewForConfig(cfg)
	//if err != nil {
	//	panic(err.Error())
	//}
	//p.kClient = clientset.CoreV1().ConfigMaps("knative-serving")

}

//
func (p *plug) readCrd(namespace string, serviceId string) *spec.GuardianSpec {
	g, err := p.gClient.Guardians(namespace).Get(context.TODO(), serviceId, metav1.GetOptions{})
	if err != nil {
		pi.Log.Infof("Err during get %s.%s: %s\n", serviceId, namespace, err.Error())
		//panic(fmt.Sprintf("No Guardian! for %s.%s", serviceId, namespace))
		return nil
	}
	pi.Log.Infof("Found guardian %s.%s\n", serviceId, namespace)
	fmt.Print((*spec.WsGate)(g.Spec).Marshal(0))
	return g.Spec
}

/*
func (p *plug) readConfigMap() {
	cm, err := p.kClient.Get(context.TODO(), "guardian", metav1.GetOptions{})
	if err != nil {
		fmt.Printf("ConfigMap Error: %v\n", err)
		panic(err.Error())
	}
	p.blockByDefault = cm.Data["BlockByDefault"] != "false"
	fmt.Printf("ConfigMap: %s\n", cm.Data["BlockByDefault"])

	//	err = json.Unmarshal([]byte(cm.Data["guardian"]), &p.wsGate)
	//	if err != nil {
	//		fmt.Printf("ConfigMap Unmarshal Error: %v\n", err)
	//		panic(err.Error())
	//	}
}
*/
func (p *plug) fetchConfig() {
	//p.readConfigMap()
	gurdianSpec := p.readCrd(p.namespace, p.serviceId)
	if gurdianSpec == nil {
		gurdianSpec = p.readCrd("knative-serving", "guardian")
	}
	if gurdianSpec == nil {
		gurdianSpec = new(spec.GuardianSpec)
		// default gurdianSpec has:
		// 		gurdianSpec.falseAllow=false
		// 		gurdianSpec.ConsultGuard.Active = false
	}
	p.wsGate = (*spec.WsGate)(gurdianSpec)

	/*
		req, err := http.NewRequest(http.MethodGet, p.guardUrl+"/config", nil)
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
			pi.Log.Infof("wsgate getConfig: read body error %v", readErr)
		}

		//pi.Log.Infof("wsgate getConfig: unmarshal %s", string(body))
		jsonErr := json.Unmarshal(body, &p.gateConfig)
		if jsonErr != nil {
			pi.Log.Infof("wsgate getConfig: unmarshel error %v", jsonErr)
		}

		//pi.Log.Infof("wsgate getConfig: ended %v ", p.gateConfig)
	*/
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
	p.serviceId = servingService
	p.namespace = servingNamespace

	p.initCrd()
	p.fetchConfig()
}

func init() {
	fmt.Printf("WSGATE!!!! Initializing!!!!!!!!!<<<<<<<<<<<__________________>>>>>>>>>>\n")
	p := new(plug)
	p.version = version
	p.name = name
	pi.RegisterPlug(p)
	fmt.Printf("WSGATE!!!! Ended Initializing!!!!!!!!!<<<<<<<<<<<__________________>>>>>>>>>>\n")
}
