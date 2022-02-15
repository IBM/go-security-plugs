package spec

import (
	"bytes"
	"fmt"
	"strings"
)

type SimpleValConfig struct { // 16 bytes
	Flags      uint64
	L_Counters [8][]uint8
	H_Counters [8][]uint8
}

type SimpleValProfile struct { // 16 bytes
	Flags         uint64
	BasicCounters [8]uint8
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
	MultSlot //10
	PlusSlot
	CommentSlot
	MinusSlot
	DotSlot
	DivSlot
	ColonSlot
	SemiSlot
	LtSlot
	EqualSlot
	GtSlot //20
	QuestionSlot
	CommaSlot
	LeftSquareBrecketSlot
	RdivideSlot
	RightSquareBrecketSlot
	PowerSlot
	UnderscoreSlot
	AccentSlot
	LeftCurlyBrecketSlot
	PipeSlot //30
	RightCurlyBrecketSlot
	HomeSlot //32
)
const ( // Slots for any code
	nonReadableCharCounterSlot = iota + 33
	//DigitSlot
	//LetterSlot
	UpperAsciiSlot
	UnicodeCharSlot
	SlashAsteriskCommentSlot
	SqlCommentSlot
	HexSlot
	LASTSLOT__
)

const (
	BasicTotalCounter = iota
	BasicLetterCounter
	BasicDigitCounter
	BasicSpecialCounter
	BasicSpaceCounter
	BasicWordCounter
	BasicNumberCounter
	BasicPathCounter
)

var CounterName = map[int]string{
	BasicTotalCounter:   "TotalCounter",
	BasicLetterCounter:  "LetterCounter",
	BasicDigitCounter:   "DigitCounter",
	BasicSpecialCounter: "SpecialCounter",
	BasicSpaceCounter:   "SpaceCounter",
	BasicWordCounter:    "WordCounter",
	BasicNumberCounter:  "NumberCounter",
	BasicPathCounter:    "PathCounter",
}

var FlagName = map[int]string{
	SpaceSlot:                  "Space",
	ExclamationSlot:            "Exclamation",
	DoubleQouteSlot:            "DoubleQoute",
	NumberSlot:                 "NumberSign",
	DollarSlot:                 "DollarSign",
	PrecentSlot:                "PrecentSign",
	SingleQouteSlot:            "SingleQoute",
	LeftRoundBrecketSlot:       "LeftRoundBrecket",
	RightRoundBrecketSlot:      "RightRoundBrecket",
	MultSlot:                   "MultiplySign",
	PlusSlot:                   "PlusSign",
	CommentSlot:                "CommentSign",
	MinusSlot:                  "MinusSign",
	DotSlot:                    "DotSign",
	DivSlot:                    "DivideSign",
	ColonSlot:                  "ColonSign",
	SemiSlot:                   "SemicolonSign",
	LtSlot:                     "LessThenSign",
	EqualSlot:                  "EqualSign",
	GtSlot:                     "GreaterThenSign",
	QuestionSlot:               "QuestionMark",
	CommaSlot:                  "CommaSign",
	LeftSquareBrecketSlot:      "LeftSquareBrecket",
	RdivideSlot:                "ReverseDivideSign",
	RightSquareBrecketSlot:     "RightSquareBrecket",
	PowerSlot:                  "PowerSign",
	UnderscoreSlot:             "UnderscoreSign",
	AccentSlot:                 "AccentSign",
	LeftCurlyBrecketSlot:       "LeftCurlyBrecket",
	PipeSlot:                   "PipeSign",
	RightCurlyBrecketSlot:      "RightCurlyBrecket",
	nonReadableCharCounterSlot: "NonReadableChar",
	//DigitSlot:                  "Digit",
	//LetterSlot:                 "Letter",
	UpperAsciiSlot:           "UpperAsciiChar",
	UnicodeCharSlot:          "UnicodeChar",
	SlashAsteriskCommentSlot: "CommentCombination",
	SqlCommentSlot:           "SqlComment",
	HexSlot:                  "HexCombination",
}

func nameFlags(f uint64) string {
	var ret bytes.Buffer
	for i := 0; i < LASTSLOT__; i++ {
		if (f & (0x1 << i)) != 0 {
			ret.WriteString(FlagName[i])
			ret.WriteString(" ")
		}
	}
	return ret.String()
}

func (svp *SimpleValProfile) NameFlags() string {
	return nameFlags(svp.Flags)
}

