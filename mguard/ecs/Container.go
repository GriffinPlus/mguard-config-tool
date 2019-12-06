package ecs

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/griffinplus/mguard-config-tool/mguard/atv"
	log "github.com/sirupsen/logrus"
)

// Container represents a mGuard ECS container.
type Container struct {
	Atv       *atv.Document
	fileCfg   file
	filePass  file
	fileSnmpd file
	fileUsers file
}

type file struct {
	Name string // name of the file in the ECS container
	Data []byte // content of the file in the ECS container
}

// NewContainer returns a new and empty ECS container.
func NewContainer() *Container {

	container := Container{
		Atv:       nil,
		fileCfg:   file{Name: "aca/cfg"},
		filePass:  file{Name: "aca/pass"},
		fileSnmpd: file{Name: "aca/snmpd"},
		fileUsers: file{Name: "aca/users"},
	}

	return &container
}

// ContainerFromATV wraps an ATV document in an ECS container.
func ContainerFromATV(atv *atv.Document) *Container {

	container := NewContainer()
	container.Atv = atv
	// TODO: populate the missing parts with default settings
	return container
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

	container := NewContainer()

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

		if header.Typeflag == tar.TypeReg {

			data, err := ioutil.ReadAll(tarReader)
			if err != nil {
				log.Debugf("    ERROR: %s", err)
				return nil, err
			}

			switch name {
			case container.fileCfg.Name:
				container.fileCfg.Data = data
			case container.filePass.Name:
				container.filePass.Data = data
			case container.fileSnmpd.Name:
				container.fileSnmpd.Data = data
			case container.fileUsers.Name:
				container.fileUsers.Data = data
			}
		}
	}

	// ensure that the container contains the expected configuration file
	if len(container.fileCfg.Data) == 0 {
		log.Errorf("The ECS container does not contain a configuration file at '%s'", container.fileCfg.Name)
		return nil, fmt.Errorf("The ECS container does not contain a configuration file at '%s'", container.fileCfg.Name)
	}

	// parse ATV document stored in the 'aca/cfg' within the ECS container
	log.Debugf("Parsing configuration file '%s' in ECS container...", container.fileCfg.Name)
	atv, err := atv.DocumentFromReader(bytes.NewReader(container.fileCfg.Data))
	if err != nil {
		log.Debugf("Parsing configuration file '%s' in ECS container failed: %s", container.fileCfg.Name, err)
		return nil, err
	}
	container.Atv = atv
	log.Debugf("Parsing configuration file '%s' succeeded.", container.fileCfg.Name)

	log.Debug("Processing ECS container succeeded.")
	return container, nil
}

// ToFile saves the ECS container to the specified file.
func (container *Container) ToFile(path string) error {

	// open the file for writing
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	// write the ECS containre
	return container.ToWriter(file)
}

// ToWriter writes the ECS container to the specified io.Writer.
func (container *Container) ToWriter(writer io.Writer) error {

	log.Debug("Writing ECS container...")

	// update file buffers first to reflect the correct state of the documents
	err := container.updateFileBuffers()
	if err != nil {
		return err
	}

	gzipWriter := gzip.NewWriter(writer)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	time := time.Now()
	err = writeDirectoryToTar(tarWriter, time, "aca")
	if err != nil {
		return err
	}

	files := []file{
		container.fileCfg,
		container.filePass,
		container.fileSnmpd,
		container.fileUsers,
	}

	for _, file := range files {
		if len(file.Data) > 0 {
			log.Debugf("  - file: %s...", file.Name)
			err := writeFileToTar(tarWriter, file, time)
			if err != nil {
				log.Debugf("    ERROR: %s", err)
				return err
			}
		} else {
			log.Debugf("  - file: %s (SKIPPING)", container.fileCfg)
		}
	}

	log.Debug("Writing ECS container succeeded.")
	return nil
}

func (container *Container) updateFileBuffers() error {

	// update configuration in the container
	buffer := bytes.Buffer{}
	err := container.Atv.ToWriter(&buffer)
	if err != nil {
		return err
	}
	container.fileCfg.Data = buffer.Bytes()

	return nil
}

func writeDirectoryToTar(tarWriter *tar.Writer, time time.Time, dirPath string) error {

	header := &tar.Header{
		Typeflag:   tar.TypeDir,
		Name:       dirPath,
		Size:       0,
		Mode:       0700,
		AccessTime: time,
		ChangeTime: time,
		ModTime:    time,
	}

	return tarWriter.WriteHeader(header)
}

func writeFileToTar(tarWriter *tar.Writer, file file, time time.Time) error {

	header := &tar.Header{
		Typeflag:   tar.TypeReg,
		Name:       file.Name,
		Size:       int64(len(file.Data)),
		Mode:       0600,
		AccessTime: time,
		ChangeTime: time,
		ModTime:    time,
	}

	err := tarWriter.WriteHeader(header)
	if err != nil {
		return err
	}

	_, err = tarWriter.Write(file.Data)
	return err
}