# govendor notes

    git config --global core.autocrlf false
    go get -u -v

    govendor init
    govendor add +external

    govendor remove github.com/martinlindhe/wmi_exporter/collector
    govendor list

    govendor build +local

    git add vendor
