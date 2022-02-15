package v1

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
)

type ReqConfig struct {
	Url     UrlConfig     `json:"url"`
	Qs      QueryConfig   `json:"qs"`
	Headers HeadersConfig `json:"headers"`
}

type ReqProfile struct {
	Url     *UrlProfile
	Query   *QueryProfile
	Headers *HeadersProfile
}

type UrlProfile struct {
	Val      *SimpleValProfile
	Segments uint8
}
type UrlConfig struct {
	Val      SimpleValConfig `json:"val"`
	Segments U8MinmaxSlice   `json:"segments"`
}
type QueryProfile struct {
	Kv *KeyValProfile
}

type QueryConfig struct {
	Kv KeyValConfig `json:"kv"`
}

type HeadersProfile struct {
	Kv *KeyValProfile
}

type HeadersConfig struct {
	Kv KeyValConfig `json:"kv"`
}

func (u *UrlProfile) Profile(path string) {
	segments := strings.Split(path, "/")
	cleanPath := strings.Join(segments, "")
	u.Val = new(SimpleValProfile)
	u.Val.Profile(cleanPath)
	numSegments := len(segments)
	if numSegments > 0xFF {
		numSegments = 0xFF
	}
	u.Segments = uint8(numSegments)
}

func (config *UrlConfig) Decide(u *UrlProfile) string {
	if str := config.Segments.Decide(u.Segments); str != "" {
		return fmt.Sprintf("URL Segmengs: %s", str)
	}

	if str := config.Val.Decide(u.Val); str != "" {
		return fmt.Sprintf("URL: %s", str)
	}
	return ""
}

func (config *UrlConfig) Marshal(depth int) string {
	var description bytes.Buffer
	shift := strings.Repeat("  ", depth)
	description.WriteString("{\n")
	description.WriteString(shift)
	description.WriteString(fmt.Sprintf("  Val: %s", config.Val.Marshal(depth+1)))
	description.WriteString(shift)
	description.WriteString(fmt.Sprintf("  Segments: %s", config.Segments.Marshal()))
	description.WriteString(shift)
	description.WriteString("}\n")
	return description.String()
}

// Allow typical URL values - use for development but not in production
func (config *UrlConfig) AddTypicalVal() {

	config.Val.Runes = make([]U8Minmax, 1)
	config.Val.Letters = make([]U8Minmax, 1)
	config.Val.Digits = make([]U8Minmax, 1)
	config.Val.Words = make([]U8Minmax, 1)
	config.Val.Numbers = make([]U8Minmax, 1)

	config.Val.Runes[0].Max = 64
	config.Val.Letters[0].Max = 64
	config.Val.Digits[0].Max = 64
	config.Val.Words[0].Max = 16
	config.Val.Numbers[0].Max = 16
	config.Val.Flags = 1 << DivSlot
	config.Segments = make([]U8Minmax, 1)
	config.Segments[0].Max = 8
}

func (q *QueryProfile) Profile(m map[string][]string) {
	q.Kv = new(KeyValProfile)
	q.Kv.Profile(m)
}

func (config *QueryConfig) Decide(q *QueryProfile) string {
	str := config.Kv.Decide(q.Kv)
	if str == "" {
		return str
	}
	return fmt.Sprintf("QueryString: %s", str)
}

// Allow typical query string values - use for development but not in production
func (config *QueryConfig) AddTypicalVal() {
	config.Kv.OtherKeynames = NewSimpleValConfig(16, 16, 16, 0, 4, 4)
	config.Kv.OtherVals = NewSimpleValConfig(32, 32, 32, 0, 16, 16)
}

func (config *QueryConfig) Marshal(depth int) string {
	var description bytes.Buffer
	shift := strings.Repeat("  ", depth)
	description.WriteString("{\n")
	description.WriteString(shift)
	description.WriteString(fmt.Sprintf("  Kv: %s", config.Kv.Marshal(depth+1)))
	description.WriteString(shift)
	description.WriteString("}\n")
	return description.String()
}

func (h *HeadersProfile) Profile(m map[string][]string) {
	h.Kv = new(KeyValProfile)
	h.Kv.Profile(m)
}

func (config *HeadersConfig) Decide(h *HeadersProfile) string {
	str := config.Kv.Decide(h.Kv)
	if str == "" {
		return str
	}
	return fmt.Sprintf("HttpHeaders: %s", str)
}

func (config *HeadersConfig) Marshal(depth int) string {
	var description bytes.Buffer
	shift := strings.Repeat("  ", depth)
	description.WriteString("{\n")
	description.WriteString(shift)
	description.WriteString(fmt.Sprintf("  Kv: %s", config.Kv.Marshal(depth+1)))
	description.WriteString(shift)
	description.WriteString("}\n")
	return description.String()
}

// Allow typical values - use for development but not in production
func (config *HeadersConfig) AddTypicalVal() {
	config.Kv.OtherKeynames = NewSimpleValConfig(16, 16, 16, 2, 4, 4)
	config.Kv.OtherVals = NewSimpleValConfig(32, 32, 32, 8, 16, 16)
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

func (config *ReqConfig) Decide(rp *ReqProfile) string {
	var ret string
	ret = config.Url.Decide(rp.Url)
	if ret == "" {
		ret = config.Qs.Decide(rp.Query)
		if ret == "" {
			ret = config.Headers.Decide(rp.Headers)
			if ret == "" {
				return ret
			}
		}
	}
	return fmt.Sprintf("HttpRequest: %s", ret)
}

func (config *ReqConfig) Marshal(depth int) string {
	var description bytes.Buffer
	shift := strings.Repeat("  ", depth)
	description.WriteString("{\n")
	description.WriteString(shift)
	description.WriteString(fmt.Sprintf("  Url: %s", config.Url.Marshal(depth+1)))
	description.WriteString(shift)
	description.WriteString(fmt.Sprintf("  Qs: %s", config.Qs.Marshal(depth+1)))
	description.WriteString(shift)
	description.WriteString(fmt.Sprintf("  Headers: %s", config.Headers.Marshal(depth+1)))
	description.WriteString(shift)
	description.WriteString("}\n")
	return description.String()
}

// Allow typical values - use for development but not in production
func (config *ReqConfig) AddTypicalVal() {
	config.Headers.AddTypicalVal()
	config.Url.AddTypicalVal()
	config.Qs.AddTypicalVal()
}
