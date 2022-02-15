package spec

import (
	"fmt"
	"net/http"
)

type WsGate struct {
	Req          ReqConfig
	ConsultGuard bool
}

type ReqConfig struct {
	Url     UrlConfig
	Qs      QueryConfig
	Headers HeadersConfig
}

type ReqProfile struct {
	Url     *UrlProfile
	Query   *QueryProfile
	Headers *HeadersProfile
}

type UrlProfile struct {
	Val *SimpleValProfile
}
type UrlConfig struct {
	Val SimpleValConfig
}
type QueryProfile struct {
	Kv *KeyValProfile
}

type QueryConfig struct {
	Kv KeyValConfig
}

type HeadersProfile struct {
	Kv *KeyValProfile
}

type HeadersConfig struct {
	Kv KeyValConfig
}

func (u *UrlProfile) Profile(path string) {
	u.Val = new(SimpleValProfile)
	u.Val.Profile(path)
}

func (u *UrlProfile) Decide(config *UrlConfig) string {
	str := u.Val.Decide(&config.Val)
	if str == "" {
		return str
	}
	return fmt.Sprintf("URL: %s", str)
}

// Allow typical URL values - use for development but not in production
func (config *UrlConfig) AddTypicalVal() {
	for i := 0; i < 8; i++ {
		config.Val.L_Counters[i] = []uint8{0}
		config.Val.H_Counters[i] = []uint8{0}
	}

	config.Val.H_Counters[BasicTotalCounter][0] = 128
	config.Val.H_Counters[BasicLetterCounter][0] = 128
	config.Val.H_Counters[BasicDigitCounter][0] = 128
	config.Val.H_Counters[BasicWordCounter][0] = 16
	config.Val.H_Counters[BasicNumberCounter][0] = 16
	config.Val.H_Counters[BasicPathCounter][0] = 8
	config.Val.Flags = 1 << DivSlot
}

func (q *QueryProfile) Profile(m map[string][]string) {
	q.Kv = new(KeyValProfile)
	q.Kv.Profile(m)
}

func (q *QueryProfile) Decide(config *QueryConfig) string {
	str := q.Kv.Decide(&config.Kv)
	if str == "" {
		return str
	}
	return fmt.Sprintf("QueryString: %s", str)
}

// Allow typical query string values - use for development but not in production
func (config *QueryConfig) AddTypicalVal() {
	config.Kv.OtherKeynames = new(SimpleValConfig)
	for i := 0; i < 8; i++ {
		config.Kv.OtherKeynames.L_Counters[i] = []uint8{0}
		config.Kv.OtherKeynames.H_Counters[i] = []uint8{0}
	}

	config.Kv.OtherKeynames.H_Counters[BasicTotalCounter][0] = 16
	config.Kv.OtherKeynames.H_Counters[BasicLetterCounter][0] = 16
	config.Kv.OtherKeynames.H_Counters[BasicDigitCounter][0] = 16
	config.Kv.OtherKeynames.H_Counters[BasicWordCounter][0] = 4
	config.Kv.OtherKeynames.H_Counters[BasicNumberCounter][0] = 4

	config.Kv.OtherVals = new(SimpleValConfig)
	for i := 0; i < 8; i++ {
		config.Kv.OtherVals.L_Counters[i] = []uint8{0}
		config.Kv.OtherVals.H_Counters[i] = []uint8{0}
	}

	config.Kv.OtherVals.H_Counters[BasicTotalCounter][0] = 32
	config.Kv.OtherVals.H_Counters[BasicLetterCounter][0] = 32
	config.Kv.OtherVals.H_Counters[BasicDigitCounter][0] = 32
	config.Kv.OtherVals.H_Counters[BasicWordCounter][0] = 16
	config.Kv.OtherVals.H_Counters[BasicNumberCounter][0] = 16
}

func (h *HeadersProfile) Profile(m map[string][]string) {
	h.Kv = new(KeyValProfile)
	h.Kv.Profile(m)
}

func (h *HeadersProfile) Decide(config *HeadersConfig) string {
	str := h.Kv.Decide(&config.Kv)
	if str == "" {
		return str
	}
	return fmt.Sprintf("HttpHeaders: %s", str)
}

// Allow typical values - use for development but not in production
func (config *HeadersConfig) AddTypicalVal() {
	config.Kv.OtherKeynames = new(SimpleValConfig)
	for i := 0; i < 8; i++ {
		config.Kv.OtherKeynames.L_Counters[i] = []uint8{0}
		config.Kv.OtherKeynames.H_Counters[i] = []uint8{0}
	}

	config.Kv.OtherKeynames.H_Counters[BasicTotalCounter][0] = 16
	config.Kv.OtherKeynames.H_Counters[BasicLetterCounter][0] = 16
	config.Kv.OtherKeynames.H_Counters[BasicDigitCounter][0] = 16
	config.Kv.OtherKeynames.H_Counters[BasicWordCounter][0] = 4
	config.Kv.OtherKeynames.H_Counters[BasicNumberCounter][0] = 4
	config.Kv.OtherKeynames.Flags = 1 << MinusSlot
	config.Kv.OtherKeynames.H_Counters[BasicSpecialCounter][0] = 2

	config.Kv.OtherVals = new(SimpleValConfig)
	for i := 0; i < 8; i++ {
		config.Kv.OtherVals.L_Counters[i] = []uint8{0}
		config.Kv.OtherVals.H_Counters[i] = []uint8{0}
	}

	config.Kv.OtherVals.H_Counters[BasicTotalCounter][0] = 32
	config.Kv.OtherVals.H_Counters[BasicLetterCounter][0] = 32
	config.Kv.OtherVals.H_Counters[BasicDigitCounter][0] = 32
	config.Kv.OtherVals.H_Counters[BasicWordCounter][0] = 16
	config.Kv.OtherVals.H_Counters[BasicNumberCounter][0] = 16
	config.Kv.OtherVals.H_Counters[BasicSpecialCounter][0] = 8

	config.Kv.OtherVals.Flags = 1<<MinusSlot | 1<<MultSlot | 1<<DivSlot | 1<<SlashAsteriskCommentSlot | 1<<DotSlot
}

func (rp *ReqProfile) Profile(req *http.Request) {
	rp.Url = new(UrlProfile)
	rp.Url.Profile(req.URL.Path)
	rp.Query = new(QueryProfile)
	rp.Query.Profile(req.URL.Query())
	rp.Headers = new(HeadersProfile)
	rp.Headers.Profile(req.Header)
}

func (rp *ReqProfile) Decide(config *ReqConfig) string {
	var ret string
	ret = rp.Url.Decide(&config.Url)
	if ret == "" {
		ret = rp.Query.Decide(&config.Qs)
		if ret == "" {
			ret = rp.Headers.Decide(&config.Headers)
			if ret == "" {
				return ret
			}
		}
	}
	return fmt.Sprintf("HttpRequest: %s", ret)
}
