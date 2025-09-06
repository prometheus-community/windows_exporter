// SPDX-License-Identifier: Apache-2.0
//
// Copyright The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sysinfoapi

import (
	"fmt"
)

type OperatingSystemSKU uint32

// https://learn.microsoft.com/en-us/windows/win32/cimwin32prov/win32-operatingsystem;
// https://learn.microsoft.com/en-us/dotnet/api/microsoft.powershell.commands.operatingsystemsku;
// https://github.com/MicrosoftDocs/memdocs/blob/3de59f60ac54a1eb1fa72d1471c62e1455b4c05c/intune/intune-service/fundamentals/filters-device-properties.md?plain=1;;
// https://github.com/JustinBrow/Action1/blob/87b026e5c3be6ef4343c3c73601916d8b4415a25/Set%20Custom%20Attribute%20Operating%20System%20SKU.ps1
func (sku OperatingSystemSKU) String() string {
	switch sku {
	case 0:
		return "Undefined"
	case 1:
		return "Ultimate"
	case 2:
		return "HomeBasic"
	case 3:
		return "HomePremium"
	case 4:
		return "Enterprise"
	case 5:
		return "HomeBasicN"
	case 6:
		return "Business"
	case 7:
		return "StandardServer"
	case 8:
		return "DatacenterServer"
	case 9:
		return "SmallBusinessServer"
	case 10:
		return "EnterpriseServer"
	case 11:
		return "Starter"
	case 12:
		return "DatacenterServerCore"
	case 13:
		return "StandardServerCore"
	case 14:
		return "EnterpriseServerCore"
	case 15:
		return "EnterpriseServerIA64Edition"
	case 16:
		return "BusinessN"
	case 17:
		return "WebServer"
	case 18:
		return "ClusterServerEdition"
	case 19:
		return "HomeServer"
	case 20:
		return "StorageExpressServer"
	case 21:
		return "StorageStandardServer"
	case 22:
		return "StorageWorkgroupServer"
	case 23:
		return "StorageEnterpriseServer"
	case 24:
		return "ServerForSmallBusiness"
	case 25:
		return "SmallBusinessServerPremium"
	case 26:
		return "TBD"
	case 27:
		return "EnterpriseN"
	case 28:
		return "UltimateN"
	case 29:
		return "WebServerCore"
	case 33:
		return "ServerFoundation"
	case 34:
		return "WindowsHomeServer"
	case 36:
		return "StandardServerV"
	case 37:
		return "DatacenterServerV"
	case 38:
		return "EnterpriseServerV"
	case 39:
		return "DatacenterServerCoreV"
	case 40:
		return "StandardServerCoreV"
	case 41:
		return "EnterpriseServerCoreV"
	case 42:
		return "HyperV"
	case 43:
		return "StorageExpressServerCore"
	case 44:
		return "StorageStandardServerCore"
	case 45:
		return "StorageWorkgroupServerCore"
	case 46:
		return "StorageEnterpriseServerCore"
	case 48:
		return "Professional"
	case 49:
		return "BusinessN"
	case 50:
		return "SBSolutionServer"
	case 63:
		return "SmallBusinessServerPremiumCore"
	case 64:
		return "ClusterServerV"
	case 72:
		return "EnterpriseEval"
	case 84:
		return "EnterpriseNEval"
	case 87:
		return "WindowsThinPC"
	case 89:
		return "WindowsEmbeddedIndustry"
	case 97:
		return "CoreARM"
	case 98:
		return "CoreN"
	case 99:
		return "CoreCountrySpecific"
	case 100:
		return "CoreSingleLanguage"
	case 101:
		return "Core"
	case 103:
		return "ProfessionalWMC"
	case 104:
		return "MobileCore"
	case 111:
		return "SKU_111"
	case 118:
		return "WindowsEmbeddedHandheld"
	case 119:
		return "PPIPro"
	case 121:
		return "Education"
	case 122:
		return "EducationN"
	case 123:
		return "IoTUAP"
	case 125:
		return "EnterpriseS"
	case 126:
		return "EnterpriseSN"
	case 129:
		return "EnterpriseSEval"
	case 131:
		return "IoTUAPCommercial"
	case 136:
		return "Holographic"
	case 138:
		return "ProfessionalSingleLanguage"
	case 143:
		return "DatacenterNanoServer"
	case 144:
		return "StandardNanoServer"
	case 147:
		return "DatacenterWSServerCore"
	case 148:
		return "StandardWSServerCore"
	case 161:
		return "ProfessionalWorkstation"
	case 162:
		return "ProfessionalN"
	case 164:
		return "ProfessionalEducation"
	case 165:
		return "ProfessionalEducationN"
	case 171:
		return "EnterpriseG"
	case 172:
		return "EnterpriseGN"
	case 175:
		return "EnterpriseForVirtualDesktops"
	case 188:
		return "IoTEnterprise"
	case 202:
		return "CloudEditionN"
	case 203:
		return "CloudEdition"
	case 407:
		return "DatacenterServerAzureEdition"
	default:
		return fmt.Sprintf("Unknown (0x%X)", uint32(sku))
	}
}
