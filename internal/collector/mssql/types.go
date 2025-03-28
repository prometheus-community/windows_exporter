package mssql

import (
	"fmt"
	"strings"

	"golang.org/x/sys/windows/registry"
)

type mssqlInstance struct {
	name            string
	majorVersion    mssqlServerMajorVersion
	patchVersion    string
	edition         string
	isFirstInstance bool
}

func newMssqlInstance(key, name string) (mssqlInstance, error) {
	regKey := fmt.Sprintf(`Software\Microsoft\Microsoft SQL Server\%s\Setup`, name)

	k, err := registry.OpenKey(registry.LOCAL_MACHINE, regKey, registry.QUERY_VALUE)
	if err != nil {
		return mssqlInstance{}, fmt.Errorf("couldn't open registry Software\\Microsoft\\Microsoft SQL Server\\%s\\Setup: %w", name, err)
	}

	defer func(key registry.Key) {
		_ = key.Close()
	}(k)

	patchVersion, _, err := k.GetStringValue("Version")
	if err != nil {
		return mssqlInstance{}, fmt.Errorf("couldn't get version from registry: %w", err)
	}

	edition, _, err := k.GetStringValue("Edition")
	if err != nil {
		return mssqlInstance{}, fmt.Errorf("couldn't get version from registry: %w", err)
	}

	_, name, _ = strings.Cut(name, ".")

	return mssqlInstance{
		edition:         edition,
		name:            name,
		majorVersion:    newMajorVersion(patchVersion),
		patchVersion:    patchVersion,
		isFirstInstance: key == "MSSQLSERVER",
	}, nil
}

func (m mssqlInstance) isVersionGreaterOrEqualThan(version mssqlServerMajorVersion) bool {
	return m.majorVersion.isGreaterOrEqualThan(version)
}

type mssqlServerMajorVersion int

const (
	// https://sqlserverbuilds.blogspot.com/
	serverVersionUnknown mssqlServerMajorVersion = 0
	serverVersion2012    mssqlServerMajorVersion = 11
	serverVersion2014    mssqlServerMajorVersion = 12
	serverVersion2016    mssqlServerMajorVersion = 13
	serverVersion2017    mssqlServerMajorVersion = 14
	serverVersion2019    mssqlServerMajorVersion = 15
	serverVersion2022    mssqlServerMajorVersion = 16
	serverVersion2025    mssqlServerMajorVersion = 17
)

func newMajorVersion(patchVersion string) mssqlServerMajorVersion {
	majorVersion, _, _ := strings.Cut(patchVersion, ".")
	switch majorVersion {
	case "11":
		return serverVersion2012
	case "12":
		return serverVersion2014
	case "13":
		return serverVersion2016
	case "14":
		return serverVersion2017
	case "15":
		return serverVersion2019
	case "16":
		return serverVersion2022
	case "17":
		return serverVersion2025
	default:
		return serverVersionUnknown
	}
}

func (m mssqlServerMajorVersion) String() string {
	switch m {
	case serverVersion2012:
		return "2012"
	case serverVersion2014:
		return "2014"
	case serverVersion2016:
		return "2016"
	case serverVersion2017:
		return "2017"
	case serverVersion2019:
		return "2019"
	case serverVersion2022:
		return "2022"
	case serverVersion2025:
		return "2025"
	default:
		return "unknown"
	}
}

func (m mssqlServerMajorVersion) isGreaterOrEqualThan(version mssqlServerMajorVersion) bool {
	return m >= version
}
