package peers

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
)

type Peer struct {
	IP   net.IP
	Port uint16
}

type TrackerURLResponse struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

func UnmarshallPeers(binPeerList []byte) ([]Peer, error) {
	peerSize := 6

	if len(binPeerList)%peerSize != 0 {
		err := fmt.Errorf("received malformed peers list of size %d", len(binPeerList))
		return []Peer{}, err
	}

	peerCount := len(binPeerList) / peerSize
	peerList := make([]Peer, peerCount)

	for i := range peerCount {
		offset := i * peerSize
		peerList[i].IP = net.IP(binPeerList[offset : offset+4])
		peerList[i].Port = binary.BigEndian.Uint16(binPeerList[offset+4 : offset+6])
	}

	return peerList, nil
}

func (p Peer) Stringify() string {
	return net.JoinHostPort(p.IP.String(), strconv.Itoa(int(p.Port)))
}
