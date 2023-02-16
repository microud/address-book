package main

import (
	"bytes"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"log"
	"net"
	"time"
)

type Capture struct {
	*pcap.Handle
	db *Database
}

type CaptureParams struct {
	Device         string
	SnapshotLength int32
	Promiscuous    bool
	Timeout        time.Duration
}

func NewCapture(db *Database, params CaptureParams) (*Capture, error) {
	handle, err := pcap.OpenLive(params.Device, params.SnapshotLength, params.Promiscuous, params.Timeout)
	if err != nil {
		return nil, err
	}

	err = handle.SetBPFFilter("(udp and (port 67 or port 68)) or arp or (icmp and (icmp[icmptype] == 8 or icmp[icmptype] == 0))")
	if err != nil {
		return nil, err
	}

	return &Capture{Handle: handle, db: db}, nil
}

func (c *Capture) Run() {
	source := gopacket.NewPacketSource(c.Handle, c.Handle.LinkType())
	for packet := range source.Packets() {
		addresses := c.getAddresses(packet)
		for _, address := range addresses {
			err := c.db.SaveAddress(address)
			if err != nil {
				log.Printf("Save address %s(%s) failed: %s", address.IP, address.MAC, err.Error())
			}
		}
	}
}

func (c *Capture) getAddressFromARPPacket(layer gopacket.Layer) []Address {
	var addresses []Address

	arpPacket := layer.(*layers.ARP)
	log.Printf("Detect address %s(%s) from ARP packet",
		net.IP(arpPacket.SourceProtAddress).String(), net.HardwareAddr(arpPacket.SourceHwAddress).String())
	addresses = append(addresses, Address{
		IP:  net.IP(arpPacket.SourceProtAddress).String(),
		MAC: net.HardwareAddr(arpPacket.SourceHwAddress).String(),
	})

	if arpPacket.Operation == 2 {
		log.Printf("Detect address %s(%s) from ARP reply packet",
			net.IP(arpPacket.DstProtAddress).String(), net.HardwareAddr(arpPacket.DstHwAddress).String())
		addresses = append(addresses, Address{
			IP:  net.IP(arpPacket.DstProtAddress).String(),
			MAC: net.HardwareAddr(arpPacket.DstHwAddress).String(),
		})
	}

	return addresses
}

func getDHCPOption(options layers.DHCPOptions, opt layers.DHCPOpt) *layers.DHCPOption {
	for _, option := range options {
		if option.Type == opt {
			return &option
		}
	}

	return nil
}

func (c *Capture) getAddressFromDHCPPacket(packet gopacket.Packet) []Address {
	dhcpLayer := packet.Layer(layers.LayerTypeDHCPv4)
	if dhcpLayer == nil {
		return nil
	}

	dhcpPacket := dhcpLayer.(*layers.DHCPv4)

	messageTypeOption := getDHCPOption(dhcpPacket.Options, layers.DHCPOptMessageType)
	if messageTypeOption == nil {
		return nil
	}

	if bytes.Compare(messageTypeOption.Data, []byte{0x03}) == 0 {
		requestedIPOption := getDHCPOption(dhcpPacket.Options, layers.DHCPOptRequestIP)
		if requestedIPOption == nil {
			return nil
		}
		clientIdentifierOption := getDHCPOption(dhcpPacket.Options, layers.DHCPOptClientID)
		if clientIdentifierOption == nil {
			return nil
		}
		log.Printf("Detect address %s(%s) from DHCP Offer packet",
			net.IP(requestedIPOption.Data).String(), net.HardwareAddr(clientIdentifierOption.Data).String())
		return []Address{
			{
				IP:  net.IP(requestedIPOption.Data).String(),
				MAC: net.HardwareAddr(clientIdentifierOption.Data).String(),
			},
		}
	}

	if bytes.Compare(messageTypeOption.Data, []byte{0x02}) == 0 {
		log.Printf("Detect address %s(%s) from DHCP Offer packet",
			dhcpPacket.YourClientIP.String(), dhcpPacket.ClientHWAddr.String())
		return []Address{
			{
				IP:  dhcpPacket.YourClientIP.String(),
				MAC: dhcpPacket.ClientHWAddr.String(),
			},
		}
	}

	if bytes.Compare(messageTypeOption.Data, []byte{0x05}) == 0 {
		log.Printf("Detect address %s(%s) from DHCP ACK packet",
			dhcpPacket.YourClientIP.String(), dhcpPacket.ClientHWAddr.String())
		return []Address{
			{
				IP:  dhcpPacket.YourClientIP.String(),
				MAC: dhcpPacket.ClientHWAddr.String(),
			},
		}
	}

	return nil
}

func (c *Capture) getAddresses(packet gopacket.Packet) []Address {
	if arpLayer := packet.Layer(layers.LayerTypeARP); arpLayer != nil {
		return c.getAddressFromARPPacket(arpLayer)
	}

	if dhcpLayer := packet.Layer(layers.LayerTypeDHCPv4); dhcpLayer != nil {
		return c.getAddressFromDHCPPacket(packet)
	}

	return []Address{}
}
