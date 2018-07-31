// common type/functions for various mssql related collectors.

package collector

import (
	"github.com/prometheus/common/log"
	"golang.org/x/sys/windows/registry"
)

type sqlInstancesType map[string]string

func getMSSQLInstances() sqlInstancesType {
	sqlInstances := make(sqlInstancesType)

	// in case querying the registry fails, initialize list to the default instance
	sqlInstances["MSSQLSERVER"] = ""

	regkey := `Software\Microsoft\Microsoft SQL Server\Instance Names\SQL`
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, regkey, registry.QUERY_VALUE)
	if err != nil {
		log.Warn("Couldn't open registry to determine SQL instances:", err)
		return sqlInstances
	}
	defer k.Close()

	instanceNames, err := k.ReadValueNames(0)
	if err != nil {
		log.Warn("Can't ReadSubKeyNames %#v", err)
		return sqlInstances
	}

	for _, instanceName := range instanceNames {
		if instanceVersion, _, err := k.GetStringValue(instanceName); err == nil {
			sqlInstances[instanceName] = instanceVersion
		}
	}

	log.Debugf("Detected MSSQL Instances: %#v\n", sqlInstances)

	return sqlInstances
}
