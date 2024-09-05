package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
)

type Config struct {
	InterfaceName string `json:"interfaceName"`
	BearerToken   string `json:"bearerToken"`
	DomainName    string `json:"domainName"`
	DnsRecordId   string `json:"dnsRecordId"`
	ZoneId        string `json:"zoneId"`
}

type RequestBody struct {
	Comment string   `json:"comment"`
	Name    string   `json:"name"`
	Proxied bool     `json:"proxied"`
	Tags    []string `json:"tags"`
	Ttl     int      `json:"ttl"`
	Content string   `json:"content"`
	Type    string   `json:"type"`
}

func getConfig(location string) (*Config, error) {
	data, err := os.ReadFile(location)
	if err != nil {
		return nil, err
	}
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func getLocalIpv6(interfaceName string) (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, i := range interfaces {
		if i.Name == interfaceName {
			addrs, err := i.Addrs()
			if err != nil {
				return "", err
			}
			for _, addr := range addrs {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}

				if ip.To4() == nil && !ip.IsPrivate() {
					return ip.String(), nil
				}
			}
		}
	}
	return "", fmt.Errorf("interface %s not found", interfaceName)
}

func updateCloudFlareRecord(config *Config, content string) error {
	client := &http.Client{}

	body := RequestBody{
		Comment: "Updated by CDR",
		Name:    config.DomainName,
		Proxied: false,
		Tags:    []string{},
		Ttl:     600,
		Content: content,
		Type:    "AAAA",
	}
	jsonBody, err := json.Marshal(body)
	fmt.Println(string(jsonBody))
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", config.ZoneId, config.DnsRecordId), bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.BearerToken))
	req.Header.Add("Content-Type", "application/json")

	esp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer esp.Body.Close()

	bodyBytes, err := io.ReadAll(esp.Body)
	if err != nil {
		return err
	}

	fmt.Println("Response Status:", esp.Status)
	fmt.Println("Response BoDy:", string(bodyBytes))
	return nil
}

func main() {
	config, err := getConfig("config.json")
	if err != nil {
		fmt.Println(err)
		return
	}

	ipv6, err := getLocalIpv6(config.InterfaceName)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = updateCloudFlareRecord(config, ipv6)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Success")
}
