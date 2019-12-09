// Package shadow provides functions to read, modify and write a linux shadow file.
package shadow

import (
	"github.com/tredoe/osutil/user/crypt/apr1_crypt"
	"github.com/tredoe/osutil/user/crypt/md5_crypt"
	"github.com/tredoe/osutil/user/crypt/sha256_crypt"
	"github.com/tredoe/osutil/user/crypt/sha512_crypt"
)

func init() {

	// initialize the crypt packages,
	// so determining the hash type from the password string works properly
	apr1_crypt.New()
	md5_crypt.New()
	sha256_crypt.New()
	sha512_crypt.New()
}
