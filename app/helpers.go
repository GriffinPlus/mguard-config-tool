package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/griffinplus/mguard-config-tool/mguard/atv"
	"github.com/griffinplus/mguard-config-tool/mguard/ecs"
	log "github.com/sirupsen/logrus"
)

// loadConfigurationFile loads an ATV file or an ECS container from the specified file and returns an ECS container
// with the mGuard configuration. If the file is an ATV file, the missing parts in the ECS container are filled with
// defaults.
func loadConfigurationFile(path string) (*ecs.Container, error) {

	log.Infof("Trying to load file (%s)...", path)

	ext := strings.ToLower(filepath.Ext(path))
	var tryOrder []string
	if ext == ".atv" {
		// this is probably an ATV file
		log.Infof("File (%s) has the extension '%s'. This could be an ATV file.", path, ext)
		tryOrder = []string{"atv", "ecs"}
	} else if ext == ".tgz" {
		// this could be an ECS container
		log.Infof("File (%s) has the extension '%s'. This could be an ECS file.", path, ext)
		tryOrder = []string{"ecs", "atv"}
	} else {
		// cannot give an educated guess
		// => try both and check whether one works...
		log.Infof("File (%s) has the extension '%s'. Cannot guess the configuration file type from the file extension.", path, ext)
		tryOrder = []string{"ecs", "atv"}
	}

	for _, format := range tryOrder {
		switch format {

		case "atv":
			log.Infof("Trying to interpret file (%s) as an ATV file...", path)
			atv, err := atv.DocumentFromFile(path)
			if err != nil {
				log.Infof("Reading file (%s) failed: %s", path, err)
				continue
			}
			log.Infof("Reading file (%s) succeeded.", path)
			ecs := ecs.ContainerFromATV(atv)
			return ecs, nil

		case "ecs":
			log.Infof("Trying to interpret file (%s) as an ECS file...", path)
			ecs, err := ecs.ContainerFromFile(path)
			if err != nil {
				log.Infof("Reading file (%s) failed: %s", path, err)
				continue
			}
			log.Infof("Reading file (%s) succeeded.", path)
			return ecs, nil

		default:
			log.Panic("Unhandled document format")
		}
	}

	// the file could neither be read as an ECS container nor as an ATV file
	return nil, fmt.Errorf("Loading file (%s) failed", path)
}
