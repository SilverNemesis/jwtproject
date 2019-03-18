package dotenv

import (
	"bufio"
	"errors"
	"os"
	"strings"
)

var tableVariable [256]bool

func init() {
	for i := 0; i < 256; i++ {
		if i >= 'a' && i <= 'z' {
			tableVariable[i] = true
		} else if i >= 'A' && i <= 'Z' {
			tableVariable[i] = true
		} else if i >= '0' && i <= '9' {
			tableVariable[i] = true
		} else if i == '_' {
			tableVariable[i] = true
		}
	}
}

// ProcessFile sets environment variables based on the contents of a file
func ProcessFile(filename string) error {
	lines, err := LoadFile(filename)
	if err != nil {
		return err
	}
	envMap, err := ParseLines(lines)
	currentEnv := map[string]bool{}
	variables := os.Environ()
	for _, variable := range variables {
		key := strings.Split(variable, "=")[0]
		currentEnv[key] = true
	}
	for key, value := range envMap {
		if !currentEnv[key] {
			os.Setenv(key, value)
		}
	}
	return nil
}

// LoadFile loads a text files as an array of strings
func LoadFile(filename string) (lines []string, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err = scanner.Err(); err != nil {
		return
	}
	return
}

// ParseLines parses an array of strings and return a key-value map
func ParseLines(lines []string) (envMap map[string]string, err error) {
	envMap = make(map[string]string)
	for _, fullLine := range lines {
		key, value, _, err := parseLine(fullLine)
		if err != nil {
			break
		}
		if key != "" {
			envMap[key] = value
		}
	}
	return
}

func parseLine(line string) (key, value, comment string, err error) {
	state := 'k'
	max := len(line)
	var cur, nxt int
	for cur = 0; cur < max; cur = nxt {
		for ; cur < max; cur++ {
			if line[cur] != ' ' {
				break
			}
		}
		if cur == max {
			break
		}
		if line[cur] == '#' {
			comment = line[cur:max]
			break
		} else {
			if state == 'k' {
				for nxt = cur + 1; nxt < max; nxt++ {
					if !tableVariable[line[nxt]] {
						break
					}
				}
				if line[cur] >= '0' && line[cur] <= '9' {
					err = errors.New("invalid key")
					break
				}
				key = line[cur:nxt]
				state = '='
			} else if state == '=' {
				if line[cur] == '=' {
					nxt = cur + 1
					state = 'v'
				} else {
					err = errors.New("invalid key")
					break
				}
			} else if state == 'v' {
				if line[cur] == '"' {
					cur++
					for nxt = cur; nxt < max; nxt++ {
						if line[nxt] == '"' {
							break
						}
					}
					if nxt < max {
						value = line[cur:nxt]
						nxt++
					} else {
						err = errors.New("mismatched quotes")
						break
					}
				} else {
					for nxt = cur + 1; nxt < max; nxt++ {
						if line[nxt] == '#' {
							break
						}
					}
					for nxt > cur && line[nxt-1] == ' ' {
						nxt--
					}
					value = line[cur:nxt]
				}
				state = 'c'
			}
		}

		if cur == nxt {
			break
		}
	}
	return
}