func (svp *SimpleValProfile) Decide(config *SimpleValConfig) string {
	if (svp.Flags & ^config.Flags) != 0 {
		return fmt.Sprintf("Unexpected Flags %s (%x) in Value", nameFlags(svp.Flags & ^config.Flags), svp.Flags & ^config.Flags)
	}
	for i := 0; i < 8; i++ {
		L := config.L_Counters[i]
		H := config.H_Counters[i]

		limits := len(L)
		if limits > len(H) { // Happends only in ilegal config
			limits = len(H)
		}

		counter := svp.BasicCounters[i]
		if limits == 0 && counter == 0 { // no config and counter is zero
			continue
		}

		success := false
		for j := 0; j < limits; j++ {
			if counter < L[j] {
				return fmt.Sprintf("Counter %s Out of Range: %d", CounterName[i], counter)
			}
			if counter <= H[j] { // found ok interval
				success = true
				break
			}
		}
		if !success {
			return fmt.Sprintf("Counter %s Out of Range: %d", CounterName[i], counter)
		}
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
	digitCounter := uint(0)
	letterCounter := uint(0)
	specialCharCounter := uint(0)
	spaceCounter := uint(0)
	wordCounter := uint(0)
	numberCounter := uint(0)
	totalCounter := uint(0)
	var zero bool
	digits := 0
	letters := 0
	for _, c := range str {
		letter := false
		digit := false
		totalCounter++
		if c < 97 {
			if c < 65 {
				if c < 32 { //0-31
					flags |= 1 << nonReadableCharCounterSlot
				} else if c < 33 { //32 space
					flags |= 0x1
					spaceCounter++
				} else if c < 48 { //33-47
					slot := uint(c - 32)
					flags |= 0x1 << slot
					specialCharCounter++
					fmt.Printf(">> c %d slot %d flags %d - %s\n", c, slot, flags, nameFlags(flags))
				} else if c < 58 { //48-57  012..9
					digitCounter++
					//flags |= 0x1 << DigitSlot
					digit = true
					digits++
				} else { //58-64
					slot := uint(c - 58 + 16)
					flags |= 0x1 << slot
					specialCharCounter++
				}
			} else if c < 91 { //65-90    ABC..Z
				if zero && c == 88 {
					flags |= 0x1 << HexSlot
				}
				letterCounter++
				//flags |= 0x1 << LetterSlot
				letter = true
				letters++
			} else { //91-96
				slot := uint(c - 91 + 23)
				flags |= 0x1 << slot
				specialCharCounter++
			}
		} else if c < 123 { //97-122   abc..z
			if zero && c == 120 {
				flags |= 0x1 << HexSlot
			}
			letterCounter++
			//flags |= 1 << LetterSlot
			letter = true
			letters++
		} else if c < 127 { //123-126
			slot := uint(c - 123 + 29)
			flags |= 0x1 << slot
			specialCharCounter++
		} else if c < 128 { //127
			flags |= 0x1 << nonReadableCharCounterSlot
		} else if c < 256 { //128-255
			flags |= 0x1 << UpperAsciiSlot
		} else { //unicode
			flags |= 0x1 << UnicodeCharSlot
		}
		zero = (c == 48)

		if flags&(0x1<<DivSlot) > 0 && flags&(0x1<<MultSlot) > 0 {
			if strings.Count(str, "/*") > 0 || strings.Count(str, "*/") > 0 {
				flags |= 1 << SlashAsteriskCommentSlot
			}
		}
		if flags&(0x1<<MinusSlot) > 0 {
			if strings.Count(str, "--") > 0 {
				flags |= 1 << SqlCommentSlot
			}
		}
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
		letters = 0
	}
	if digits > 0 {
		numberCounter++
		digits = 0
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
		if spaceCounter > 0xFF {
			spaceCounter = 0xFF
		}
	}

	svp.BasicCounters[BasicTotalCounter] = uint8(totalCounter)
	svp.BasicCounters[BasicDigitCounter] = uint8(digitCounter)
	svp.BasicCounters[BasicLetterCounter] = uint8(letterCounter)
	svp.BasicCounters[BasicSpecialCounter] = uint8(specialCharCounter)
	svp.BasicCounters[BasicSpaceCounter] = uint8(spaceCounter)
	svp.BasicCounters[BasicWordCounter] = uint8(wordCounter)
	svp.BasicCounters[BasicNumberCounter] = uint8(numberCounter)
	//dataProfile.BasicCounters[BasicPathCounter] = uint8(partsCounter)
	svp.Flags = flags
	//fmt.Printf("Simple Data Profile: %v\n", svp)
}

// Allow generic value based on example (whitelisting)
// Call multiple times top present multiple examples
func (config *SimpleValConfig) AddValExample(str string) {
	svp := new(SimpleValProfile)
	svp.Profile(str)
	config.Flags |= svp.Flags
	for i := 0; i < 8; i++ {
		R := svp.BasicCounters[i]
		L := config.L_Counters[i]
		H := config.H_Counters[i]
		if len(L) == 0 {
			config.L_Counters[i] = append(L, R)
			config.H_Counters[i] = append(H, R)
		} else {
			if L[0] > R {
				L[0] = R
			}
			if H[0] < R {
				H[0] = R
			}
		}
	}

}

func (config *SimpleValConfig) NameFlags() string {
	return nameFlags(config.Flags)
}

func (config *SimpleValConfig) Describe() string {
	var description bytes.Buffer
	description.WriteString("Flags: ")
	description.WriteString(config.NameFlags())

	for i := 0; i < 8; i++ {
		description.WriteString(" - ")
		description.WriteString(CounterName[i])
		description.WriteString(": ")
		L := config.L_Counters[i]
		H := config.H_Counters[i]
		if len(L) == 0 {
			description.WriteString("No Limit")
		} else {
			description.WriteString(fmt.Sprintf("%d-%d", L[0], H[0]))
			for j := 1; j < len(L); j++ {
				description.WriteString(fmt.Sprintf(", %d-%d", L[j], H[j]))
			}
		}
	}
	return description.String()
}
