package config

import (
	"os"
	"regexp"
	"strconv"
	"strings"
)

const DEFAULT_SECTION = "DEFAULT"

var boolString = map[string]bool{
	"t":     true,
	"true":  true,
	"y":     true,
	"yes":   true,
	"on":    true,
	"1":     true,
	"f":     false,
	"false": false,
	"n":     false,
	"no":    false,
	"off":   false,
	"0":     false,
}

// tValue holds the input position for a value.
type tValue struct {
	position int    // Option order
	v        string // value
}

type Config struct {
	comment   string
	separator string

	// Sections order
	lastIdSection int            // Last section identifier
	idSection     map[string]int // Section : position

	// The last option identifier used for each section.
	lastIdOption map[string]int // Section : last identifier

	// Section -> option : value
	data map[string]map[string]*tValue
}

func ReadFile(fname string) (*Config, error) {
	return __read(fname, __new(false, true))
}

var varRegExp = regexp.MustCompile(`%\(([a-zA-Z0-9_.\-]+)\)s`)  // %(variable)s
var envVarRegExp = regexp.MustCompile(`\${([a-zA-Z0-9_.\-]+)}`) // ${envvar}

func (c *Config) GetString(section string, option string) (value string, err error) {
	value, err = c.rawString(section, option)
	if err != nil {
		return "", err
	}

	// % variables
	computedVal, err := c.computeVar(&value, varRegExp, 2, 2, func(varName *string) string {
		lowerVar := *varName
		// search variable in default section as well as current section
		varVal, _ := c.data[DEFAULT_SECTION][lowerVar]
		if _, ok := c.data[section][lowerVar]; ok {
			varVal = c.data[section][lowerVar]
		}
		return varVal.v
	})
	value = *computedVal

	if err != nil {
		return value, err
	}

	// $ environment variables
	computedVal, err = c.computeVar(&value, envVarRegExp, 2, 1, func(varName *string) string {
		return os.Getenv(*varName)
	})
	value = *computedVal
	return value, err
}

func (c *Config) GetValue(section string, option string, defaultValue interface{}) interface{} {
	if section == "" {
		section = DEFAULT_SECTION
	}
	s, err := c.GetString(section, option)
	if err != nil {
		return defaultValue
	}
	switch defaultValue.(type) {
	case string:
		return s
	case int:
		v, err1 := strconv.Atoi(s)
		if err1 != nil {
			return defaultValue
		}
		return v
	case bool:
		v, ok := boolString[strings.ToLower(s)]
		if !ok {
			return defaultValue
		}
		return v
	case float64:
		v, err1 := strconv.ParseFloat(s, 64)
		if err1 != nil {
			return defaultValue
		}
		return v
	}
	return nil
}
