package v1

import (
	"bytes"
	"fmt"
	"strings"
)

type SimpleValConfig struct { // 16 bytes
	Flags uint64 `json:"flags"`
	//Counters     [8]U8MinmaxSlice
	Runes        U8MinmaxSlice `json:"runes"`
	Digits       U8MinmaxSlice `json:"digits"`
	Letters      U8MinmaxSlice `json:"letters"`
	SpecialChars U8MinmaxSlice `json:"schars"`
	Words        U8MinmaxSlice `json:"words"`
	Numbers      U8MinmaxSlice `json:"numbers"`
	UnicodeFlags Uint64Slice   `json:"unicodeFlags"` //[]uint64
}

type SimpleValProfile struct { // 16 bytes
	Flags uint64
	//BasicCounters [8]uint8
	Runes        uint8
	Digits       uint8
	Letters      uint8
	SpecialChars uint8
	Words        uint8
	Numbers      uint8
	UnicodeFlags Uint64Slice //[]uint64
}

// Slots and counters for AsciiDaya:
// 0-31 (32) nonReadableRCharCounter
// 32-47 (16) slots 0-15 respectivly
// 48-57 (10) digitCounter
// 58-64 (6) slots 16-22
// 65-90 (26) smallLetterCounter
// 91-96 (6) slots 23-28
// 97-122 (26) capitalLetterCounter
// 123-126 (4) slots 29-32
// 127 (1) nonReadableRCharCounter
// Slots:
// <SPACE> ! " # $ % & ' ( ) * + , - . / : ; < = > ? @ [ \ ] ^ _ ` { | } ~
//    0    1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2
const ( // Slots for Ascii 0-127
	SpaceSlot = iota
	ExclamationSlot
	DoubleQouteSlot
	NumberSlot
	DollarSlot
	PrecentSlot
	andSlot
	SingleQouteSlot
	LeftRoundBrecketSlot
	RightRoundBrecketSlot
	MultSlot // 10
	PlusSlot
	CommentSlot
	MinusSlot
	DotSlot
	DivSlot
	ColonSlot
	SemiSlot
	LtSlot
	EqualSlot
	GtSlot // 20
	QuestionSlot
	CommaSlot
	LeftSquareBrecketSlot
	RdivideSlot
	RightSquareBrecketSlot
	PowerSlot
	UnderscoreSlot
	AccentSlot
	LeftCurlyBrecketSlot
	PipeSlot // 30
	RightCurlyBrecketSlot
	HomeSlot            // 32
	nonReadableCharSlot // 33
	UnicodeCharSlot     // 34
)
const ( // Slots for any code
	SlashAsteriskCommentSlot = iota + 35
	SqlCommentSlot
	HexSlot
	LASTSLOT__
)

const (
	TotalCounter = iota
	LetterCounter
	DigitCounter
	SpecialCharCounter
	WordCounter
	NumberCounter
	SpareCounter1__
	SpareCounter2__
)

var CounterName = map[int]string{
	TotalCounter:       "TotalCounter",
	LetterCounter:      "LetterCounter",
	DigitCounter:       "DigitCounter",
	SpecialCharCounter: "SpecialCharCounter",
	WordCounter:        "WordCounter",
	NumberCounter:      "NumberCounter",
	SpareCounter1__:    "<UnusedCounter>",
	SpareCounter2__:    "<UnusedCounter>",
}

