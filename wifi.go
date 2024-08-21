package thing2

import (
	"strings"
)

type wifiAuthMap map[string]string //key: ssid; value: passphrase

func wifiAuths() wifiAuthMap {
	var ssids = getenv("WIFI_SSIDS", "")
	var passphrases = getenv("WIFI_PASSPHRASES", "")

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
