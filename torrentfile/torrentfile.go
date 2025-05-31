package torrentfile

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"torry/downloader"
	"torry/peers"

	bencode "github.com/jackpal/bencode-go"
)

type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type bencodeTorrentInfo struct {
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
}

type bencodeTorrent struct {
	Announce string             `bencode:"announce"`
	Info     bencodeTorrentInfo `bencode:"info"`
}

func (btfo *bencodeTorrent) toProcessedTorrentFile() (TorrentFile, error) {
	infoHash, err := btfo.Info.hash()
	if err != nil {
		return TorrentFile{}, err
	}

	pieceHashes, err := btfo.Info.splitPieceHashes()
	if err != nil {
		return TorrentFile{}, err
	}

	t := TorrentFile{
		Announce:    btfo.Announce,
		InfoHash:    infoHash,
		PieceHashes: pieceHashes,
		PieceLength: btfo.Info.PieceLength,
		Length:      btfo.Info.Length,
		Name:        btfo.Info.Name,
	}

	return t, nil
}

func (btfi *bencodeTorrentInfo) hash() ([20]byte, error) {
	var buff bytes.Buffer

	err := bencode.Marshal(&buff, *btfi)
	if err != nil {
		fmt.Println("Error marshalling torrent info into buffer", err)
	}

	infohash := sha1.Sum(buff.Bytes())

	return infohash, nil
}

func (btfi *bencodeTorrentInfo) splitPieceHashes() ([][20]byte, error) {
	hashLength := 20

	// note: casting the binary hashes into bytes and storing in a buffer
	buff := []byte(btfi.Pieces)

	if len(buff)%hashLength != 0 {
		err := fmt.Errorf("malformed pieces of length %d", len(buff))
		return [][20]byte{}, err
	}

	hashesCount := len(buff) / hashLength
	hashes := make([][20]byte, hashesCount)

	for i := range hashesCount {
		/*
			note: basically constucting n slices of 20-byte arrays of hashes
			where n is the hashesCount calculated earlier. hashes[0] is the
			first piece hash
		*/
		copy(hashes[i][:], buff[i*hashLength:(i+1)*hashLength])
	}

	return hashes, nil
}

func OpenTorrentFile(filePath string) (TorrentFile, error) {

	file, err := os.Open(filePath)

	if err != nil {
		fmt.Println("Error opening torrent file", err)
	}

	defer file.Close()

	bt := bencodeTorrent{}

	/*
		note: we are parsing the content of the torrent file and
		storing it's values in an struct/object (bencodeTorrent)
	*/
	err = bencode.Unmarshal(file, &bt)

	if err != nil {
		fmt.Println("Error Unmarshalling:", err)
	}

	return bt.toProcessedTorrentFile()
}

func (tf *TorrentFile) BuildTrackerURL(peerID [20]byte, port uint16) (string, error) {
	base, err := url.Parse(tf.Announce)

	if err != nil {
		fmt.Println("Error Parsing URL:", err)
	}

	params := url.Values{
		"info_hash":  []string{string(tf.InfoHash[:])},
		"peer_id":    []string{string(peerID[:])},
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(int(tf.Length))},
	}

	base.RawQuery = params.Encode()

	return base.String(), nil
}

func (t *TorrentFile) getPeers(peerID [20]byte, port uint16) ([]peers.Peer, error) {
	trackerURL, err := t.BuildTrackerURL(peerID, port)
	/* note
	This is what the URL looks like after encoding the params
		http://bttracker.debian.org:6969/announce?compact=1&downloaded=0&info
		_hash=%A5%94%DF%3A%B9%B4%D3%95b%CC%05%F91%98i%A2%18%12%1FG&left=67842
		8672&peer_id=e%F5_GM%B3hRJ~%A69%04%EC%9Ft%98%B3%14%24&port=2131&uploa
		ded=0
	*/
	if err != nil {
		fmt.Println(err)
	}

	resp, err := http.Get(trackerURL)

	if err != nil {
		fmt.Println("Error sending request to tracker URL", err)
	}

	defer resp.Body.Close()

	trackerResponse := peers.TrackerURLResponse{}

	err = bencode.Unmarshal(resp.Body, &trackerResponse)

	if err != nil {
		fmt.Println("Error parsing tracker response", err)
	}

	/*
		note: casting the bencoded peers binary string to a slice of bytes for
		processing/unmarshalling
	*/
	peers, err := peers.UnmarshallPeers([]byte(trackerResponse.Peers))
	if err != nil {
		return nil, err
	}

	return peers, nil
}

func (t *TorrentFile) D2f(path string) error {
	var peerID [20]byte
	_, err := rand.Read(peerID[:])
	if err != nil {
		return err
	}

	peers, err := t.getPeers(peerID, 6881)
	if err != nil {
		return err
	}

	torrent := downloader.Torrent{
		Peers:       peers,
		PeerID:      peerID,
		InfoHash:    t.InfoHash,
		PieceHashes: t.PieceHashes,
		PieceLength: t.PieceLength,
		Length:      t.Length,
		Name:        t.Name,
	}

	fmt.Println(torrent.Name)

	buf, err := torrent.Download()
	if err != nil {
		return err
	}

	outFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer outFile.Close()
	_, err = outFile.Write(buf)
	if err != nil {
		return err
	}
	return nil
}