var FlagName = map[int]string{
	SpaceSlot:                "Space",
	ExclamationSlot:          "Exclamation",
	DoubleQouteSlot:          "DoubleQoute",
	NumberSlot:               "NumberSign",
	DollarSlot:               "DollarSign",
	PrecentSlot:              "PrecentSign",
	SingleQouteSlot:          "SingleQoute",
	LeftRoundBrecketSlot:     "LeftRoundBrecket",
	RightRoundBrecketSlot:    "RightRoundBrecket",
	MultSlot:                 "MultiplySign",
	PlusSlot:                 "PlusSign",
	CommentSlot:              "CommentSign",
	MinusSlot:                "MinusSign",
	DotSlot:                  "DotSign",
	DivSlot:                  "DivideSign",
	ColonSlot:                "ColonSign",
	SemiSlot:                 "SemicolonSign",
	LtSlot:                   "LessThenSign",
	EqualSlot:                "EqualSign",
	GtSlot:                   "GreaterThenSign",
	QuestionSlot:             "QuestionMark",
	CommaSlot:                "CommaSign",
	LeftSquareBrecketSlot:    "LeftSquareBrecket",
	RdivideSlot:              "ReverseDivideSign",
	RightSquareBrecketSlot:   "RightSquareBrecket",
	PowerSlot:                "PowerSign",
	UnderscoreSlot:           "UnderscoreSign",
	AccentSlot:               "AccentSign",
	LeftCurlyBrecketSlot:     "LeftCurlyBrecket",
	PipeSlot:                 "PipeSign",
	RightCurlyBrecketSlot:    "RightCurlyBrecket",
	nonReadableCharSlot:      "NonReadableChar",
	UnicodeCharSlot:          "UnicodeChar",
	SlashAsteriskCommentSlot: "CommentCombination",
	SqlCommentSlot:           "SqlComment",
	HexSlot:                  "HexCombination",
}

func SetFlags(slots []int) (f uint64) {
	for _, slot := range slots {
		f = f | (0x1 << slot)
	}
	return
}
func NameFlags(f uint64) string {
	var ret bytes.Buffer
	mask := uint64(0x1)
	for i := 0; i < LASTSLOT__; i++ {
		if (f & mask) != 0 {
			ret.WriteString(FlagName[i])
			ret.WriteString(" ")
			f = f ^ mask
		}
		mask = mask << 1
	}
	if f != 0 {
		ret.WriteString("<UnnamedFlags>")
	}
	return ret.String()
}

func NewSimpleValConfig(runes, letters, digits, specialChars, words, numbers uint8) *SimpleValConfig {
	svc := new(SimpleValConfig)
	svc.Runes = make([]U8Minmax, 1)
	svc.Letters = make([]U8Minmax, 1)
	svc.Digits = make([]U8Minmax, 1)
	svc.SpecialChars = make([]U8Minmax, 1)
	svc.Words = make([]U8Minmax, 1)
	svc.Numbers = make([]U8Minmax, 1)

	svc.Runes[0].Max = runes
	svc.Letters[0].Max = letters
	svc.Digits[0].Max = digits
	svc.SpecialChars[0].Max = specialChars
	svc.Words[0].Max = words
	svc.Numbers[0].Max = numbers
	return svc
}

func (svp *SimpleValProfile) NameFlags() string {
	return NameFlags(svp.Flags)
}

//func (svp *SimpleValProfile) Decide(config *SimpleValConfig) string {
func (config *SimpleValConfig) Decide(svp *SimpleValProfile) string {

	if (svp.Flags & ^config.Flags) != 0 {
		return fmt.Sprintf("Unexpected Flags %s (%x) in Value", NameFlags(svp.Flags & ^config.Flags), svp.Flags & ^config.Flags)
	}
	if ret := config.UnicodeFlags.Decide(svp.UnicodeFlags); ret != "" {
		return ret
	}
	if ret := config.Runes.Decide(svp.Runes); ret != "" {
		return fmt.Sprintf("Runes: %s", ret)
	}
	if ret := config.Digits.Decide(svp.Digits); ret != "" {
		return fmt.Sprintf("Digits: %s", ret)
	}
	if ret := config.Letters.Decide(svp.Letters); ret != "" {
		return fmt.Sprintf("Letters: %s", ret)
	}
	if ret := config.SpecialChars.Decide(svp.SpecialChars); ret != "" {
		return fmt.Sprintf("SpecialChars: %s", ret)
	}
	if ret := config.Words.Decide(svp.Words); ret != "" {
		return fmt.Sprintf("Words: %s", ret)
	}
	if ret := config.Numbers.Decide(svp.Numbers); ret != "" {
		return fmt.Sprintf("Numbers: %s", ret)
	}
	return ""
}

