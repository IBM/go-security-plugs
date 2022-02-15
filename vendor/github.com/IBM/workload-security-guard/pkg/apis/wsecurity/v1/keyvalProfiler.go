package v1

import (
	"bytes"
	"fmt"
	"strings"
)

type KeyValConfig struct {
	Vals          map[string]*SimpleValConfig `json:"vals"`          // Profile the value of whitelisted keys
	MinimalSet    map[string]void             `json:"minimalSet"`    // Mandatory keys
	OtherVals     *SimpleValConfig            `json:"otherVals"`     // Profile the values of other keys
	OtherKeynames *SimpleValConfig            `json:"otherKeynames"` // Profile the keynames of other keys
}

type KeyValProfile struct {
	Vals map[string]*SimpleValProfile
}

type void struct{}

// Profile a generic map of key vals where we expect:
// keys belonging to some contstant list of keys
// vals have some defined charactaristics
func (kvp *KeyValProfile) Profile(m map[string][]string) {
	if len(m) == 0 { // no keys
		return
	}
	kvp.Vals = make(map[string]*SimpleValProfile, len(m))
	for k, v := range m {
		//var keyConfig *SimpleValConfig
		//if config.Vals != nil {
		//	keyConfig = config.Vals[k]
		//}
		//if keyConfig == nil {
		//}

		val := strings.Join(v, " ")
		kvp.Vals[k] = new(SimpleValProfile)
		kvp.Vals[k].Profile(val)
	}
}

func (config *KeyValConfig) Decide(kvp *KeyValProfile) string {
	//if config == nil || !config.Enable {
	//	return ""
	//}

	// Duplicate minimalSet map
	var required void
	minimalSet := make(map[string]void, len(config.MinimalSet))

	for k := range config.MinimalSet {
		minimalSet[k] = required
	}

	// For each key-val, decide! and remove from minimalSet
	if kvp.Vals != nil {

		for k, v := range kvp.Vals {
			delete(minimalSet, k) // Remove from minimalSet
			// Decide based on a known key
			if config.Vals != nil && config.Vals[k] != nil {
				if ret := config.Vals[k].Decide(v); ret != "" {
					return fmt.Sprintf("KeyVal known Key %s: %s", k, ret)
				}
				continue
			}
			// Not a known key...
			if config.OtherKeynames == nil || config.OtherVals == nil {
				return fmt.Sprintf("KeyVal key %s is not known", k)
			}
			// Decide keyname of not known key
			var keynames SimpleValProfile
			keynames.Profile(k)
			if ret := config.OtherKeynames.Decide(&keynames); ret != "" {
				return fmt.Sprintf("KeyVal other keyname %s: %s", k, ret)
			}
			// Decide val of not known key
			if ret := config.OtherVals.Decide(v); ret != "" {
				return fmt.Sprintf("KeyVal other keyname %s: %s", k, ret)
			}
			continue
		}
	}
	// Once we oked all keys, check if there are missing mandatory keys
	if len(minimalSet) > 0 {
		keys := make([]string, len(minimalSet))
		for k := range minimalSet {
			keys = append(keys, k)
		}
		return fmt.Sprintf("KeyVal missing mandatory keys %s", strings.Join(keys, ", "))
	}
	return ""
}

// Allow a list of specific keys and an example of their values
// Can be called multiple times to add keys or to add examples for values
// Use this when the keynames are known in advance
// Call multiple times to show the entire range of values per key
// For keys not known in advance, use WhitelistByExample() instead
func (config *KeyValConfig) WhitelistKnownKeys(m map[string]string) {
	if config.Vals == nil {
		config.Vals = make(map[string]*SimpleValConfig, len(m))
	}
	for k, v := range m {
		if config.Vals[k] == nil {
			config.Vals[k] = new(SimpleValConfig)
		}
		config.Vals[k].AddValExample(v)
	}
}

// Define which of the known keynames is mandatory (if any)
// Must call WhitelistKnownKeys before setting keys as Mandatory
func (config *KeyValConfig) SetMandatoryKeys(minimalSet []string) {
	if config.Vals == nil {
		panic("Keys should be set with WhitelistKnownKeys before becoming Mandatory")
	}

	if config.MinimalSet == nil {
		config.MinimalSet = make(map[string]void, len(minimalSet))
	}

	var required void
	for _, k := range minimalSet {
		if _, exists := config.Vals[k]; !exists {
			panic(fmt.Sprintf("Key \"%s\" should be set with WhitelistKnownKeys before becoming Mandatory", k))
		}
		config.MinimalSet[k] = required
	}
}

// Allow keynames and their values based on examples
// Can be called multiple times to add examples for keynames or values
// Use this when the keynames are not known in advance
// Call multiple times to show the entire range of keynames and values
// When keys are known in advance, use WhitelistKnownKeys() instead
func (config *KeyValConfig) WhitelistByExample(k string, v string) {
	if config.OtherKeynames == nil {
		config.OtherKeynames = new(SimpleValConfig)

	}
	config.OtherKeynames.AddValExample(k)

	if config.OtherVals == nil {
		config.OtherVals = new(SimpleValConfig)

	}
	config.OtherVals.AddValExample(v)
}

func (config *KeyValConfig) Describe() string {
	var description bytes.Buffer

	if config.Vals != nil {
		for k, v := range config.Vals {
			if _, exists := config.MinimalSet[k]; exists {
				description.WriteString(" | MandatoryKey: ")
			} else {
				description.WriteString(" | OptionalKey: ")
			}
			description.WriteString(k)
			description.WriteString(" => ")
			description.WriteString(v.Describe())
		}
	}

	if config.OtherVals != nil {
		description.WriteString(" | OtherVals: ")
		description.WriteString(config.OtherVals.Describe())
	}
	if config.OtherKeynames != nil {
		description.WriteString(" | OtherKeynames: ")
		description.WriteString(config.OtherKeynames.Describe())
	}

	return description.String()
}

func (config *KeyValConfig) Marshal(depth int) string {
	var description bytes.Buffer
	var started bool
	shift := strings.Repeat("  ", depth)
	description.WriteString("{\n")

	if len(config.Vals) > 0 {
		description.WriteString(shift)
		description.WriteString("  Vals: {\n")
		for k, v := range config.Vals {
			description.WriteString(shift)
			description.WriteString(fmt.Sprintf("  , %s: %s", k, v.Marshal(depth+1)))
		}
		description.WriteString(shift)
		description.WriteString("  }\n")
		started = true
	}

	if config.MinimalSet != nil {
		if started {
			description.WriteString(", ")
		} else {
			description.WriteString("  ")
		}
		description.WriteString(shift)
		description.WriteString(", MinimalSet: {\n")
		for k, v := range config.MinimalSet {
			description.WriteString(shift)
			description.WriteString(fmt.Sprintf("  , %s: %s\n", k, v))
		}
		description.WriteString(shift)
		description.WriteString("  }\n")
		started = true
	}
	if config.OtherKeynames != nil {
		if started {
			description.WriteString(", ")
		} else {
			description.WriteString("  ")
		}
		description.WriteString(shift)
		description.WriteString(fmt.Sprintf(", OtherKeynames: %s/n", config.OtherKeynames.Marshal(depth+1)))
		started = true
	}
	if config.OtherVals != nil {
		if started {
			description.WriteString(", ")
		} else {
			description.WriteString("  ")
		}
		description.WriteString(shift)
		description.WriteString(fmt.Sprintf(", OtherVals: %s/n", config.OtherVals.Marshal(depth+1)))
	}
	description.WriteString(shift)
	description.WriteString("}\n")
	return description.String()
}
