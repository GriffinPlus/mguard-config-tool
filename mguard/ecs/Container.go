package ecs

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/griffinplus/mguard-config-tool/mguard/atv"
	"github.com/griffinplus/mguard-config-tool/shadow"
	log "github.com/sirupsen/logrus"
)

// Container represents a mGuard ECS container.
type Container struct {
	Atv       *atv.File
	Users     *shadow.File
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
		Users:     nil,
		fileCfg:   file{Name: "aca/cfg"},
		filePass:  file{Name: "aca/pass", Data: []byte(DefaultPassFileContent)},
		fileSnmpd: file{Name: "aca/snmpd", Data: []byte(DefaultSnmpdFileContent)},
		fileUsers: file{Name: "aca/users"},
	}

	return &container
}

// ContainerFromATV wraps an ATV document in an ECS container.
func ContainerFromATV(atv *atv.File) *Container {

	container := NewContainer()
	container.Atv = atv
	container.Users = createDefaultShadowFile()
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

	// parse ATV document stored within the ECS container
	log.Debugf("Parsing configuration file '%s' in ECS container...", container.fileCfg.Name)
	atv, err := atv.FromReader(bytes.NewReader(container.fileCfg.Data))
	if err != nil {
		log.Debugf("Parsing configuration file '%s' in ECS container failed: %s", container.fileCfg.Name, err)
		return nil, err
	}
	container.Atv = atv
	log.Debugf("Parsing configuration file '%s' succeeded.", container.fileCfg.Name)

	// ensure that the container contains the expected users file
	if len(container.fileUsers.Data) == 0 {
		log.Errorf("The ECS container does not contain a password file at '%s'", container.fileUsers.Name)
		return nil, fmt.Errorf("The ECS container does not contain a password file at '%s'", container.fileUsers.Name)
	}

	// load users file stored within the ECS container
	log.Debugf("Parsing user file '%s' in ECS container...", container.fileUsers.Name)
	users, err := shadow.FileFromReader(bytes.NewReader(container.fileUsers.Data))
	if err != nil {
		log.Debugf("Parsing user file '%s' in ECS container failed: %s", container.fileCfg.Name, err)
		return nil, err
	}
	container.Users = users
	log.Debugf("Parsing user file '%s' succeeded.", container.fileUsers.Name)

	log.Debug("Processing ECS container succeeded.")
	return container, nil
}

// Dupe returns a copy of the ECS container.
func (container *Container) Dupe() *Container {

	copy := Container{
		Atv:       container.Atv.Dupe(),
		Users:     container.Users.Dupe(),
		fileCfg:   container.fileCfg,
		filePass:  container.filePass,
		fileSnmpd: container.fileSnmpd,
		fileUsers: container.fileUsers,
	}

	return &copy
}

// ToFile saves the ECS container to the specified file.
func (container *Container) ToFile(path string) error {

	// create directories on the way, if necessary
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return err
	}

	// open the file for writing
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	// write the ECS container
	return container.ToWriter(file)
}

// ToEncryptedFile writes the ECS container encrypted with the specified device certifivate to the specified file.
func (container *Container) ToEncryptedFile(path string, deviceCertificate *x509.Certificate) error {

	// determine the path of the openssl executable
	opensslExecutablePath, err := GetOpensslExecutablePath()
	if err != nil {
		return err
	}

	// create temporary directory to perform the encryption in
	scratchDir, err := ioutil.TempDir("", "ecs-encryption")
	if err != nil {
		log.Errorf("%v", err)
		return err
	}
	defer os.RemoveAll(scratchDir)

	// save device certificate (PEM encoded)
	block := &pem.Block{
		Type:    "CERTIFICATE",
		Headers: map[string]string{},
		Bytes:   deviceCertificate.Raw,
	}
	err = ioutil.WriteFile(filepath.Join(scratchDir, "device.pem"), pem.EncodeToMemory(block), 644)
	if err != nil {
		return err
	}

	// write ECS container into a buffer
	var ecsBuffer bytes.Buffer
	err = container.ToWriter(&ecsBuffer)
	if err != nil {
		log.Errorf("Serializing ECS container failed: %s", err)
		return err
	}

	// encrypt the ECS container
	var encryptedEcs bytes.Buffer
	var stderr bytes.Buffer
	opensslCmd := exec.Command(opensslExecutablePath, "smime", "-encrypt", "-binary", "-outform", "PEM", "device.pem")
	opensslCmd.Dir = scratchDir
	opensslCmd.Stdin = bytes.NewReader(ecsBuffer.Bytes())
	opensslCmd.Stdout = &encryptedEcs
	opensslCmd.Stderr = &stderr
	err = opensslCmd.Run()
	if err != nil {
		log.Debugf("OpenSSL failed:\n%s\n", stderr.String())
		log.Errorf("Encrypting ECS container failed: %s", err)
		return err
	}

	// write the encrypted ECS container to the final destination
	log.Infof("Writing encrypted ECS file (%s)...", path)
	err = ioutil.WriteFile(path, encryptedEcs.Bytes(), 644)
	if err != nil {
		log.Errorf("Writing encrypted ECS file (%s) failed: %s", path, err)
		return err
	}

	return nil
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

	// update the configuration in the container
	if container.Atv != nil {
		log.Debugf("Updating '%s' in ECS container...", container.fileCfg.Name)
		buffer := bytes.Buffer{}
		err := container.Atv.ToWriter(&buffer)
		if err != nil {
			return err
		}
		container.fileCfg.Data = buffer.Bytes()
	}

	// update the user file in the container
	if container.Users != nil {
		log.Debugf("Updating '%s' in ECS container...", container.fileUsers.Name)
		buffer := bytes.Buffer{}
		err := container.Users.ToWriter(&buffer)
		if err != nil {
			return err
		}
		container.fileUsers.Data = buffer.Bytes()
	}

	return nil
}

// writeDirectoryToTar writes the specified directory into the tar archive.
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

// writeFileToTar writes the specified regular file into the tar archive.
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
