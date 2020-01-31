#!/bin/sh

# ***WARNING***
# Use unix line breaks (LF) only, not windows line breaks (CR LF).
# Everything other than LF will render the script unexecutable.

# abort execution on the first error
set -e

# load initial configuration
gaiconfig --silent --set-all < /bootstrap/preconfig.atv

# save the current configuration into a profile, so users can restore it
gaiconfig --get-all > "/gai/profiles/Custom Factory Defaults.atv"

# set passwords
ROOT_PASSWORD="root"
ADMIN_PASSWORD="mGuard"
echo -e "$ROOT_PASSWORD\n$ROOT_PASSWORD" | passwd root
echo -e "$ADMIN_PASSWORD\n$ADMIN_PASSWORD" | passwd admin
