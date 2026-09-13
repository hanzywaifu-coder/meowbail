package protocol

import (
	"go.mau.fi/whatsmeow/proto/waCompanionReg"
	"go.mau.fi/whatsmeow/proto/waWa6"
	"go.mau.fi/whatsmeow/store"
	"google.golang.org/protobuf/proto"
)

// BrowserProfile jenis profil browser WhatsApp Web untuk identifikasi koneksi
type BrowserProfile string

const (
	BrowserChromeUbuntu  BrowserProfile = "Ubuntu Chrome"
	BrowserChromeMac     BrowserProfile = "macOS Chrome"
	BrowserChromeWindows BrowserProfile = "Windows Chrome"
	BrowserSafariMac     BrowserProfile = "macOS Safari"
	BrowserFirefoxLinux  BrowserProfile = "Linux Firefox"
	BrowserDesktop       BrowserProfile = "Desktop"
)

// SetBrowserProfile mengatur identitas OS & platform browser pada handshake WhatsApp
func SetBrowserProfile(profile BrowserProfile) {
	switch profile {
	case BrowserChromeUbuntu:
		store.SetOSInfo("Ubuntu", [3]uint32{22, 4, 4})
		store.DeviceProps.PlatformType = waCompanionReg.DeviceProps_CHROME.Enum()
	case BrowserChromeMac:
		store.SetOSInfo("Mac OS", [3]uint32{14, 4, 1})
		store.DeviceProps.PlatformType = waCompanionReg.DeviceProps_CHROME.Enum()
	case BrowserChromeWindows:
		store.SetOSInfo("Windows", [3]uint32{10, 0, 22631})
		store.DeviceProps.PlatformType = waCompanionReg.DeviceProps_CHROME.Enum()
	case BrowserSafariMac:
		store.SetOSInfo("Mac OS", [3]uint32{14, 4, 1})
		store.DeviceProps.PlatformType = waCompanionReg.DeviceProps_SAFARI.Enum()
	case BrowserFirefoxLinux:
		store.SetOSInfo("Ubuntu", [3]uint32{22, 4, 4})
		store.DeviceProps.PlatformType = waCompanionReg.DeviceProps_FIREFOX.Enum()
	case BrowserDesktop:
		store.SetOSInfo("Windows", [3]uint32{10, 0, 22631})
		store.DeviceProps.PlatformType = waCompanionReg.DeviceProps_DESKTOP.Enum()
		// Gunakan WIN_HYBRID untuk Desktop modern agar tidak ditolak server (428 error)
		if store.BaseClientPayload.WebInfo != nil {
			store.BaseClientPayload.WebInfo.WebSubPlatform = waWa6.ClientPayload_WebInfo_WIN_HYBRID.Enum()
		}
	default:
		store.SetOSInfo("Ubuntu", [3]uint32{22, 4, 4})
		store.DeviceProps.PlatformType = waCompanionReg.DeviceProps_CHROME.Enum()
	}
	// Dukungan message association (album / thread grouping)
	if store.DeviceProps.HistorySyncConfig != nil {
		store.DeviceProps.HistorySyncConfig.SupportMessageAssociation = proto.Bool(true)
	}
}
