package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
)

const preserveNullRow = false

type Line struct {
	rating         string
	gPlusPlaceId   string
	unixReviewTime string
	gPlusUserId    string
}

func main() {
	if err := scan(); err != nil {
		log.Fatalln(err.Error())
	}
}

func scan() error {
	scanner := bufio.NewScanner(os.Stdin)
	lineNum := 1
	for scanner.Scan() {
		// TODO(nekketsuuu): read a line if it's too long
		// TODO(nekketsuuu): read by bytes?
		l, err := unmarshal(scanner.Text())
		if err != nil {
			return errors.New("An error occurred while scannning line " + strconv.Itoa(lineNum) + ": " + err.Error())
		}

		if preserveNullRow ||
			(l.unixReviewTime != "null" && l.rating != "null" && l.gPlusPlaceId != "null" && l.gPlusUserId != "null") {
			// tenuki (be careful of type of each value)
			fmt.Printf("{\"rating\": %s, \"gPlusPlaceID\": \"%s\", \"unixReviewTime\": %s, \"gPlusUserID\": \"%s\"}\n",
				l.rating, l.gPlusPlaceId, l.unixReviewTime, l.gPlusUserId)
		}
		lineNum++
	}
	// discard errors
	return nil
}

type decodeState struct {
	data    string
	offset  int
	dataLen int
}

func initState(data string) decodeState {
	state := decodeState{}
	state.data = data
	state.offset = 0
	state.dataLen = len(data)
	return state
}
func (state *decodeState) isEOL() bool {
	return state.offset >= state.dataLen
}
func (state *decodeState) incOffset() error {
	state.offset++
	if state.isEOL() {
		return errors.New("decodeState.incOffset: line overflow")
	}
	return nil
}
func (state *decodeState) skipSpaces() error {
	for state.data[state.offset] == byte(' ') {
		if err := state.incOffset(); err != nil {
			return err
		}
	}
	return nil
}
func (state *decodeState) newError(message string) error {
	return errors.New("Decode error at offset " + strconv.Itoa(state.offset) + ": " + message)
}

func unmarshal(line string) (Line, error) {
	state := initState(line)
	l := Line{}

	if state.data[state.offset] != byte('{') {
		return l, errors.New("A line does not begin with \"")
	}
	if err := state.incOffset(); err != nil {
		return l, errors.New("A line does not end with }")
	}
	if err := state.skipSpaces(); err != nil {
		return l, errors.New("A line does not end with }")
	}

	for state.data[state.offset] != byte('}') {
		key, err := parseKey(&state)
		if err != nil {
			return l, state.newError(err.Error())
		}

		value, err := parseValue(&state)
		if err != nil {
			return l, state.newError(err.Error())
		}

		if key == "rating" {
			l.rating = value
		} else if key == "gPlusPlaceId" {
			l.gPlusPlaceId = value
		} else if key == "unixReviewTime" {
			l.unixReviewTime = value
		} else if key == "gPlusUserId" {
			l.gPlusUserId = value
		}
	}

	return l, nil
}

func parseKey(state *decodeState) (string, error) {
	// tenuki
	key, err := decodeString(state)
	if err != nil {
		return "", errors.New("An error occurred while parsing a key: " + err.Error())
	}

	if state.data[state.offset] != byte(':') {
		return "", errors.New("Key syntax error: there is no colon after key name")
	}
	if err := state.incOffset(); err != nil {
		return "", errors.New("An value is expected after a key \"" + key + "\"")
	}
	if err := state.skipSpaces(); err != nil {
		return "", errors.New("An value is expected after a key \"" + key + "\"")
	}
	return key, nil
}

