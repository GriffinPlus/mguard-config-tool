package ecs

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"

	"github.com/griffinplus/mguard-config-tool/mguard/atv"
	log "github.com/sirupsen/logrus"
)

// Container represents a mGuard ECS container.
type Container struct {
	Atv       *atv.Document // the configuration file from 'aca/cfg' as a document
	fileCfg   []byte        // content of the 'aca/cfg' file in the ECS container
	filePass  []byte        // content of the 'aca/pass' file in the ECS container
	fileSnmpd []byte        // content of the 'aca/snmpd' file in the ECS container
	fileUsers []byte        // content of the 'aca/users' file in the ECS container
}

// ContainerFromATV wraps an ATV document in an ECS container.
func ContainerFromATV(atv *atv.Document) *Container {

	// TODO: populate the missing parts with default settings

	return &Container{
		Atv:       atv,
		fileCfg:   nil,
		filePass:  nil,
		fileSnmpd: nil,
		fileUsers: nil,
	}
}

// ContainerFromFile reads the specified ECS container from disk.
func ContainerFromFile(path string) (*Container, error) {

	// open the file for reading
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// read the ATV file
	return ContainerFromReader(file)
}

// ContainerFromReader reads an ECS container from the specified io.Reader.
func ContainerFromReader(reader io.Reader) (*Container, error) {

	container := &Container{}

	// an ECS container is a simple gzip'ed tar archive
	// => unzip and extract files from the archive
	log.Debug("Processing ECS container...")

	gzf, err := gzip.NewReader(reader)
	if err != nil {
		return nil, err
	}

	tarReader := tar.NewReader(gzf)

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		name := header.Name
		log.Debugf("  - file: %s...", name)

		switch header.Typeflag {

		case tar.TypeDir:
			continue

		case tar.TypeReg:

			err = nil

			switch name {
			case "aca/cfg":
				container.fileCfg, err = ioutil.ReadAll(tarReader)
			case "aca/pass":
				container.filePass, err = ioutil.ReadAll(tarReader)
			case "aca/snmpd":
				container.fileSnmpd, err = ioutil.ReadAll(tarReader)
			case "aca/users":
				container.fileUsers, err = ioutil.ReadAll(tarReader)
			}

			if err != nil {
				log.Debugf("    ERROR: %s", err)
				return nil, err
			}

		default:
			log.Panicf("Unhandled type flag (%c) in TAR archive", header.Typeflag)
		}
	}

	// parse ATV document stored in the 'aca/cfg' within the ECS container
	log.Debug("Parsing configuration file 'aca/cfg' in ECS container...")
	atv, err := atv.DocumentFromReader(bytes.NewReader(container.fileCfg))
	if err != nil {
		log.Debugf("Parsing configuration file 'aca/cfg' in ECS container failed: %s", err)
		return nil, err
	}
	container.Atv = atv
	log.Debug("Parsing configuration file 'aca/cfg' succeeded.")

	log.Debug("Processing ECS container succeeded.")
	return container, nil
}

// ToFile saves the ECS container to the specified file.
func (doc *Container) ToFile(path string) error {

	// open the file for writing
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// write the ECS containre
	return doc.ToWriter(file)
}

// ToWriter writes the ECS container to the specified io.Writer.
func (doc *Container) ToWriter(writer io.Writer) error {
	// TODO
	return nil
}
