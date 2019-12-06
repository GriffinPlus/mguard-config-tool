package ecs

import "github.com/griffinplus/mguard-config-tool/shadow"

// DefaultPassFileContent contains the default content of the 'aca/pass' file of an ECS container.
const DefaultPassFileContent = `root\n`

// DefaultSnmpdFileContent contains the default content of the 'aca/snmpd' file of an ECS container.
const DefaultSnmpdFileContent = `createUser "admin" MD5 "SnmpAdmin" DES "SnmpAdmin"\n`

// createDefaultShadowFile creates a new shadow file with default passwords that can be put into the 'aca/users'
// file of an ECS container.
func createDefaultShadowFile() *shadow.File {

	file := shadow.NewFile()
	file.AddUser("root", "root")
	file.AddUser("admin", "mGuard")
	file.AddUser("user", "")
	file.AddUser("netadmin", "")
	file.AddUser("audit", "")
	file.AddUser("userfwd", "")
	return file
}