func parseValue(state *decodeState) (string, error) {
	var value string
	var err error

	// read before comma
	if s := state.data[state.offset]; s == byte('u') || s == byte('"') || s == byte('\'') {
		value, err = decodeString(state)
		if err != nil {
			return "", errors.New("An error occurred while parsing a value: " + err.Error())
		}
	} else if state.data[state.offset] == byte('N') {
		if state.offset+3 >= state.dataLen || state.data[state.offset:state.offset+4] != "None" {
			return "", errors.New("An error occurred while parsing a value: Syntax error: expected \"None\"")
		}
		state.offset += 3
		if err := state.incOffset(); err != nil {
			return "", errors.New("Syntax error: expected a letter after \"None\"")
		}
		value = "null"
	} else if s := state.data[state.offset]; s == byte('+') || s == byte('-') || s == byte('0') || s == byte('1') || s == byte('2') || s == byte('3') || s == byte('4') || s == byte('5') || s == byte('6') || s == byte('7') || s == byte('8') || s == byte('9') {
		value, err = decodeNumber(state)
		if err != nil {
			return "", errors.New("An error occurred while parsing a value: " + err.Error())
		}
	} else if state.data[state.offset] == byte('[') {
		// tenuki
		beginIndex := state.offset
		if err := state.incOffset(); err != nil {
			return "", errors.New("Syntax error: there is not a closed bracket of an array")
		}
		if state.data[state.offset] == byte(']') {
			value = ""
		} else {
			// the array is not zero-length
			for {
				// parse an element
				// super tenuki (assume all elements are string)
				if _, err := decodeString(state); err != nil {
					return "", errors.New("Syntax error while parsing an array value: " + err.Error())
				}
				// exit if the next character is closed square-bracket
				if state.data[state.offset] == byte(']') {
					value = state.data[beginIndex:state.offset]
					if err := state.incOffset(); err != nil {
						return "", errors.New("Syntax error: expected some character after \"]\"")
					}
					break
				}
				// skip a comma
				if state.data[state.offset] != byte(',') {
					return "", errors.New("Syntax error: expected a comma after a element of an array")
				}
				if err := state.incOffset(); err != nil {
					return "", errors.New("Syntax error: there must be a next element of an array")
				}
				if err := state.skipSpaces(); err != nil {
					return "", errors.New("Syntax error: there must be a next element of an array")
				}
			}
		}
	} else {
		return "", errors.New("Other than string, number, None are supported as JSON value")
	}

	// read comma if exists
	if state.data[state.offset] == byte(',') {
		if err := state.incOffset(); err != nil {
			return "", errors.New("Another key & value pair is expected after a comma")
		}
	}
	if err := state.skipSpaces(); err != nil {
		return "", errors.New("Some non-space character is expected at the end of the line")
	}

	return value, nil
}

func decodeString(state *decodeState) (string, error) {
	// skip prefix "u"
	if state.data[state.offset] == byte('u') {
		if err := state.incOffset(); err != nil {
			return "", errors.New("Syntax Error: expected quoted string after \"u\"")
		}
	}
	// scan strings between quotations
	if state.data[state.offset] == byte('"') {
		return readBetweenQuotes(state, byte('"'))
	} else if state.data[state.offset] == byte('\'') {
		return readBetweenQuotes(state, byte('\''))
	} else {
		return "", errors.New("Syntax Error: expected quoted string")
	}
}

func readBetweenQuotes(state *decodeState, quote byte) (string, error) {
	if err := state.incOffset(); err != nil {
		return "", errors.New("Syntax error: expected end quote \"" + string(quote) + "\"")
	}
	beginIndex := state.offset
	for {
		if state.data[state.offset] == quote {
			if err := state.incOffset(); err != nil {
				// TODO(nekketsuuu): wrong error message?
				return "", errors.New("Syntax error: expected end quote \"" + string(quote) + "\"")
			}
			// Success!
			return state.data[beginIndex : state.offset-1], nil
		} else if state.data[state.offset] == '\\' {
			// Actually there is escaped letters whose letter length is not two, but two is enough.
			if err := state.incOffset(); err != nil {
				return "", errors.New("Syntax error: expected an escaped letter")
			}
			if err := state.incOffset(); err != nil {
				return "", errors.New("Syntax error: expected end quote \"" + string(quote) + "\"")
			}
		} else {
			if err := state.incOffset(); err != nil {
				return "", errors.New("Syntax error: expected end quote \"" + string(quote) + "\"")
			}
		}
	}
}

// tenuki
var patNumber = regexp.MustCompile(`^[+-]?[0-9\.]+`)

func decodeNumber(state *decodeState) (string, error) {
	number := patNumber.FindString(state.data[state.offset:])
	state.offset += len(number)
	if err := state.incOffset(); err != nil {
		return "", errors.New("Syntax error: expected some string after a number")
	}
	return number, nil
}
