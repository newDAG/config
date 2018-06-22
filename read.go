package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode"
)

const (
	_DEPTH_VALUES         = 200
	DEFAULT_COMMENT       = "# "
	ALTERNATIVE_COMMENT   = "; "
	DEFAULT_SEPARATOR     = ":"
	ALTERNATIVE_SEPARATOR = "="
)

// New creates an empty configuration representation.
// This representation can be filled with AddSection and AddOption and then

// comment: has to be `DEFAULT_COMMENT` or `ALTERNATIVE_COMMENT`
// separator: has to be `DEFAULT_SEPARATOR` or `ALTERNATIVE_SEPARATOR`
// preSpace: indicate if is inserted a space before of the separator
// postSpace: indicate if is added a space after of the separator
func __new(preSpace, postSpace bool) *Config {
	comment := DEFAULT_COMMENT
	separator := DEFAULT_SEPARATOR

	// == Get spaces around separator
	if preSpace {
		separator = " " + separator
	}

	if postSpace {
		separator += " "
	}
	//==

	c := new(Config)

	c.comment = comment
	c.separator = separator
	c.idSection = make(map[string]int)
	c.lastIdOption = make(map[string]int)
	c.data = make(map[string]map[string]*tValue)

	c.AddSection(DEFAULT_SECTION) // Default section always exists.

	return c
}

func __read(fname string, c *Config) (*Config, error) {
	file, err := os.Open(fname)
	if err != nil {
		return nil, err
	}

	if err = c.read(bufio.NewReader(file)); err != nil {
		return nil, err
	}

	if err = file.Close(); err != nil {
		return nil, err
	}

	return c, nil
}

func stripComments(l string) string {
	// Comments are preceded by space or TAB
	for _, c := range []string{" ;", "\t;", " #", "\t#"} {
		if i := strings.Index(l, c); i != -1 {
			l = l[0:i]
		}
	}
	return l
}

func (c *Config) read(buf *bufio.Reader) (err error) {
	var section, option string
	var scanner = bufio.NewScanner(buf)
	for scanner.Scan() {
		l := strings.TrimRightFunc(stripComments(scanner.Text()), unicode.IsSpace)

		// Switch written for readability (not performance)
		switch {
		// Empty line and comments
		case len(l) == 0, l[0] == '#', l[0] == ';':
			continue

		// New section. The [ must be at the start of the line
		case l[0] == '[' && l[len(l)-1] == ']':
			option = "" // reset multi-line value
			section = strings.TrimSpace(l[1 : len(l)-1])
			c.AddSection(section)

		// Continuation of multi-line value
		// starts with whitespace, we're in a section and working on an option
		case section != "" && option != "" && (l[0] == ' ' || l[0] == '\t'):
			prev, _ := c.rawString(section, option)
			value := strings.TrimSpace(l)
			c.AddOption(section, option, prev+"\n"+value)

		// Other alternatives
		default:
			i := strings.IndexAny(l, "=:")

			switch {
			// Option and value
			case i > 0 && l[0] != ' ' && l[0] != '\t': // found an =: and it's not a multiline continuation
				option = strings.TrimSpace(l[0:i])
				value := strings.TrimSpace(l[i+1:])
				c.AddOption(section, option, value)

			default:
				return errors.New("could not parse line: " + l)
			}
		}
	}
	return scanner.Err()
}

//-----section---------------------------------------------------------------------------
func (c *Config) AddSection(section string) bool {
	// DEFAULT_SECTION
	if section == "" {
		return false
	}

	if _, ok := c.data[section]; ok {
		return false
	}

	c.data[section] = make(map[string]*tValue)

	// Section order
	c.idSection[section] = c.lastIdSection
	c.lastIdSection++

	return true
}

// HasSection checks if the configuration has the given section.
// (The default section always exists.)
func (c *Config) HasSection(section string) bool {
	_, ok := c.data[section]

	return ok
}

// Sections returns the list of sections in the configuration.
// (The default section always exists).
func (c *Config) Sections() (sections []string) {
	sections = make([]string, len(c.idSection))
	pos := 0 // Position in sections

	for i := 0; i < c.lastIdSection; i++ {
		for section, id := range c.idSection {
			if id == i {
				sections[pos] = section
				pos++
			}
		}
	}

	return sections
}

//---option------------------------------------------------------------------------------------

func (c *Config) AddOption(section string, option string, value string) bool {
	c.AddSection(section) // Make sure section exists

	if section == "" {
		section = DEFAULT_SECTION
	}

	_, ok := c.data[section][option]

	c.data[section][option] = &tValue{c.lastIdOption[section], value}
	c.lastIdOption[section]++

	return !ok
}

func (c *Config) HasOption(section string, option string) bool {
	if _, ok := c.data[section]; !ok {
		return false
	}

	_, okd := c.data[DEFAULT_SECTION][option]
	_, oknd := c.data[section][option]

	return okd || oknd
}

//--get value---------------------------------------------------------------------------------
// Substitutes values, calculated by callback, on matching regex
func (c *Config) computeVar(beforeValue *string, regx *regexp.Regexp, headsz, tailsz int, withVar func(*string) string) (*string, error) {
	var i int
	computedVal := beforeValue
	for i = 0; i < _DEPTH_VALUES; i++ { // keep a sane depth

		vr := regx.FindStringSubmatchIndex(*computedVal)
		if len(vr) == 0 {
			break
		}

		varname := (*computedVal)[vr[headsz]:vr[headsz+1]]
		varVal := withVar(&varname)
		if varVal == "" {
			return &varVal, errors.New(fmt.Sprintf("Option not found: %s", varname))
		}

		// substitute by new value and take off leading '%(' and trailing ')s'
		//  %(foo)s => headsz=2, tailsz=2
		//  ${foo}  => headsz=2, tailsz=1
		newVal := (*computedVal)[0:vr[headsz]-headsz] + varVal + (*computedVal)[vr[headsz+1]+tailsz:]
		computedVal = &newVal
	}

	if i == _DEPTH_VALUES {
		retVal := ""
		return &retVal,
			fmt.Errorf("Possible cycle while unfolding variables: max depth of %d reached", _DEPTH_VALUES)
	}

	return computedVal, nil
}

// rawString gets the (raw) string value for the given option in the section.
// The raw string value is not subjected to unfolding, which was illustrated in
// the beginning of this documentation.
//
// It returns an error if either the section or the option do not exist.
func (c *Config) rawString(section string, option string) (value string, err error) {
	if _, ok := c.data[section]; ok {
		if tValue, ok := c.data[section][option]; ok {
			return tValue.v, nil
		}
	}
	return "", errors.New("option not found: " + option)
}
