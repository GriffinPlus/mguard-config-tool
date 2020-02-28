package atv

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// MergeConfiguration defines which settings should be merged from one ATV document into another.
type MergeConfiguration struct {
	paths documentSettingPaths
}

var commentRegex = regexp.MustCompile(`^\s*([^#]*)\s*(#.*)?$`)

// LoadMergeConfiguration loads a merge configuration file.
func LoadMergeConfiguration(path string) (*MergeConfiguration, error) {

	// open file for reading
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// read file
	config := MergeConfiguration{}
	scanner := bufio.NewScanner(file)
	for lineNo := 1; scanner.Scan(); lineNo++ {

		// remove preceding and trailing whitespaces
		line := strings.TrimSpace(scanner.Text())

		// remove comments
		matches := commentRegex.FindStringSubmatch(line)
		if len(matches[0]) == 0 || len(matches[1]) == 0 {
			continue // empty line or a line with just a comment
		}

		lineWithoutComment := matches[1]
		settingPath, err := parseDocumentSettingPath(lineWithoutComment)
		if err != nil {
			return nil, fmt.Errorf("Reading merge configuration file failed (line: %d). Error: %s", lineNo, err)
		}
		config.paths = append(config.paths, settingPath)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &config, nil
}

// ShouldMergeSetting indicates whether the specified setting should be merged.
func (cfg *MergeConfiguration) ShouldMergeSetting(path documentSettingPath) bool {

	if cfg == nil {
		return false
	}

	for _, p := range cfg.paths {
		if p.String() == path.String() {
			return true
		}
	}

	return false
}
