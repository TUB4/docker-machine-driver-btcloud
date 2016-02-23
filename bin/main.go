package main

import (
    "github.com/tub4/docker-machine-driver-btcloud"
    "github.com/docker/machine/libmachine/drivers/plugin"
)

func main() {
    plugin.RegisterDriver(btcloud.NewDriver("", ""))
}
