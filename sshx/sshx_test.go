// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sshx

import (
	"encoding/hex"
	"fmt"
	"log"
	"testing"

	"github.com/dreadl0ck/gopacket"
	"github.com/dreadl0ck/gopacket/pcap"
)

var (
	clientKexInitData, _ = hex.DecodeString("14d075a1eee91d5fa5d53fd802727542e60000007e6469666669652d68656c6c6d616e2d67726f75702d65786368616e67652d7368613235362c6469666669652d68656c6c6d616e2d67726f75702d65786368616e67652d736861312c6469666669652d68656c6c6d616e2d67726f757031342d736861312c6469666669652d68656c6c6d616e2d67726f7570312d73686131000000837373682d7273612d636572742d763031406f70656e7373682e636f6d2c7373682d7273612d636572742d763030406f70656e7373682e636f6d2c7373682d7273612c7373682d6473732d636572742d763031406f70656e7373682e636f6d2c7373682d6473732d636572742d763030406f70656e7373682e636f6d2c7373682d647373000000cb6165733132382d6374722c6165733139322d6374722c6165733235362d6374722c617263666f75723235362c617263666f75723132382c6165733132382d67636d406f70656e7373682e636f6d2c6165733235362d67636d406f70656e7373682e636f6d2c6165733132382d6362632c336465732d6362632c626c6f77666973682d6362632c636173743132382d6362632c6165733139322d6362632c6165733235362d6362632c617263666f75722c72696a6e6461656c2d636263406c797361746f722e6c69752e7365000000cb6165733132382d6374722c6165733139322d6374722c6165733235362d6374722c617263666f75723235362c617263666f75723132382c6165733132382d67636d406f70656e7373682e636f6d2c6165733235362d67636d406f70656e7373682e636f6d2c6165733132382d6362632c336465732d6362632c626c6f77666973682d6362632c636173743132382d6362632c6165733139322d6362632c6165733235362d6362632c617263666f75722c72696a6e6461656c2d636263406c797361746f722e6c69752e736500000192686d61632d6d64352d65746d406f70656e7373682e636f6d2c686d61632d736861312d65746d406f70656e7373682e636f6d2c756d61632d36342d65746d406f70656e7373682e636f6d2c756d61632d3132382d65746d406f70656e7373682e636f6d2c686d61632d736861322d3235362d65746d406f70656e7373682e636f6d2c686d61632d736861322d3531322d65746d406f70656e7373682e636f6d2c686d61632d726970656d643136302d65746d406f70656e7373682e636f6d2c686d61632d736861312d39362d65746d406f70656e7373682e636f6d2c686d61632d6d64352d39362d65746d406f70656e7373682e636f6d2c686d61632d6d64352c686d61632d736861312c756d61632d3634406f70656e7373682e636f6d2c756d61632d313238406f70656e7373682e636f6d2c686d61632d736861322d3235362c686d61632d736861322d3531322c686d61632d726970656d643136302c686d61632d726970656d64313630406f70656e7373682e636f6d2c686d61632d736861312d39362c686d61632d6d64352d393600000192686d61632d6d64352d65746d406f70656e7373682e636f6d2c686d61632d736861312d65746d406f70656e7373682e636f6d2c756d61632d36342d65746d406f70656e7373682e636f6d2c756d61632d3132382d65746d406f70656e7373682e636f6d2c686d61632d736861322d3235362d65746d406f70656e7373682e636f6d2c686d61632d736861322d3531322d65746d406f70656e7373682e636f6d2c686d61632d726970656d643136302d65746d406f70656e7373682e636f6d2c686d61632d736861312d39362d65746d406f70656e7373682e636f6d2c686d61632d6d64352d39362d65746d406f70656e7373682e636f6d2c686d61632d6d64352c686d61632d736861312c756d61632d3634406f70656e7373682e636f6d2c756d61632d313238406f70656e7373682e636f6d2c686d61632d736861322d3235362c686d61632d736861322d3531322c686d61632d726970656d643136302c686d61632d726970656d64313630406f70656e7373682e636f6d2c686d61632d736861312d39362c686d61632d6d64352d39360000001a6e6f6e652c7a6c6962406f70656e7373682e636f6d2c7a6c69620000001a6e6f6e652c7a6c6962406f70656e7373682e636f6d2c7a6c696200000000000000000000000000")
	serverKexInitData, _ = hex.DecodeString("146e3b627270e950ec5c643782f273f80b000000d4637572766532353531392d736861323536406c69627373682e6f72672c656364682d736861322d6e697374703235362c656364682d736861322d6e697374703338342c656364682d736861322d6e697374703532312c6469666669652d68656c6c6d616e2d67726f75702d65786368616e67652d7368613235362c6469666669652d68656c6c6d616e2d67726f75702d65786368616e67652d736861312c6469666669652d68656c6c6d616e2d67726f757031342d736861312c6469666669652d68656c6c6d616e2d67726f7570312d736861310000001b7373682d7273612c65636473612d736861322d6e69737470323536000000e96165733132382d6374722c6165733139322d6374722c6165733235362d6374722c617263666f75723235362c617263666f75723132382c6165733132382d67636d406f70656e7373682e636f6d2c6165733235362d67636d406f70656e7373682e636f6d2c63686163686132302d706f6c7931333035406f70656e7373682e636f6d2c6165733132382d6362632c336465732d6362632c626c6f77666973682d6362632c636173743132382d6362632c6165733139322d6362632c6165733235362d6362632c617263666f75722c72696a6e6461656c2d636263406c797361746f722e6c69752e7365000000e96165733132382d6374722c6165733139322d6374722c6165733235362d6374722c617263666f75723235362c617263666f75723132382c6165733132382d67636d406f70656e7373682e636f6d2c6165733235362d67636d406f70656e7373682e636f6d2c63686163686132302d706f6c7931333035406f70656e7373682e636f6d2c6165733132382d6362632c336465732d6362632c626c6f77666973682d6362632c636173743132382d6362632c6165733139322d6362632c6165733235362d6362632c617263666f75722c72696a6e6461656c2d636263406c797361746f722e6c69752e736500000192686d61632d6d64352d65746d406f70656e7373682e636f6d2c686d61632d736861312d65746d406f70656e7373682e636f6d2c756d61632d36342d65746d406f70656e7373682e636f6d2c756d61632d3132382d65746d406f70656e7373682e636f6d2c686d61632d736861322d3235362d65746d406f70656e7373682e636f6d2c686d61632d736861322d3531322d65746d406f70656e7373682e636f6d2c686d61632d726970656d643136302d65746d406f70656e7373682e636f6d2c686d61632d736861312d39362d65746d406f70656e7373682e636f6d2c686d61632d6d64352d39362d65746d406f70656e7373682e636f6d2c686d61632d6d64352c686d61632d736861312c756d61632d3634406f70656e7373682e636f6d2c756d61632d313238406f70656e7373682e636f6d2c686d61632d736861322d3235362c686d61632d736861322d3531322c686d61632d726970656d643136302c686d61632d726970656d64313630406f70656e7373682e636f6d2c686d61632d736861312d39362c686d61632d6d64352d393600000192686d61632d6d64352d65746d406f70656e7373682e636f6d2c686d61632d736861312d65746d406f70656e7373682e636f6d2c756d61632d36342d65746d406f70656e7373682e636f6d2c756d61632d3132382d65746d406f70656e7373682e636f6d2c686d61632d736861322d3235362d65746d406f70656e7373682e636f6d2c686d61632d736861322d3531322d65746d406f70656e7373682e636f6d2c686d61632d726970656d643136302d65746d406f70656e7373682e636f6d2c686d61632d736861312d39362d65746d406f70656e7373682e636f6d2c686d61632d6d64352d39362d65746d406f70656e7373682e636f6d2c686d61632d6d64352c686d61632d736861312c756d61632d3634406f70656e7373682e636f6d2c756d61632d313238406f70656e7373682e636f6d2c686d61632d736861322d3235362c686d61632d736861322d3531322c686d61632d726970656d643136302c686d61632d726970656d64313630406f70656e7373682e636f6d2c686d61632d736861312d39362c686d61632d6d64352d3936000000156e6f6e652c7a6c6962406f70656e7373682e636f6d000000156e6f6e652c7a6c6962406f70656e7373682e636f6d00000000000000000000000000")
)

