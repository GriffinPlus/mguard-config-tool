package certmgr

import (
	"archive/tar"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// CertificateManager represents the mGuard device certificate manager.
type CertificateManager struct {
	certificateCacheDirectory string
	deviceDatabaseUser        string
	deviceDatabasePassword    string
}

// NewCertificateManager returns a new device certificate manager.
// The cache path may be empty to disable caching.
// Username and password may be empty to let the certificate manager look into the mguard-device-database.yaml for credentials.
func NewCertificateManager(certificateCacheDirectory, deviceDatabaseUser, deviceDatabasePassword string) (*CertificateManager, error) {

	mgr := CertificateManager{
		certificateCacheDirectory: certificateCacheDirectory,
		deviceDatabaseUser:        deviceDatabaseUser,
		deviceDatabasePassword:    deviceDatabasePassword,
	}

	// load credentials file, if username/password is empty
	if len(mgr.deviceDatabaseUser) == 0 || len(mgr.deviceDatabasePassword) == 0 {

		log.Debugf("Loading device database credentials from 'mguard-device-database.yaml'...")

		// create new device database configuration
		conf := viper.New()
		conf.SetDefault("credentials.user", "")
		conf.SetDefault("credentials.password", "")

		// load device database configuration from 'mguard-device-database.yaml'
		exePath, _ := exePath()
		configPath := filepath.Join(filepath.Dir(exePath), "mguard-device-database.yaml")
		basename := filepath.Base(configPath)
		configName := strings.TrimSuffix(basename, filepath.Ext(basename))
		configDir := filepath.Dir(configPath) + string(os.PathSeparator)
		conf.SetConfigName(configName)
		conf.SetConfigType("yaml")
		conf.AddConfigPath(configDir)
		err := conf.ReadInConfig()
		if err != nil {
			log.Errorf("Loading device database credentials from 'mguard-device-database.yaml' failed: %s", err)
			return nil, err
		}

		// fetch settings out of the configuration
		mgr.deviceDatabaseUser = conf.GetString("credentials.user")
		mgr.deviceDatabasePassword = conf.GetString("credentials.password")
	}

	return &mgr, nil
}

// GetCertificate tries to get the device certificate for the mGuard with the specified serial number.
// It serves the certificate from its cache, if possible, and downloads the the certificate from the
// device database, if necessary.
func (mgr *CertificateManager) GetCertificate(serial string) (*x509.Certificate, error) {

	// try to fetch the certificate from the cache
	if len(mgr.certificateCacheDirectory) > 0 {
		certificate, _ := mgr.getCertificateFromCache(serial)
		if certificate != nil {
			return certificate, nil
		}
	}

	// try to download the certificate from the device database
	certificate, err := mgr.downloadCertificate(serial)
	if err != nil {
		return nil, err
	}

	// put certificate into the cache
	if len(mgr.certificateCacheDirectory) > 0 {
		mgr.putCertificateIntoCache(serial, certificate)
	}

	return certificate, nil
}

// downloadCertificate tries to download the device certificate for the mGuard with the specified serial number from the device database.
func (mgr *CertificateManager) downloadCertificate(serial string) (*x509.Certificate, error) {

	// ensure that the username/password for the device database is initialized
	if len(mgr.deviceDatabaseUser) == 0 || len(mgr.deviceDatabasePassword) == 0 {
		return nil, fmt.Errorf("The username/password for the device database is not set.")
	}

	// the certificate could not be served from the cache
	// => try to download it from the device database
	maxAttempts := 5
	for attempt := 1; attempt <= maxAttempts; attempt++ {

		// download the certificate of the mGuard with the specified serial number
		log.Infof("Downloading device certificate...")
		url := fmt.Sprintf("http://online.license.innominate.com/cgi-bin/autodevcert.cgi?SERIALNUMBER=%s&USERNAME=%s&PASSWORD=%s", serial, mgr.deviceDatabaseUser, mgr.deviceDatabasePassword)
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
	return nil, fmt.Errorf("Downloading device certificate failed")
}

// getCachedCertificate checks whether the certificate for the mGuard with the specified serial number
// is cached and returns it. The function returns nil, if there is no such certificate in the cache.
func (mgr *CertificateManager) getCertificateFromCache(serial string) (*x509.Certificate, error) {

	// try to open the certificate file
	path := filepath.Join(mgr.certificateCacheDirectory, serial+".der")
	data, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	// try to parse data as DER encoded certificates
	certs, err := x509.ParseCertificates(data)
	if err != nil {
		// there is something wrong with the cached certificate
		// => delete the cached file and pretend that it never existed, so it is downloaded next
		log.Errorf("Certificate file (%s) exists, but it could not be parsed. Removing it from the cache.\nERROR: %s", path, err)
		os.Remove(path)
		return nil, nil
	}

	return certs[0], nil
}

// cacheCertificate puts or updates the certificate of the mGuard with the specified serial number.
func (mgr *CertificateManager) putCertificateIntoCache(serial string, certificate *x509.Certificate) error {

	path := filepath.Join(mgr.certificateCacheDirectory, serial+".der")

	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err == nil {
		err = ioutil.WriteFile(path, certificate.Raw, 644)
	}

	if err != nil {
		log.Errorf("Saving certificate file (%s) failed.\nERROR: %s", path, err)
	}

	return err
}
