package v1

import (
	"bytes"
	"fmt"
	"strings"
)

// Decission process
// If request profile allowed by ReqConfig: - Main Critiria
//        <Allow> + Log and gather statistics
// If Consult.Active and did not cross Consult.RequestsPerMinuete
//         If request profile allowed by Guard:  - Secondary Critiria
//                <Allow> + Log and gather statistics
// Log and gather statistics about request not allowed
// If ForceAllow
//          <Allow>		// used for example when ReqConfig is not ready
// <Block>
type Consult struct { // If guard needs to be consulted but is unavaliable => block
	Active             bool   `json:"active"` // False means never consult guard
	RequestsPerMinuete uint16 `json:"rpm"`    // Maximum rpm allows for consulting guard
}

type WsGate struct {
	Req          ReqConfig `json:"req"`        // Main critiria for blocking/allowing
	ConsultGuard Consult   `json:"consult"`    // If blocked by main critiria, consult guard (if avaliable)
	ForceAllow   bool      `json:"forceAllow"` // Allow no matter what! Overides all blocking.
}

func (g *WsGate) Marshal(depth int) string {
	var description bytes.Buffer
	shift := strings.Repeat("  ", depth)
	description.WriteString("{\n")
	description.WriteString(shift)
	description.WriteString(fmt.Sprintf("  Req: %s", g.Req.Marshal(depth+1)))
	description.WriteString(shift)
	description.WriteString(fmt.Sprintf("  ConsultGuard: %v", g.ConsultGuard))
	description.WriteString(shift)
	description.WriteString(fmt.Sprintf("  ForceAllow: %t", g.ForceAllow))
	description.WriteString(shift)
	description.WriteString("}\n")
	return description.String()
}
