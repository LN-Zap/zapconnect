package main

import (
	"fmt"
	"net"
	"io/ioutil"
	"github.com/Baozisoftware/qrcode-terminal-go"
	"encoding/json"
	b64 "encoding/base64"
	"net/http"
	"os"
	"strings"
	"encoding/pem"
)

type certificates struct {
	Cert		string `json:"c"`
	Macaroon	string `json:"m"`
	Ip 			string `json:"ip,omitempty"`
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func getPublicIP() string {
	resp, err := http.Get("http://ipv4.myexternalip.com/raw")
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Stderr.WriteString("\n")
		os.Exit(1)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusOK {
	    bodyBytes, _ := ioutil.ReadAll(resp.Body)
	    return strings.TrimSpace(string(bodyBytes))
	}

	return ""
}

func main() {
	loadedConfig, err := loadConfig()
	if err != nil {
		fmt.Println()
		return
	}

	certBytes, err := ioutil.ReadFile(loadedConfig.TLSCertPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	block, _ := pem.Decode(certBytes)
	if block == nil || block.Type != "CERTIFICATE" {
		fmt.Println("failed to decode PEM block containing certificate")
	}

	certificate := b64.StdEncoding.EncodeToString([]byte(block.Bytes))

	macBytes, err := ioutil.ReadFile(loadedConfig.AdminMacPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	macaroonB64 := b64.StdEncoding.EncodeToString([]byte(macBytes))

	ipString := ""
	if loadedConfig.ZapConnect.LocalIp {
		ipString = getLocalIP()
	} else if loadedConfig.ZapConnect.Localhost {
		ipString = "127.0.0.1"
	} else {
		ipString = "" //manually type in the public ip address in "" as a temporary fix 
		//getPublicIP()
	}

	addr := loadedConfig.RPCListeners[0]
	_, port, err := net.SplitHostPort(addr.String())
	if err != nil {
		fmt.Println(err)
		return
	}

	ipString = net.JoinHostPort(ipString, port)

	cert := &certificates{
		Cert:     certificate,
		Macaroon: macaroonB64,
		Ip:       ipString}
	certB, _ := json.Marshal(cert)


	if loadedConfig.ZapConnect.Json {
		fmt.Println(string(certB))
	} else {
		obj := qrcodeTerminal.New()
		obj.Get(string(certB)).Print()
		fmt.Println("\n⚠️  Press \"cmd + -\" a few times to see the full QR Code!\nIf that doesn't work run \"zapconnect -j\" to get a code you can copy paste into the app.")
	}
}
