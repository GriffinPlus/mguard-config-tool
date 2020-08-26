package atv

import (
	"fmt"
	"net"
)

type migration_8_0_2_to_8_1_0 struct{}

// FromVersion returns the document version the migration start with.
func (_ migration_8_0_2_to_8_1_0) FromVersion() Version {
	return Version{
		Major:  8,
		Minor:  0,
		Patch:  2,
		Suffix: "default",
	}
}

// ToVersion returns the document version the migration ends with.
func (_ migration_8_0_2_to_8_1_0) ToVersion() Version {
	return Version{
		Major:  8,
		Minor:  1,
		Patch:  0,
		Suffix: "default",
	}
}

// Migrate performs the migration.
func (migration migration_8_0_2_to_8_1_0) Migrate(file *File) (*File, error) {

	newFile := file.Dupe()

	// VPN_CONNECTION.x.VPN_ENABLED => VPN_CONNECTION.x.VPN_START
	err := migration.migration1(newFile)
	if err != nil {
		return nil, err
	}

	// VPN_CONNECTION.x.FW_INCOMING.y.TARGET => VPN_CONNECTION.x.FW_INCOMING.y.TARGET_REF
	// VPN_CONNECTION.x.FW_OUTGOING.y.TARGET => VPN_CONNECTION.x.FW_OUTGOING.y.TARGET_REF
	err = migration.migration2(newFile)
	if err != nil {
		return nil, err
	}

	// VPN_CONNECTION.x.TUNNEL.y.LOCAL_1TO1NAT => VPN_CONNECTION.x.TUNNEL.y.LOCAL_N_TO_N_NAT
	err = migration.migration3(newFile)
	if err != nil {
		return nil, err
	}

	// VPN_EXTERNAL_SWITCH_REF + VPN_RS_EXTERNAL_SWITCH_TYPE
	// => with VPN_CONNECTION.x.CONTROL + VPN_CONNECTION.x.CONTROL_INV
	err = migration.migration4(newFile)
	if err != nil {
		return nil, err
	}

	newFile.SetVersion(migration.ToVersion())
	return newFile, nil
}

// migration1 replaces VPN_CONNECTION.x.VPN_ENABLED with VPN_CONNECTION.x.VPN_START.
func (migration migration_8_0_2_to_8_1_0) migration1(file *File) error {

	vpnConnection, err := file.doc.GetSetting("VPN_CONNECTION")
	if err != nil {
		return err
	}

	if vpnConnection != nil {
		if vpnConnection.TableValue != nil {
			for i, _ := range vpnConnection.TableValue.Rows {

				// get setting to migrate (VPN_CONNECTION.x.VPN_ENABLED)
				vpnEnabledPath := fmt.Sprintf("VPN_CONNECTION.%d.VPN_ENABLED", i)
				vpnEnabledSetting, err := file.doc.GetSetting(vpnEnabledPath)
				if err != nil {
					return err
				}

				// abort, if there the setting does not exist
				if vpnEnabledSetting == nil {
					return nil
				}

				// get value of setting to migrate
				vpnEnabledValue, err := vpnEnabledSetting.GetValue()
				if err != nil {
					return err
				}

				// add migrated setting
				vpnStartPath := fmt.Sprintf("VPN_CONNECTION.%d.VPN_START", i)
				if vpnEnabledValue == "yes" {
					file.doc.SetSimpleValueSetting(vpnStartPath, "started")
				} else if vpnEnabledValue == "no" {
					file.doc.SetSimpleValueSetting(vpnStartPath, "stopped")
				} else {
					return fmt.Errorf("VPN_ENABLED is neither 'yes' nor 'no'")
				}

				// remove obsolete setting
				file.doc.RemoveSetting(vpnEnabledPath)
				return nil
			}
		}
	}

	return nil
}

