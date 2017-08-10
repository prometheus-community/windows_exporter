# S.M.A.R.T. collector

Collects [S.M.A.R.T.](https://en.wikipedia.org/wiki/S.M.A.R.T.) hard disk data about the health of your disks.


### Usage

To enable:

    .\wmi_exporter.exe "-collectors.enabled" "smart"



### Troubleshooting

In order to function, vendor-specific drivers for your disk may need to be installed.

The user running wmi_exporter.exe needs the proper privileges to read the root\wmi namespace.

To see what disks has S.M.A.R.T. available on your machine, you can run
    
    gwmi -namespace root\wmi -class MSStorageDriver_ATAPISmartData


### Disappearing metrics

If you experience gaps in the collected smart metrics, most likely Windows put the hard disk to sleep.

To prevent the Hard Disk from going to sleep, click on the Battery / Power icon in the taskbar and select More Power options. In the Control Panel windows which opens, select Change Plan settings for your current Power Plan. In the next window, select Change advanced power settings.

In the Power Options box that opens, click the + sign next to the Hard Disk option. Here you will see the required settings under Turn off hard disk after heading. Change the value to 0.

Click on Apply > OK and exit. This setting will prevent your hard disk from entering the Sleep mode.

[source](http://www.thewindowsclub.com/prevent-hard-drive-going-sleep-windows)
