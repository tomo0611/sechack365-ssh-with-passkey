package main

import "time"

type LoginToken struct {
	Token      string    `json:"token"`
	HostName   string    `json:"hostname"`
	UserName   string    `json:"username"`
	IpAddr     string    `json:"ipaddr"`
	DateTime   time.Time `json:"requested_at"`
	YourIpAddr string    `json:"your_ipaddr,omitempty"`
}
