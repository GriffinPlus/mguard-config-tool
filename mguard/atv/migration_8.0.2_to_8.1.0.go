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
	migration.migration1(newFile)
	newFile.SetVersion(migration.ToVersion())
	return newFile, nil
}

// migration1 replaces VPN_CONNECTION.x.TUNNEL.y.LOCAL_1TO1NAT with VPN_CONNECTION.x.TUNNEL.y.LOCAL_N_TO_N_NAT.
func (migration migration_8_0_2_to_8_1_0) migration1(file *File) error {

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

							var local *string
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

							if local_1to1nat == nil {
								return fmt.Errorf("VPN_CONNECTION.%d.TUNNEL.%d.LOCAL_1TO1NAT is missing", i, j)
							}

							local_ip, local_ipnet, err := net.ParseCIDR(*local)
							if err != nil {
								return err
							}

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

							return nil
						}
					}
				}
			}
		}
	}

	return nil
}
