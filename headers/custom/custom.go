package custom

import (
	"log"
	"os"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// GetSizeStoredInPagingFiles returns the total size of paging files across all discs.
func GetSizeStoredInPagingFiles() (int64, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Control\Session Manager\Memory Management`, registry.QUERY_VALUE)
	if err != nil {
		return 0, err
	}
	pagingFiles, _, err := k.GetStringsValue("ExistingPageFiles")
	if err != nil {
		return 0, err
	}

	var size int64 = 0
	for _, pagingFile := range pagingFiles {
		fileString := strings.ReplaceAll(pagingFile, `\??\`, "")
		file, err := os.Stat(fileString)
		if err != nil {
			return 0, err
		}
		size += file.Size()
	}
	return size, nil
}

// GetProductDetails returns the ProductName and CurrentBuildNumber values from the registry.
func GetProductDetails() (string, string) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		log.Fatal(err)
	}
	pn, _, err := k.GetStringValue("ProductName")
	if err != nil {
		log.Fatal(err)
	}

	bn, _, err := k.GetStringValue("CurrentBuildNumber")
	if err != nil {
		log.Fatal(err)
	}

	if err := k.Close(); err != nil {
		log.Fatal(err)
	}
	return pn, bn
}
