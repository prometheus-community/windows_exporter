# dns collector

The dns collector exposes metrics about the DNS server


|                                  |                           |
|----------------------------------|---------------------------|
| Metric name prefix               | `bitlocker`               |
| Classes                          | `Win32_EncryptableVolume` |
| Enabled by default               | No                        |
| Metric name prefix (error stats) | `windows_bitlocker`       |

## Flags

None

## Metrics

| Name                                  | Description                                                              | Type  | Labels                        |
|---------------------------------------|--------------------------------------------------------------------------|-------|-------------------------------|
| `windows_bitlocker_volume_info`       | Information about the encryptable volume.                                | gauge | `type`,`volume`,`volume_path` |
| `windows_bitlocker_conversion_status` | Encryption state of the volume.                                          | gauge | `status`,`volume`             |
| `windows_bitlocker_encryption_method` | Algorithm used to encrypt the volume.                                    | gauge | `method`,`volume`             |
| `windows_bitlocker_protection_status` | Status of the volume, whether or not BitLocker is protecting the volume. | gauge | `status`,`volume`             |

### Example metric
```
# HELP windows_bitlocker_conversion_status Encryption state of the volume.
# TYPE windows_bitlocker_conversion_status gauge
windows_bitlocker_conversion_status{status="DECRYPTION IN PROGRESS",volume="C:"} 0
windows_bitlocker_conversion_status{status="DECRYPTION IN PROGRESS",volume="D:"} 0
windows_bitlocker_conversion_status{status="DECRYPTION IN PROGRESS",volume="E:"} 0
windows_bitlocker_conversion_status{status="DECRYPTION PAUSED",volume="C:"} 0
windows_bitlocker_conversion_status{status="DECRYPTION PAUSED",volume="D:"} 0
windows_bitlocker_conversion_status{status="DECRYPTION PAUSED",volume="E:"} 0
windows_bitlocker_conversion_status{status="ENCRYPTION IN PROGRESS",volume="C:"} 0
windows_bitlocker_conversion_status{status="ENCRYPTION IN PROGRESS",volume="D:"} 0
windows_bitlocker_conversion_status{status="ENCRYPTION IN PROGRESS",volume="E:"} 0
windows_bitlocker_conversion_status{status="ENCRYPTION PAUSED",volume="C:"} 0
windows_bitlocker_conversion_status{status="ENCRYPTION PAUSED",volume="D:"} 0
windows_bitlocker_conversion_status{status="ENCRYPTION PAUSED",volume="E:"} 0
windows_bitlocker_conversion_status{status="FULLY DECRYPTED",volume="C:"} 1
windows_bitlocker_conversion_status{status="FULLY DECRYPTED",volume="D:"} 1
windows_bitlocker_conversion_status{status="FULLY DECRYPTED",volume="E:"} 1
windows_bitlocker_conversion_status{status="FULLY ENCRYPTED",volume="C:"} 0
windows_bitlocker_conversion_status{status="FULLY ENCRYPTED",volume="D:"} 0
windows_bitlocker_conversion_status{status="FULLY ENCRYPTED",volume="E:"} 0
# HELP windows_bitlocker_encryption_method Algorithm used to encrypt the volume.
# TYPE windows_bitlocker_encryption_method gauge
windows_bitlocker_encryption_method{method="AES 128",volume="C:"} 0
windows_bitlocker_encryption_method{method="AES 128",volume="D:"} 0
windows_bitlocker_encryption_method{method="AES 128",volume="E:"} 0
windows_bitlocker_encryption_method{method="AES 128 WITH DIFFUSER",volume="C:"} 0
windows_bitlocker_encryption_method{method="AES 128 WITH DIFFUSER",volume="D:"} 0
windows_bitlocker_encryption_method{method="AES 128 WITH DIFFUSER",volume="E:"} 0
windows_bitlocker_encryption_method{method="AES 256",volume="C:"} 0
windows_bitlocker_encryption_method{method="AES 256",volume="D:"} 0
windows_bitlocker_encryption_method{method="AES 256",volume="E:"} 0
windows_bitlocker_encryption_method{method="AES 256 WITH DIFFUSER",volume="C:"} 0
windows_bitlocker_encryption_method{method="AES 256 WITH DIFFUSER",volume="D:"} 0
windows_bitlocker_encryption_method{method="AES 256 WITH DIFFUSER",volume="E:"} 0
windows_bitlocker_encryption_method{method="HARDWARE ENCRYPTION",volume="C:"} 0
windows_bitlocker_encryption_method{method="HARDWARE ENCRYPTION",volume="D:"} 0
windows_bitlocker_encryption_method{method="HARDWARE ENCRYPTION",volume="E:"} 0
windows_bitlocker_encryption_method{method="NOT ENCRYPTED",volume="C:"} 1
windows_bitlocker_encryption_method{method="NOT ENCRYPTED",volume="D:"} 1
windows_bitlocker_encryption_method{method="NOT ENCRYPTED",volume="E:"} 1
windows_bitlocker_encryption_method{method="XTS-AES 128",volume="C:"} 0
windows_bitlocker_encryption_method{method="XTS-AES 128",volume="D:"} 0
windows_bitlocker_encryption_method{method="XTS-AES 128",volume="E:"} 0
windows_bitlocker_encryption_method{method="XTS-AES 256 WITH DIFFUSER",volume="C:"} 0
windows_bitlocker_encryption_method{method="XTS-AES 256 WITH DIFFUSER",volume="D:"} 0
windows_bitlocker_encryption_method{method="XTS-AES 256 WITH DIFFUSER",volume="E:"} 0
# HELP windows_bitlocker_protection_status Status of the volume, whether or not BitLocker is protecting the volume.
# TYPE windows_bitlocker_protection_status gauge
windows_bitlocker_protection_status{status="PROTECTION OFF",volume="C:"} 1
windows_bitlocker_protection_status{status="PROTECTION OFF",volume="D:"} 1
windows_bitlocker_protection_status{status="PROTECTION OFF",volume="E:"} 1
windows_bitlocker_protection_status{status="PROTECTION ON",volume="C:"} 0
windows_bitlocker_protection_status{status="PROTECTION ON",volume="D:"} 0
windows_bitlocker_protection_status{status="PROTECTION ON",volume="E:"} 0
windows_bitlocker_protection_status{status="PROTECTION UNKNOWN",volume="C:"} 0
windows_bitlocker_protection_status{status="PROTECTION UNKNOWN",volume="D:"} 0
windows_bitlocker_protection_status{status="PROTECTION UNKNOWN",volume="E:"} 0
# HELP windows_bitlocker_volume_info Information about the encryptable volume.
# TYPE windows_bitlocker_volume_info counter
windows_bitlocker_volume_info{type="FIXED DISK",volume="D:",volume_path="\\\\?\\Volume{f998d65e-2976-456a-bc1a-87029e40e34f}\\"} 1
windows_bitlocker_volume_info{type="FIXED DISK",volume="E:",volume_path="\\\\?\\Volume{d7410882-1aa3-11f0-97cb-68545ae07765}\\"} 1
windows_bitlocker_volume_info{type="SYSTEM",volume="C:",volume_path="\\\\?\\Volume{eeb47cb6-f390-4b7b-a107-3d70152440a4}\\"} 1
```

## Useful queries
_This collector does not yet have any useful queries added, we would appreciate your help adding them!_

## Alerting examples
_This collector does not yet have alerting examples, we would appreciate your help adding them!_
