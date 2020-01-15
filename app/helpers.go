package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/griffinplus/mguard-config-tool/mguard/atv"
	"github.com/griffinplus/mguard-config-tool/mguard/ecs"
	log "github.com/sirupsen/logrus"
)

// loadConfigurationFile loads the specified ATV/ECS file and returns an ECS container with the mGuard configuration.
// If the file is an ATV file, the missing parts in the ECS container are filled with defaults. If the path is an
// empty string and stdin is not a console, it trys to read the ATV/ECS file from stdin.
func loadConfigurationFile(path string) (*ecs.Container, error) {

	if len(path) == 0 {

		// file was not specified, try to read the file from stdin
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			// read ATV/ECS file from stdin
			log.Info("Trying to read file from stdin...")
			data, err := ioutil.ReadAll(os.Stdin)
			log.Infof("Read %d bytes from stdin.", len(data))
			if err != nil {
				return nil, err
			}

			// try to read ECS container from stdin
			log.Info("Trying to interpret data as ECS container...")
			container, err := ecs.ContainerFromReader(bytes.NewBuffer(data))
			if err == nil {
				log.Info("Reading ECS container succeeded.")
				return container, nil
			}
			log.Infof("Reading ECS container failed: %s", err)

			// try to read ATV container from stdin
			log.Info("Trying to interpret data as ATV document...")
			atv, err := atv.DocumentFromReader(bytes.NewBuffer(data))
			if err == nil {
				log.Info("Reading ATV file succeeded.")
				container := ecs.ContainerFromATV(atv)
				return container, nil
			}
			log.Infof("Reading ATV file failed: %s", err)

			return nil, fmt.Errorf("Data piped in via stdin does not seem to be an ECS/ATV file")
		}

		return nil, fmt.Errorf("Configuration file was not specified and stdin is no pipe")

	} else {

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
				container := ecs.ContainerFromATV(atv)
				return container, nil

			case "ecs":
				log.Infof("Trying to interpret file (%s) as an ECS file...", path)
				container, err := ecs.ContainerFromFile(path)
				if err != nil {
					log.Infof("Reading file (%s) failed: %s", path, err)
					continue
				}
				log.Infof("Reading file (%s) succeeded.", path)
				return container, nil

			default:
				log.Panic("Unhandled document format")
			}
		}

		// the file could neither be read as an ECS container nor as an ATV file
		return nil, fmt.Errorf("Loading file (%s) failed", path)
	}
}

// exePath gets the full path of the executable.
func exePath() (string, error) {

	prog := os.Args[0]
	p, err := filepath.Abs(prog)
	if err != nil {
		return "", err
	}

	fi, err := os.Stat(p)
	if err == nil {
		if !fi.Mode().IsDir() {
			return p, nil
		}
		err = fmt.Errorf("%s is directory", p)
	}

	if filepath.Ext(p) == "" {
		p += ".exe"
		fi, err := os.Stat(p)
		if err == nil {
			if !fi.Mode().IsDir() {
				return p, nil
			}
			err = fmt.Errorf("%s is directory", p)
		}
	}

	return "", err
}
