package certmgr

import (
	"archive/tar"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/grantae/certinfo"
	"github.com/mozilla-services/pkcs7"
	log "github.com/sirupsen/logrus"
)

func downloadDeviceCertificate(serial, user, password string) (*x509.Certificate, error) {

	maxAttempts := 5
	for attempt := 1; attempt <= maxAttempts; attempt++ {

		// download the certificate of the mGuard with the specified serial number
		log.Infof("Downloading device certificate...")
		url := fmt.Sprintf("http://online.license.innominate.com/cgi-bin/autodevcert.cgi?SERIALNUMBER=%s&USERNAME=%s&PASSWORD=%s", serial, user, password)
		response, err := http.Get(url)
		if err != nil {
			log.Errorf("Downloading device certificate failed (attempt %d/%d): %s", attempt, maxAttempts, err)
			continue
		}
		defer response.Body.Close()

		// abort, if the device database responded with an error
		if response.StatusCode != 200 {
			log.Errorf("Downloading device certificate failed, the device database returned '%s'. Aborting...", response.Status)
			return nil, fmt.Errorf("Downloading device certificate failed, the device database returned '%s'. Aborting...", response.Status)
		}

		// the certificate is expected to be in a tar file
		// => try to extract it
		log.Debug("Processing downloaded device certificate container...")
		tarReader := tar.NewReader(response.Body)
		for {
			header, err := tarReader.Next()

			if err == io.EOF {
				break
			}

			if err != nil {
				return nil, err
			}

			// name := header.Name
			if header.Typeflag == tar.TypeReg {

				// the name of the file containing the certificate has the extension *.tpmdevcrt
				if path.Ext(header.Name) == ".tpmdevcrt" {

					data, err := ioutil.ReadAll(tarReader)
					if err != nil {
						log.Debugf("    ERROR: %s", err)
						return nil, err
					}

					// parse certificate(s) in file
					certs, err := readCertificatesFromBuffer(data)
					if err != nil {
						return nil, err
					}

					// return certificate
					// (the downloaded file should only contain one certificate)
					return certs[0], nil
				}
			}
		}
	}

	// downloading failed
	return nil, fmt.Errorf("Downloading the certificate failed")
}

// LoadCertificatesFromFile loads specified file and parses it as a container of X.509 certificates.
func loadCertificatesFromFile(path string) ([]*x509.Certificate, error) {

	// read file
	log.Debugf("Reading file (%s)...", path)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Errorf("The specified file (%s) does not exist.", path)
			return nil, fmt.Errorf("The specified file (%s) does not exist", path)
		}
		log.Errorf("Reading the specified file (%s) failed: %s", path, err)
		return nil, fmt.Errorf("Reading the specified file (%s) failed: %s", path, err)
	}

	return readCertificatesFromBuffer(data)
}

func readCertificatesFromBuffer(data []byte) ([]*x509.Certificate, error) {

	// try to parse file
	log.Debug("Processing certificates...")
	certs, err := parseCertificates(data)
	if err != nil {
		log.Errorf("Parsing certificates failed: %s", err)
		return nil, fmt.Errorf("Parsing certificates failed: %s", err)
	}

	// print certificates
	for i, cert := range certs {
		printCertificate(cert, fmt.Sprintf("Parsing certificate (%d/%d) succeeded...", i+1, len(certs)))
	}

	return certs, nil
}

func parseCertificates(data []byte) ([]*x509.Certificate, error) {

	var certificates []*x509.Certificate

	// try to parse data as DER encoded certificates
	// (the DER encoded certificates must be concatened without any space in between)
	certs, err := x509.ParseCertificates(data)
	if err == nil {
		return certs, nil
	}

	// try to parse data as a PEM encoded certificates (skip private keys, if present)
	var block *pem.Block
	for rest := data; len(rest) > 0; {

		// decode block
		block, rest = pem.Decode(rest)
		if block != nil {

			var blockData []byte
			if !x509.IsEncryptedPEMBlock(block) {

				blockData = block.Bytes

				// evaluate PEM block
				if block.Type == "CERTIFICATE" {

					cert, err := x509.ParseCertificate(blockData)
					if err != nil {
						log.Errorf("Data contains a PEM encoded 'CERTIFICATE' block, but parsing it failed: %s", err)
						return nil, fmt.Errorf("Data contains a PEM encoded 'CERTIFICATE' block, but parsing it failed: %s", err)
					}

					certificates = append(certificates, cert)

				} else if block.Type == "RSA PRIVATE KEY" { // PKCS#1 (RSA only)

					// A private key => skip!
					log.Infof("Data contains a PEM encoded 'RSA PRIVATE KEY' block, skipping block...")

				} else if block.Type == "PRIVATE KEY" { // PKCS#8

					// a private key => skip!
					log.Infof("Data contains a PEM encoded 'PRIVATE KEY' block, skipping block...")

				} else if block.Type == "ENCRYPTED PRIVATE KEY" { // PKCS#8

					// a private key => skip!
					log.Infof("Data contains a PEM encoded 'ENCRYPTED PRIVATE KEY' block, skipping block...")

				} else {

					log.Warningf("Data contains an unknown PEM encoded block (%s). skipping block...", block.Type)

				}

			} else {

				log.Infof("Data contains an encrypted PEM block '%s', skipping block...", block.Type)

			}

			// decoding succeeded
			// => file seems to be valid text
			// => trim whitespaces to ensure termination condition works properly with trailing whitespaces
			rest = []byte(strings.TrimSpace(string(rest)))
			continue
		}

		// PEM encoded block was not found...
		break
	}

	// abort, if the file is a PEM formatted file that contains at least one certificate
	if len(certificates) > 0 {
		return certificates, nil
	}

	// try to parse data as PKCS#7 archive (can only contain certificates)
	pkcs7, err := pkcs7.Parse(data)
	if err == nil {
		return pkcs7.Certificates, nil
	}

	log.Errorf("Data does not contain any processable certificates.")
	return nil, fmt.Errorf("Data does not contain any processable certificates")
}

func printCertificate(certificate *x509.Certificate, message string) {

	if log.IsLevelEnabled(log.DebugLevel) {
		result, err := certinfo.CertificateText(certificate)
		if err != nil {
			log.Errorf("Printing certificate failed: %s", err)
			return
		}
		log.Debug(message, "\n", result)
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
