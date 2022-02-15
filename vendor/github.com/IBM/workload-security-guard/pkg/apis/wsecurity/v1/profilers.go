package v1

import (
	"bytes"
	"fmt"
)

type U8Minmax struct {
	Min uint8 `json:"min"`
	Max uint8 `json:"max"`
}

type U8MinmaxSlice []U8Minmax
type Uint64Slice []uint64

func (base Uint64Slice) Add(val Uint64Slice) Uint64Slice {
	if missing := len(val) - len(base); missing > 0 {
		// Dynamically allocate as many blockElements as needed for this config
		base = append(base, make([]uint64, missing)...)
	}
	for i, v := range val {
		base[i] = base[i] | v
	}
	return base
}

func (base Uint64Slice) Decide(val Uint64Slice) string {
	for i, v := range val {
		if v == 0 {
			continue
		}
		if i < len(base) && (v & ^base[i]) == 0 {
			continue
		}
		return fmt.Sprintf("Unexpected Unicode Flag %x on Element %d", v, i)
	}
	return ""
}

func (uint64Slice Uint64Slice) Describe() string {
	if uint64Slice != nil {
		var description bytes.Buffer
		description.WriteString("Unicode slice: ")
		description.WriteString(fmt.Sprintf("%x", uint64Slice[0]))
		for i := 1; i < len(uint64Slice); i++ {
			description.WriteString(fmt.Sprintf(", %x", uint64Slice[i]))
		}
		return description.String()
	}
	return ""
}

func (uint64Slice Uint64Slice) Marshal() string {
	if uint64Slice == nil {
		return "null"
	}
	var description bytes.Buffer
	description.WriteString(fmt.Sprintf("[%x", uint64Slice[0]))
	for i := 1; i < len(uint64Slice); i++ {
		description.WriteString(fmt.Sprintf(",%x", uint64Slice[i]))
	}
	description.WriteString("]")
	return description.String()

}

func (mms U8MinmaxSlice) Decide(v uint8) string {
	if len(mms) > 0 && v > 0 {
		for j := 0; j < len(mms); j++ {
			if v < mms[j].Min {
				break
			}
			if v <= mms[j].Max { // found ok interval
				return ""
			}
		}
		return fmt.Sprintf("Counter out of Range: %d", v)
	}
	return ""
}

func (mms U8MinmaxSlice) AddValExample(v uint8) U8MinmaxSlice {
	if len(mms) == 0 {
		mms = append(mms, U8Minmax{v, v})
	} else {
		if mms[0].Min > v {
			mms[0].Min = v
		}
		if mms[0].Max < v {
			mms[0].Max = v
		}
	}
	return mms
}

func (mms U8MinmaxSlice) Describe() string {
	if len(mms) == 0 {
		return "No Limit"
	}
	var description bytes.Buffer
	description.WriteString(fmt.Sprintf("%d-%d", mms[0].Min, mms[0].Max))
	for j := 1; j < len(mms); j++ {
		description.WriteString(fmt.Sprintf(", %d-%d", mms[j].Min, mms[j].Max))
	}
	return description.String()
}

func (mms U8MinmaxSlice) Marshal() string {
	if len(mms) == 0 {
		return "null"
	}
	var description bytes.Buffer
	description.WriteString(fmt.Sprintf("[{Min:%d,Max: %d", mms[0].Min, mms[0].Max))
	for j := 1; j < len(mms); j++ {
		description.WriteString(fmt.Sprintf("}, {Min:%d,Max: %d", mms[j].Min, mms[j].Max))
	}
	description.WriteString("}]")
	return description.String()
}
