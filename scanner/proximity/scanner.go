package proximity

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

var (
	iface            string = "mon0"
	snapLen          int32  = 256
	radioLayer       layers.RadioTap
	dot11layer       layers.Dot11
	channel          = 5
	captureIntervals []time.Time
)

/*
	TODO:
	0. store readings in local db
	1. Generate config file
	2. Gulp interface
	3. Create a sending listening mechanism!
*/

/*
type device struct {
	mac       string
	signal    int
	timestamp time
}

type station struct {
	mac       string
	signal    int
	timestamp time
}*/

/*
	Main issues with going for probe-requests
	1. If already connect to any network, no request will be sent
	2. If wifi
*/

func Run() {
	go monitorNetworkTraffic()
	//go channelManager()
}

func channelManager() {
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for range ticker.C {
			//fmt.Println("Changing channel to ", channel+1)
			changeChannel()
		}
	}()
}

func sendData() {

}

/*
* monitorNetworkTraffic initiates packet monitoring using pcap.
* It will use the network interface defined in the initial config.
* Packets will be detected by the handlePacket function below
* TODO: Consider using a more complex filter
 */
func monitorNetworkTraffic() {
	fmt.Println("starting to scan")
	handle, err := pcap.OpenLive(iface, snapLen, true, pcap.BlockForever)
	fmt.Println(err)
	if err != nil {
		log.Fatal(err)
	}

	defer handle.Close()

	// Set filter

	/*
		type ctl subtype rts will get ready to send signals
		type mgt subtype beacon will get the station beacons (if wanting to map the space)
		see: https://mrncciew.com/2014/10/02/cwap-802-11-control-frame-types/

		see also https://medium.freecodecamp.org/tracking-analyzing-over-200-000-peoples-every-step-at-mit-e736a507ddbf
	*/

	var filter = "type mgt subtype probe-req" //TODO: ADD DESTINATION/SOURCE TO THE FILTER TO AVOID GETTING TOO MANY PACKETS
	err = handle.SetBPFFilter(filter)
	if err != nil {
		fmt.Println("error")
		log.Fatal(err)
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	parser := gopacket.NewDecodingLayerParser(
		layers.LayerTypeRadioTap,
		&radioLayer,
		&dot11layer,
	)

	foundLayerTypes := []gopacket.LayerType{}

	for packet := range packetSource.Packets() {

		parser.DecodeLayers(packet.Data(), &foundLayerTypes)

		if len(foundLayerTypes) >= 2 && radioLayer.DBMAntennaSignal != 0 {
			//

			//fmt.Printf("PACKET \n%+v\n", packet.Metadata().CaptureInfo.Timestamp)
			//fmt.Println(time.Now())

			reAddr := dot11layer.Address1.String()
			trAddr := dot11layer.Address2.String()
			freq := radioLayer.ChannelFrequency
			ptype := dot11layer.Type
			fmt.Printf("Addr1: %s, Addr2: %s, Type: %s Signal: %d Freq: %d\n", reAddr, trAddr, ptype, radioLayer.DBMAntennaSignal, freq)

			captureIntervals = append(captureIntervals, time.Now())
			//addr1 := dot11layer.Address1.String()
			//addr2 := dot11layer.Address2.String()
			//addr3 := dot11layer.Address3.String()
			fmt.Println(captureIntervals)
		}
	}
}

func changeChannel() {
	if channel < 13 {
		channel++
	} else {
		channel = 1
	}

	cmdString := "sudo iw dev " + iface + " set channel " + strconv.Itoa(channel)
	//fmt.Println(cmdString)

	cmd := exec.Command("/bin/sh", "-c", cmdString) //-c flag allows us to execute the full string (instead of seperating each element with ,)
	var output bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		//Combining error with stderr to avoid all errors being exit 1
		fmt.Println(err)
	}
}