// Profile generic value where we expect:
// some short combination of chars
// mainly english letters and/or digits (ascii)
// potentially some small content of special chars
// typically no unicode
func (svp *SimpleValProfile) Profile(str string) {
	var flags uint64
	unicodeFlags := []uint64{}
	digitCounter := uint(0)
	letterCounter := uint(0)
	specialCharCounter := uint(0)
	wordCounter := uint(0)
	numberCounter := uint(0)
	totalCounter := uint(0)
	var zero, asterisk, slash, minus bool
	digits := 0
	letters := 0
	for _, c := range str {
		letter := false
		digit := false
		totalCounter++
		if c < 'a' { //0-96
			if c < 'A' { // 0-64
				if c < '0' { //0-47
					if c > 32 { //33-47
						slot := uint(c - 32)
						flags |= 0x1 << slot
						specialCharCounter++
						if c == '/' {
							if asterisk {
								flags |= 1 << SlashAsteriskCommentSlot
							}
						}
						if slash && c == '*' {
							flags |= 1 << SlashAsteriskCommentSlot
						}
						if minus && c == '-' {
							flags |= 1 << SqlCommentSlot
						}

					} else if c < 32 { //0-31
						flags |= 1 << nonReadableCharSlot
					} else { //32 space
						flags |= 0x1
					}
				} else if c <= '9' { //48-57  012..9
					digitCounter++
					digit = true
					digits++
				} else { //58-64
					slot := uint(c - 58 + 16)
					flags |= 0x1 << slot
					specialCharCounter++
				}
			} else if c <= 'Z' { //65-90    ABC..Z
				if zero && c == 'X' {
					flags |= 0x1 << HexSlot
				}
				letterCounter++
				letter = true
				letters++
			} else { //91-96
				slot := uint(c - 91 + 23)
				flags |= 0x1 << slot
				specialCharCounter++
			}
		} else if c <= 'z' { //97-122   abc..z
			if zero && c == 'x' {
				flags |= 0x1 << HexSlot
			}
			letterCounter++
			letter = true
			letters++
		} else if c < 127 { //123-126
			slot := uint(c - 123 + 29)
			flags |= 0x1 << slot
			specialCharCounter++
		} else if c < 128 { //127
			flags |= 0x1 << nonReadableCharSlot
		} else {
			// Unicode -  128 and onwards
			flags |= 0x1 << UnicodeCharSlot
			// Next we use a rought but quick way to profile unicodes using blocks of 128 codes
			// Block 0 is 128-255, block 1 is 256-383...
			// BlockBit represent the bit in a blockElement. Each blockElement carry 64 bits
			block := (c / 0x80) - 1
			blockBit := int(block & 0x3F)
			blockElement := int(block / 0x40)
			if blockElement >= len(unicodeFlags) {
				// Dynamically allocate as many blockElements as needed for this profile
				unicodeFlags = append(unicodeFlags, make([]uint64, blockElement-len(unicodeFlags)+1)...)
			}
			unicodeFlags[blockElement] |= 0x1 << blockBit
		}
		zero = (c == '0')
		asterisk = (c == '*')
		slash = (c == '/')
		minus = (c == '-')

		if letters > 0 && !letter {
			wordCounter++
			letters = 0
		}
		if digits > 0 && !digit {
			numberCounter++
			digits = 0
		}
	}
	if letters > 0 {
		wordCounter++
	}
	if digits > 0 {
		numberCounter++
	}
	if totalCounter > 0xFF {
		totalCounter = 0xFF
		if digitCounter > 0xFF {
			digitCounter = 0xFF
		}
		if letterCounter > 0xFF {
			letterCounter = 0xFF
		}
		if specialCharCounter > 0xFF {
			specialCharCounter = 0xFF
		}
		if numberCounter > 0xFF {
			numberCounter = 0xFF
		}
		if wordCounter > 0xFF {
			wordCounter = 0xFF
		}
	}

	svp.Runes = uint8(totalCounter)
	svp.Digits = uint8(digitCounter)
	svp.Letters = uint8(letterCounter)
	svp.SpecialChars = uint8(specialCharCounter)
	svp.Words = uint8(wordCounter)
	svp.Numbers = uint8(numberCounter)

	svp.Flags = flags
	if len(unicodeFlags) > 0 {
		svp.UnicodeFlags = unicodeFlags
	}
	//fmt.Println(svp.Describe())

}

