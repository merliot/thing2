package thing2

import (
	"strings"
)

var (
	ssids       = getEnv("WIFI_SSIDS", "")
	passphrases = getEnv("WIFI_PASSPHRASES", "")
	wifiAuths   = wifiAuthsInit()
)

type wifiAuthMap map[string]string //key: ssid; value: passphrase

func wifiAuthsInit() wifiAuthMap {
	auths := make(wifiAuthMap)
	if ssids == "" {
		return auths
	}
	keys := strings.Split(ssids, ",")
	values := strings.Split(passphrases, ",")
	for i, key := range keys {
		if i < len(values) {
			auths[key] = values[i]
		}
	}
	return auths
}
