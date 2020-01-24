This directory is monitored for ATV/ECS configuration files (hot-folder).
As soon as a configuration file is dropped into this folder, the service merges the base configuration with the dropped
configuration and writes the resulting configuration into 'output-merged-configs' and a sdcard update package into
'output-update-packages'.