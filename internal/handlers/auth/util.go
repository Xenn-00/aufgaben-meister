package auth_handlers

import "strings"

func detectDeviceType(ua string) string {
	ua = strings.ToLower(ua)

	switch {
	case strings.Contains(ua, "android"):
		return "Android"
	case strings.Contains(ua, "iphone"):
		return "iPhone"
	case strings.Contains(ua, "ipad"):
		return "iPad"
	case strings.Contains(ua, "windows"):
		return "Windows PC"
	case strings.Contains(ua, "macintosh"):
		return "MacOS"
	case strings.Contains(ua, "linux"):
		return "Linux"
	default:
		return "Unknown Device"
	}
}
