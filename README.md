# mguard-config-tool

[![Build Status](https://dev.azure.com/griffinplus/mGuard-Config-Tool/_apis/build/status/mguard-config-tool?branchName=master)](https://dev.azure.com/griffinplus/mGuard-Config-Tool/_build/latest?definitionId=16&branchName=master)
[![Release](https://img.shields.io/github/release/griffinplus/mguard-config-tool.svg?logo=github&label=Release)](https://github.com/GriffinPlus/mguard-config-tool/releases)

-----

## Overview and Motivation

The *mGuard* security router series is a family of firewall/router devices that is produced by the PHOENIX CONTACT
Cyber Security GmbH, a member of the [PHOENIX CONTACT group](https://www.phoenixcontact.com/). An *mGuard* in connection
with the [mGuard Secure Cloud](https://us.cloud.mguard.com/) can be used for remote servicing machines in the field
via IPSec-VPN.

Setting up an *mGuard* in the *mGuard Secure Cloud* generates a configuration file containing everything that is needed
to connect to the cloud. The configuration file can then be loaded into the *mGuard* to set it up properly. As a
configuration file is a snapshot of all settings, loading a configuration file replaces existing settings in the
*mGuard*. Therefore it is not possible to have custom settings along with the configuration file generated by the
cloud. A tool that is able to merge two configurations into one would be handy. This is where the *mGuard-Config-Tool*
comes into play.

The *mGuard-Config-Tool* aims to ease handling *mGuard* configuration files. It's main features are:

- Support for ATV files and ECS containers (unencrypted + encrypted)
- Tasks
  - User Management: Add users and set/verify passwords
  - Conditioning: Condition a configuration and convert formats (ATV <=> ECS)
  - Merging: Merge two configurations into one
- On Windows: Service for merging configurations and creating update packages

## Prerequisites

Writing encrypted ECS containers requires the *openssl* application to be installed. The directory containing the *openssl*
executable should be in the search path.

### Linux

On linux systems the *openssl* can be installed using the system's package manager. Usually it is already in the search path.

### Windows

On Windows systems the *openssl* can be downloaded [here](http://slproweb.com/products/Win32OpenSSL.html).
The search path must be adjusted manually on your system to point to the appropriate directory (most likely:
`C:\Program Files\OpenSSL-Win64\bin`). If you use the service mode only, you can also explicitly specify the path of the
*openssl* executable in the service configuration (see below).

## Releases

*mGuard-Config-Tool* is written in GO which makes it highly portable.

[Downloads](https://github.com/GriffinPlus/mguard-config-tool/releases) are provided for the following combinations of
popular target operating systems and platforms:

- Linux
  - Intel x86 Platform (`386`)
  - Intel x64 Platform (`amd64`)
  - ARM 32-bit Platform (`arm`)
  - ARM 64-bit Platform (`arm64`)
- Windows
  - Intel x86 Platform (`386`)
  - Intel x64 Platform (`amd64`)

If any other target operating system and/or platform is needed and the combination is supported by GO, please open an
issue and we'll add support for it.

## Usage

The *mGuard-Config-Tool* is a console application that uses subcommands to run different tasks. The application writes
log messages to *stderr*, so it's easy to pipe them into a log file or discard them entirely. The *stdout* channel is
used for emitting content only.

### Subcommand: user

The `user` subcommand provides access to the user management. All operations run on ECS files only, but support an
implicit conversion if an ATV file is specified. ATV files do not contain information about user accounts and
passwords. By default the ECS container to work on is expected to be passed via *stdin* to ease scripting without
generating temporary files. The output of these operations is an ECS container that is written to *stdout*.
Optionally input and output can be regular files as well by specifying `--ecs-in` and `--ecs-out` appropriately.

```
user - Add user and set/verify user passwords (ECS containers only).

  Usage:
        user [add|password]

  Subcommands:
    add        Add a user.
    password   Set or verify the password of a user.

  Flags:
       --version   Displays the program version string.
    -h --help      Displays help with available flag, subcommand, and positional value parameters.
       --verbose   Include additional messages that might help when problems occur.
```

The subcommand `user add` allows to add new users:

```
add - Add a user.

  Usage:
        add [username] [password]

  Positional Variables:
        username   Login name of the user. (Required)
        password   Password of the user. (Required)

  Flags:
       --version   Displays the program version string.
    -h --help      Displays help with available flag, subcommand, and positional value parameters.
       --ecs-in    The ECS container (instead of stdin).
       --ecs-out   File receiving the updated ECS container (instead of stdout).
       --verbose   Include additional messages that might help when problems occur.
```

The subcommand `user password set` sets the password of a user.

```
set - Set the password a user (ECS containers only).

  Usage:
        set [username] [password]

  Positional Variables:
        username   Login name of the user. (Required)
        password   Password of the user. (Required)

  Flags:
       --version   Displays the program version string.
    -h --help      Displays help with available flag, subcommand, and positional value parameters.
       --ecs-in    The ECS container (instead of stdin).
       --ecs-out   File receiving the updated ECS container (instead of stdout).
       --verbose   Include additional messages that might help when problems occur.
```

The subcommand `user password verify` verifys the password of a user. Depending on the outcome of the verification the
exit code can be one of the following:

- 0 : The specified password is correct.
- 1 : The specified password is not correct.
- any other : An error occurred.

```
verify - Verify the password of a user (ECS containers only).

  Usage:
        verify [username] [password]

  Positional Variables:
        username   Login name of the user. (Required)
        password   Password of the user. (Required)

  Flags:
       --version   Displays the program version string.
    -h --help      Displays help with available flag, subcommand, and positional value parameters.
       --ecs-in    The ECS container (instead of stdin).
       --verbose   Include additional messages that might help when problems occur.
```

### Subcommand: condition

The `condition` subcommand provides access to conditioning and conversion. *Conditioning* takes a configuration file,
parses the configuration and writes the configuration properly formatted. *Conversion* actually conditions a
configuration and writes it using a different format (ATV or ECS). This way an ATV file that was originally created for
use with the mGuard web interface can be converted to an ECS file that can be used to configure an mGuard via SDCard
and vice versa.

By default the ECS container to work on is expected to be passed via *stdin* to ease scripting without generating
temporary files. The output of the operation is an ECS container that is written to *stdout*. Optionally input and
output can be regular files as well by specifying `--ecs-in`, `--ecs-out` and `--atv-out` appropriately.

If an ATV file is passed in and an ECS container is written the conditioned ATV file is stored in the ECS container and
the missing parts in the ECS container are initialized with defaults. The defaults are the same a new mGuard comes with.
If an ECS container is passwd in and an ATV file is written, the configuration stored in the ECS container is simply
extracted and saved as an ATV file.

```
condition - Condition and/or convert a mGuard configuration file

  Flags:
       --version   Displays the program version string.
    -h --help      Displays help with available flag, subcommand, and positional value parameters.
       --in        File containing the mGuard configuration to condition (ATV format or ECS container)
       --atv-out   File receiving the conditioned configuration (ATV format, instead of stdout)
       --ecs-out   File receiving the conditioned configuration (ECS container, instead of stdout)
       --verbose   Include additional messages that might help when problems occur.
```

### Subcommand: merge

The `merge` subcommand merges two configuration files into one. The first specified file is taken as the base for the
result, then the configuration stored in the second file is "stacked" upon the first configuration. Simple values are
overwritten. Table values (lists) are merged depending on their row id. Rows with the same row id are overwritten, other
rows are appended to the table. If the first file is an ATV file, it is implicitly converted to an ECS file first.
Therefore the resulting ECS container contains default settings for all elements except the encorporated mGuard
configuration (`/aca/cfg`).

If the second configuration file has a lower version than the first one, the *mGuard-Config-Tool* trys to migrate the
second configuration up to the version of the first configuration file. If the second configuration file has a higher
version than the first configuration file, merging will be aborted. The migrations are implemented in all conscience,
but they are not complete. The merged configuration must be tested thoroughly before bringing it in production. Please
also see the known limitations below.

By default all settings are merged from the second configuration into the first configuration. Optionally you can merge
selectively by specifying a merge configuration using `--config`. The merge configuration is just a list of settings
that should be merged. At present only top-level settings are supported. Everything behind a `#` character is treated
as a comment ([example](./app/data/configs/mguard-secure-cloud.merge)).

By default the output of the operation is an ECS container that is written to *stdout*. The output can be written to
a regular file as well by specifying `--ecs-out` and `--atv-out` appropriately.

```
merge - Merge two mGuard configuration files into one

  Usage:
	merge [1st-file] [2nd-file]

  Positional Variables: 
	1st-file   First configuration file to merge (Required)
	2nd-file   Second configuration file to merge (Required)

  Flags: 
       --version   Displays the program version string.
    -h --help      Displays help with available flag, subcommand, and positional value parameters.
       --config    Merge configuration file
       --atv-out   File receiving the merged configuration (ATV format, instead of stdout)
       --ecs-out   File receiving the merged configuration (ECS container, instead of stdout)
       --verbose   Include additional messages that might help when problems occur.
```

### Subcommand: service \*\***WINDOWS ONLY**\*\*

The `service` subcommand provides access to the *Configuration Preparation Service* (CPS). The CPS is part of the
*mGuard-Config-Tool*, stays in the background and monitors a specific directory for ATV/ECS files (hot-folder technique).
As soon as a new ATV/ECS file is dropped into the hot-folder, the service merges a specific mGuard base configuration with
the dropped configuration and generates ATV/ECS files with the merged configuration. Furthermore the service generates a
zip file containing everything that is needed to flash a specific firmware and to load an initial configuration into a mGuard
device. This is particularly useful when preparing mGuards in production and allows to run (and update) the *mGuard-Config-Tool*
on a server. By default the service is registered to run as `LocalSystem` which means that it runs with administrative rights
on the machine it is installed on. This ensures that you do not run into permission issues when running it locally, but you
are strongly encouraged to create a service account with the minimum rights the service needs to access the configured
directories. This step is also required when working with shared folders on a network as `LocalSystem` is often not allowed
to access file shares.

The name of the ATV/ECS files dropped into the hot-folder must match a specific pattern to get processed. The pattern depends
on whether the service has to generate encrypted ECS containers. If the service generates unencrypted ECS containers or no
ECS containers at all, the pattern is `*.(atv|ecs|tgz)`. If the service is configured to generate encrypted ECS containers
the pattern includes the serial number of the mGuard and is `<serial>.(atv|ecs|tgz)`. The serial number uniquely identifies
mGuard devices and allows to access the service to retrieve the appropriate device certificate that is needed to encrypt the
generated ECS files.

The `install` and `uninstall` subcommand installs respectively uninstalls the *mGuard-Config-Tool* as a windows service.
The `start` and `stop` subcommands communicate with the Service Control Manager (SCM) to start/stop the installed service.
As for all services, the usual service control panel (`services.msc`) can also be used. The `debug` subcommand runs the
service without installation and is used for debugging purposes only.

```
service - Controls the mGuard Configuration Preparation Service (CPS)

  Usage:
	service [install|uninstall|start|stop|debug]

  Subcommands: 
    install     Install the windows service
    uninstall   Uninstall the windows service
    start       Start the installed windows service
    stop        Stop the installed windows service
    debug       Run as a command line application for debugging purposes

  Flags: 
       --version   Displays the program version string.
    -h --help      Displays help with available flag, subcommand, and positional value parameters.
       --verbose   Include additional messages that might help when problems occur.
```

The service can be configured using a YAML configuration file that is expected beside `mguard-config-tool.exe`. It must
be named `mguard-config-tool.yaml`. The default configuration is generated in the first run, if permissions allow that.
The default configuration file looks like the following:

```yaml
cache:
  path: ./data/cache                               # directory: cache for various files (e.g. downloaded certificates)
device_database:                                   # credentials for the mGuard device database (only needed when creating encrypted ECS files)
  user: ""                                         # username to use when authenticating against the mGuard device database
  password: ""                                     # password to use when authenticating against the mGuard device database
input:
  base_configuration:
    path: ./data/configs/default.atv               # file: base configuration (usually an ATV file)
  merge_configuration:
    path: ./data/configs/mguard-secure-cloud.merge # file: merge configuration (empty => merge all settings)
  hotfolder:
    path: ./data/input                             # directory: the hot-folder that is monitored for ATV/ECS files to process
  sdcard_template:
    path: ./data/sdcard-template                   # directory: basic sdcard structure (with firmware files)
output:
  merged_configurations:
    path: ./data/output-merged-configs             # directory: merged configurations are put here
    write_atv: true                                # controls whether to generate an ATV file with the merged configuration (true, false)
    write_unencrypted_ecs: true                    # controls whether to generate an unencrypted ECS file with the merged configuration (true, false)
    write_encrypted_ecs: true                      # controls whether to generate an encrypted ECS file with the merged configuration (true, false)
update_packages:
    path: ./data/output-update-packages            # directory: update packages with firmware and the merged configuration are put here
tools:
  openssl:
    path: ""                                       # file: openssl executable (empty => search the PATH variable for the executable)
```

The configured directories are created, if necessary and permissions allow that. Before the service can run, the base
configuration file and the sdcard template files must be provided in the configured directories. If the base configuration or
the sdcard template files are missing, the service will fail to start.

## Known Limitations

- Comments: Reading a configuration file discards comments, so a written configuration file does not contain any comments.
- Migrations: Only a selection of migrations is implemented to make our own use cases work. The lack of documentation about
  ATV documents and migrations forced us to deduce needed migration steps from observed behavior. If you discover further
  steps that are needed to migrate from one version to another, please let us know by opening an issue.

## Issues and Contributions

As *mGuard-Config-Tool* is **not** a project of PHOENIX CONTACT, please don't request help for it there!

If you encounter problems using *mGuard-Config-Tool*, please file an [issue](https://github.com/GriffinPlus/mguard-config-tool/issues).

If you have an idea on how to improve *mGuard-Config-Tool*, please also file an [issue](https://github.com/GriffinPlus/mguard-config-tool/issues).
Pull requests are also very appreciated. In case of major changes, please open an issue before to discuss the changes.
This helps to coordinate development and avoids wasting time.