// migration2 replaces VPN_CONNECTION.x.FW_INCOMING.y.TARGET with VPN_CONNECTION.x.FW_INCOMING.y.TARGET_REF and
// VPN_CONNECTION.x.FW_OUTGOING.y.TARGET with VPN_CONNECTION.x.FW_OUTGOING.y.TARGET_REF
func (migration migration_8_0_2_to_8_1_0) migration2(file *File) error {

	vpnConnection, err := file.doc.GetSetting("VPN_CONNECTION")
	if err != nil {
		return err
	}

	if vpnConnection != nil {
		if vpnConnection.TableValue != nil {
			for i, vpnConnectionRow := range vpnConnection.TableValue.Rows {
				for _, item := range vpnConnectionRow.Items {

					if item.Name == "FW_INCOMING" && item.TableValue != nil {

						// VPN_CONNECTION.x.FW_INCOMING.y.TARGET
						for j, _ := range item.TableValue.Rows {

							// try to get the setting
							inTargetPath := fmt.Sprintf("VPN_CONNECTION.%d.FW_INCOMING.%d.TARGET", i, j)
							inTargetSetting, err := file.doc.GetSetting(inTargetPath)
							if err != nil {
								return err
							}

							// abort, if the setting does not exist
							if inTargetSetting == nil {
								break
							}

							// change the setting's name (the new version supports the same values + a rowref)
							inTargetSetting.ClearValue()
							inTargetSetting.Name = "TARGET_REF"

							break
						}

					} else if item.Name == "FW_OUTGOING" && item.TableValue != nil {

						// VPN_CONNECTION.x.FW_OUTGOING.y.TARGET
						for j, _ := range item.TableValue.Rows {

							// try to get the setting
							outTargetPath := fmt.Sprintf("VPN_CONNECTION.%d.FW_OUTGOING.%d.TARGET", i, j)
							outTargetSetting, err := file.doc.GetSetting(outTargetPath)
							if err != nil {
								return err
							}

							// abort, if the setting does not exist
							if outTargetSetting == nil {
								break
							}

							// change the setting's name (the new version supports the same values + a rowref)
							outTargetSetting.ClearValue()
							outTargetSetting.Name = "TARGET_REF"

							break
						}
					}

				}
			}
		}
	}

	return nil
}

// migration3 replaces VPN_CONNECTION.x.TUNNEL.y.LOCAL_1TO1NAT with VPN_CONNECTION.x.TUNNEL.y.LOCAL_N_TO_N_NAT.
func (migration migration_8_0_2_to_8_1_0) migration3(file *File) error {

	vpnConnection, err := file.doc.GetSetting("VPN_CONNECTION")
	if err != nil {
		return err
	}

	if vpnConnection != nil {
		if vpnConnection.TableValue != nil {
			for i, vpnConnectionRow := range vpnConnection.TableValue.Rows {
				for _, item := range vpnConnectionRow.Items {
					if item.Name == "TUNNEL" && item.TableValue != nil {
						for j, tunnelRow := range item.TableValue.Rows {

							none := "none"
							var local *string = &none
							var local_1to1nat *string
							for _, item := range tunnelRow.Items {
								switch item.Name {
								case "LOCAL":
									s, _ := item.GetValue()
									local = &s
								case "LOCAL_1TO1NAT":
									s, _ := item.GetValue()
									local_1to1nat = &s
								}
							}

							if local == nil {
								return fmt.Errorf("VPN_CONNECTION.%d.TUNNEL.%d.LOCAL is missing", i, j)
							}

							local_ip, local_ipnet, err := net.ParseCIDR(*local)
							if err != nil {
								return err
							}

							// migrate VPN_CONNECTION.x.TUNNEL.y.LOCAL_1TO1NAT
							if local_1to1nat != nil {

								// parse the ip address
								local_1to1nat_ip := net.ParseIP(*local_1to1nat)
								if local_1to1nat_ip == nil {
									return fmt.Errorf("%s is not a valid IP address", *local_1to1nat)
								}

								// create replacement
								maskbits, _ := local_ipnet.Mask.Size()
								newValue := documentTableValue{
									Rows: []*documentTableRow{
										&documentTableRow{
											Items: []*documentSetting{
												&documentSetting{Name: "COMMENT", SimpleValue: &documentSimpleValue{Value: ""}},
												&documentSetting{Name: "FROM_NET", SimpleValue: &documentSimpleValue{Value: local_1to1nat_ip.String()}},
												&documentSetting{Name: "MASK", SimpleValue: &documentSimpleValue{Value: fmt.Sprintf("%d", maskbits)}},
												&documentSetting{Name: "TO_NET", SimpleValue: &documentSimpleValue{Value: local_ip.String()}},
											},
										},
									},
								}

								// replace LOCAL_1TO1NAT with LOCAL_N_TO_N_NAT
								for i, item := range tunnelRow.Items {
									if item.Name == "LOCAL_1TO1NAT" {
										tunnelRow.Items[i] = &documentSetting{
											Name:       "LOCAL_N_TO_N_NAT",
											TableValue: &newValue,
										}
										break
									}
								}
							}

							return nil
						}
					}
				}
			}
		}
	}

	return nil
}

