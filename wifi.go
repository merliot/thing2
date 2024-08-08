package thing2

import "strings"

var (
	ssids       = GetEnv("WIFI_SSIDS", "")
	passphrases = GetEnv("WIFI_PASSPHRASES", "")
	wifiAuths   = wifiAuthsInit()
)

type wifiAuthMap map[string]string

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
