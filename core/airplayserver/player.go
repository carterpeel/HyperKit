package airplayserver

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/carterpeel/bobcaygeon/player"
	"github.com/carterpeel/bobcaygeon/rtsp"
	"github.com/hajimehoshi/oto"
	log "github.com/sirupsen/logrus"
	"hyperkit/core/airplayserver/bluetoothproxy"
	"os"
	"sync"
	"sync/atomic"
)

// LocalPlayer is a player that will just play the audio locally
type LocalPlayer struct {
	volLock sync.RWMutex
	volume  float64

	pipeFile   string
	btpx       *bluetoothproxy.BluetoothProxy
	curSession *rtsp.Session
	pauseChan  chan struct{}
	pctx       *oto.Context
}

// NewBluetoothPlayer instantiates a new LocalPlayer
func NewBluetoothPlayer(pipeFile string, bluetoothName string) (lp *LocalPlayer, err error) {
	lp = &LocalPlayer{
		volume:   1,
		pipeFile: pipeFile,
		volLock:  sync.RWMutex{},
	}

	lp.pctx, err = oto.NewContext(44100, 2, 2, 10000)
	if err != nil {
		return nil, fmt.Errorf("error initializing player: %v", err)
	}

	log.Infof("Attempting to proxy device %v...", bluetoothName)
	if lp.btpx, err = bluetoothproxy.ProxyBluetoothDevice(bluetoothName); err != nil {
		return nil, err
	}

	return lp, nil
}

// Play will play the packets received on the specified session
func (lp *LocalPlayer) Play(session *rtsp.Session) {
	go lp.playStream(session)
}

func (lp *LocalPlayer) Pause() {
	lp.pauseChan <- struct{}{}
}

// SetVolume accepts a float between 0 (mute) and 1 (full volume)
func (lp *LocalPlayer) SetVolume(volume float64) {
	lp.volLock.Lock()
	defer lp.volLock.Unlock()
	lp.volume = volume

}

// SetTrack sets the track for the player
func (lp *LocalPlayer) SetTrack(album string, artist string, title string) {
	// no op for now
}

// SetAlbumArt sets the album art for the player
func (lp *LocalPlayer) SetAlbumArt(artwork []byte) {
	// no op for now
}

// SetMute will mute or unmute the player
func (lp *LocalPlayer) SetMute(isMuted bool) {
	// no op for now
}

// GetIsMuted returns muted state
func (lp *LocalPlayer) GetIsMuted() bool {
	return false
}

// GetTrack returns the track
func (lp *LocalPlayer) GetTrack() player.Track {
	return player.Track{}
}

func (lp *LocalPlayer) QuitCurrentSession() {
	if lp.curSession != nil {
		go func() {
			ch := make(chan struct{})
			lp.curSession.Close(ch)
			ch <- struct{}{}
			close(ch)
		}()
	}
}

func (lp *LocalPlayer) playStream(session *rtsp.Session) {
	p := lp.pctx.NewPlayer()
	defer func(p *oto.Player) {
		_ = p.Close()
	}(p)

	pipeFd, err := os.OpenFile(lp.pipeFile, os.O_WRONLY, 0600)
	if err != nil {
		log.Errorf("Error opening FIFO pipe: %v\n", err)
		return
	}
	defer func(pipeFd *os.File) {
		log.Infof("Closing FIFO pipe '%v'", lp.pipeFile)
		_ = pipeFd.Close()
	}(pipeFd)

	log.Debugf("Opened FIFO pipe '%v'", lp.pipeFile)

	decoder := GetCodec(session)
	for d := range session.DataChan {
		lp.volLock.RLock()
		vol := lp.volume
		lp.volLock.RUnlock()
		decoded, err := decoder(d)
		if err != nil {
			log.Warnf("Error decoding packet: %v\n", err)
		}

		// Atomic value indicating failure during write
		failed := &atomic.Value{}
		failed.Store(false)

		wg := new(sync.WaitGroup)
		wg.Add(2)

		adj := AdjustAudio(decoded, vol)

		go func() {
			defer wg.Done()
			if _, err = p.Write(adj); err != nil {
				log.Debugf("Caught EOF on otoctx audio stream: %v\n", err)
				failed.Store(true)
			}
		}()
		go func() {
			defer wg.Done()
			if _, err = pipeFd.Write(adj); err != nil {
				log.Debugf("Caught EOF on pipeFd write stream: %v\n", err)
				failed.Store(true)
			}
		}()
		wg.Wait()
		if failed.Load().(bool) == true {
			log.Infoln("Data stream ended! Closing stream writer...")
			return
		}
	}
}

func AdjustAudio(raw []byte, vol float64) []byte {
	if vol == 1 {
		return raw
	}
	adjusted := new(bytes.Buffer)
	for i := 0; i < len(raw); i = i + 2 {
		var val int16
		b := raw[i : i+2]
		buf := bytes.NewReader(b)
		if err := binary.Read(buf, binary.LittleEndian, &val); err != nil {
			log.Warnf("Error reading binary data: %v\n", err)
			return raw
		}
		val = int16(vol * float64(val))
		val = min(32767, val)
		val = max(-32767, val)
		_ = binary.Write(adjusted, binary.LittleEndian, val)
	}

	return adjusted.Bytes()
}