func (svp *SimpleValProfile) Describe() string {
	var description bytes.Buffer
	description.WriteString("Flags: ")
	description.WriteString(svp.NameFlags())
	description.WriteString(svp.UnicodeFlags.Describe())
	description.WriteString(fmt.Sprintf("Runes: %d", svp.Runes))
	description.WriteString(fmt.Sprintf("Letters: %d", svp.Letters))
	description.WriteString(fmt.Sprintf("Digits: %d", svp.Digits))
	description.WriteString(fmt.Sprintf("SpecialChars: %d", svp.SpecialChars))
	description.WriteString(fmt.Sprintf("Words: %d", svp.Words))
	description.WriteString(fmt.Sprintf("Numbers: %d", svp.Numbers))

	return description.String()
}

// Allow generic value based on example (whitelisting)
// Call multiple times top present multiple examples
func (config *SimpleValConfig) AddValExample(str string) {
	svp := new(SimpleValProfile)
	svp.Profile(str)
	config.Flags |= svp.Flags
	config.Runes = config.Runes.AddValExample(svp.Runes)
	config.Digits = config.Digits.AddValExample(svp.Digits)
	config.Letters = config.Letters.AddValExample(svp.Letters)
	config.SpecialChars = config.SpecialChars.AddValExample(svp.SpecialChars)
	config.Words = config.Words.AddValExample(svp.Words)
	config.Numbers = config.Numbers.AddValExample(svp.Numbers)
	config.UnicodeFlags = config.UnicodeFlags.Add(svp.UnicodeFlags)
}

func (config *SimpleValConfig) NameFlags() string {
	return NameFlags(config.Flags)
}

func (config *SimpleValConfig) Describe() string {
	var description bytes.Buffer
	description.WriteString("Flags: ")
	description.WriteString(config.NameFlags())

	description.WriteString(config.UnicodeFlags.Describe())

	description.WriteString("Runes: ")
	description.WriteString(config.Runes.Describe())
	description.WriteString("Letters: ")
	description.WriteString(config.Letters.Describe())
	description.WriteString("Digits: ")
	description.WriteString(config.Digits.Describe())
	description.WriteString("SpecialChars: ")
	description.WriteString(config.SpecialChars.Describe())
	description.WriteString("Words: ")
	description.WriteString(config.Words.Describe())
	description.WriteString("Numbers: ")
	description.WriteString(config.Numbers.Describe())
	return description.String()
}

func (config *SimpleValConfig) Marshal(depth int) string {
	var description bytes.Buffer
	shift := strings.Repeat("  ", depth)
	description.WriteString("{\n")
	description.WriteString(shift)
	description.WriteString(fmt.Sprintf("  Flags: 0x%x,\n", config.Flags))
	description.WriteString(shift)
	description.WriteString(fmt.Sprintf("  UnicodeFlags: %s,\n", config.UnicodeFlags.Marshal()))
	description.WriteString(shift)
	description.WriteString(fmt.Sprintf("  Runes: %s,\n", config.Runes.Marshal()))
	description.WriteString(shift)
	description.WriteString(fmt.Sprintf("  Letters: %s,\n", config.Letters.Marshal()))
	description.WriteString(shift)
	description.WriteString(fmt.Sprintf("  Digits: %s,\n", config.Digits.Marshal()))
	description.WriteString(shift)
	description.WriteString(fmt.Sprintf("  SpecialChars: %s,\n", config.SpecialChars.Marshal()))
	description.WriteString(shift)
	description.WriteString(fmt.Sprintf("  Words: %s,\n", config.Words.Marshal()))
	description.WriteString(shift)
	description.WriteString(fmt.Sprintf("  Numbers: %s,\n", config.Numbers.Marshal()))
	description.WriteString(shift)
	description.WriteString("}\n")
	return description.String()
}
