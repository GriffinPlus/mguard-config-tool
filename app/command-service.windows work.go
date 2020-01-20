// +build windows

package main

// processFileInHotfolder is called when a new file arrives in the input directory for configuration files.
// It processes .atv/.ecs files, merges them with the common configuration and writes them to the output directory.
func (cmd *ServiceCommand) processFileInHotfolder(path string) error {
	return nil
}