func TestUnmarshalClientKexInit(t *testing.T) {
	fmt.Println(hex.Dump(clientKexInitData))

	var initMsg KexInitMsg
	err := Unmarshal(clientKexInitData, &initMsg)
	if err != nil {
		t.Fatal(err)
	}

	// spew.Dump(initMsg)
}

func TestUnmarshalServerKexInit(t *testing.T) {
	fmt.Println(hex.Dump(clientKexInitData))

	var initMsg KexInitMsg
	err := Unmarshal(serverKexInitData, &initMsg)
	if err != nil {
		t.Fatal(err)
	}

	// spew.Dump(initMsg)
}

func TestGetClientHello(t *testing.T) {
	fmt.Println("Opening file ssh.pcap")
	handle, err := pcap.OpenOffline("ssh.pcap")
	if err != nil {
		log.Fatal(err)
	}

	// set bpf
	//err = handle.SetBPFFilter(*flagBPF)
	//if err != nil {
	//	log.Fatal(err)
	//}

	// create packet source
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	// handle packets
	for packet := range packetSource.Packets() {
		// process TLS client hello
		clientHello := GetClientHello(packet)
		if clientHello != nil {
			destination := "[" + packet.NetworkLayer().NetworkFlow().Dst().String() + ":" + packet.TransportLayer().TransportFlow().Dst().String() + "]"
			log.Printf("%s Client hello ", destination)
		}
	}
}