// migration4 replaces VPN_EXTERNAL_SWITCH_REF and VPN_RS_EXTERNAL_SWITCH_TYPE with VPN_CONNECTION.x.CONTROL and VPN_CONNECTION.x.CONTROL_INV
func (migration migration_8_0_2_to_8_1_0) migration4(file *File) error {

	// get setting VPN_EXTERNAL_SWITCH_REF that indicates which vpn connection is controlled by the button/switch
	vpnExternalSwitchRef, err := file.doc.GetAttribute("VPN_EXTERNAL_SWITCH_REF", "rowref")
	if err != nil {
		return err
	}

	// get setting VPN_RS_EXTERNAL_SWITCH_TYPE that indicates whether a button or a switch is connected to the input
	vpnExternalSwitchType, err := file.doc.GetSetting("VPN_RS_EXTERNAL_SWITCH_TYPE")
	if err != nil {
		return err
	}

	// abort, if there is no vpn connection linked to the button/switch
	// (firmware version supports only one button/switch)
	if vpnExternalSwitchRef == nil {
		return nil
	}

	// determine the type (button or switch) of the input
	// (fall back to 'button', if not specified)
	vpnExternalSwitchTypeValue := "button"
	if vpnExternalSwitchType != nil {
		vpnExternalSwitchTypeValue, err = vpnExternalSwitchType.GetValue()
		if err != nil {
			return err
		}
	}

	// perform the actual migration
	vpnConnection, err := file.doc.GetSetting("VPN_CONNECTION")
	if err != nil {
		return err
	}
	if vpnConnection != nil {
		if vpnConnection.TableValue != nil {
			for i, vpnConnectionRow := range vpnConnection.TableValue.Rows {
				if vpnConnectionRow.RowID != nil && string(*vpnConnectionRow.RowID) == *vpnExternalSwitchRef {

					// set new settings (VPN_CONNECTION.x.CONTROL and VPN_CONNECTION.x.CONTROL_INV)
					pathPrefix := fmt.Sprintf("VPN_CONNECTION.%d.", i)

					// the firmware version supports only one input, so 'cmd1' is always right
					err = file.doc.SetSimpleValueSetting(pathPrefix+"CONTROL", "cmd1")
					if err != nil {
						return err
					}

					// the firmware version does not support inverting the input, so 'no' should be also always be correct
					file.doc.SetSimpleValueSetting(pathPrefix+"CONTROL_INV", "no")
					if err != nil {
						return err
					}

					// set the switch type (SERVICE_SWITCH1_TYPE)
					serviceSwitch1TypeSetting := &documentSetting{
						Name:        "SERVICE_SWITCH1_TYPE",
						SimpleValue: &documentSimpleValue{Value: vpnExternalSwitchTypeValue}}
					err := file.doc.SetSetting(serviceSwitch1TypeSetting)
					if err != nil {
						return err
					}

					// remove deprecated setting VPN_EXTERNAL_SWITCH_REF
					err = file.doc.RemoveSetting("VPN_EXTERNAL_SWITCH_REF")
					if err != nil {
						return err
					}

					// remove deprecated setting VPN_RS_EXTERNAL_SWITCH_TYPE
					err = file.doc.RemoveSetting("VPN_RS_EXTERNAL_SWITCH_TYPE")
					if err != nil {
						return err
					}

					return nil
				}
			}
		}
	}

	return fmt.Errorf("The document does not contain a VPN_CONNECTION with rowid %s", *vpnExternalSwitchRef)
}
