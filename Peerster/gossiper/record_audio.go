package gossiper

// https://blog.golang.org/c-go-cgo

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/gordonklaus/portaudio"
	wave "github.com/zenwerk/go-wave"
	"go.dedis.ch/onet/log"
)

func errCheck(err error) {

	if err != nil {
		log.Error(err)
		panic(err)
	}
}

const escape = '\x00'

func setup() (*portaudio.Stream, *wave.Writer, []byte) {
	if len(os.Args) != 2 {
		fmt.Printf("Usage : %s <audiofilename.wav>\n", os.Args[0])
		os.Exit(0)
	}

	audioFileName := os.Args[1]
	fmt.Println("Recording. Press ESC to quit.")

	if !strings.HasSuffix(audioFileName, ".wav") {
		audioFileName += ".wav"
	}
	waveFile, err := os.Create(audioFileName)
	errCheck(err)

	// https://github.com/gordonklaus/portaudio
	inputChannels := 1
	outputChannels := 0
	sampleRate := 44100
	buffer := make([]byte, 64)

	// Initialize portaudio and create stream
	portaudio.Initialize()
	stream, err := portaudio.OpenDefaultStream(inputChannels, outputChannels, float64(sampleRate), len(buffer), buffer)
	errCheck(err)

	// https://github.com/zenwerk/go-wave
	// Wave file writer setup
	param := wave.WriterParam{
		Out:           waveFile,
		Channel:       inputChannels,
		SampleRate:    sampleRate,
		BitsPerSample: 8, // for 16, change to WriteSample16()
	}

	waveWriter, err := wave.NewWriter(param)
	errCheck(err)

	return stream, waveWriter, buffer
}

func RecordAudioFromMic() {

	stream, waveWriter, buffer := setup()
	go detectEscapeKeyInput(stream, waveWriter)

	// recording in progress ticker. From good old DOS days.
	ticker := []string{
		"-",
		"\\",
		"/",
		"|",
	}
	rand.Seed(time.Now().UnixNano())

	// start reading from microphone
	errCheck(stream.Start())
	for {
		errCheck(stream.Read())

		fmt.Printf("\r Now recording. [%v]", ticker[rand.Intn(len(ticker)-1)])

		// write input bytes to file
		_, err := waveWriter.Write([]byte(buffer)) // WriteSample16 for 16 bits
		errCheck(err)
	}
	errCheck(stream.Stop())
}

// Detect when user presses Escape (used to stop recording)
// https://github.com/eiannone/keyboard
func detectEscapeKeyInput(stream *portaudio.Stream, waveWriter *wave.Writer) {
	key, _, err := keyboard.GetSingleKey()
	if err != nil {
		panic(err)
	}
	if key == escape {
		// On escape key-press, stop recording and close wave writer and stream, terminate portaudio
		fmt.Println("Clearing resources")
		waveWriter.Close()
		stream.Close()
		portaudio.Terminate()
		os.Exit(0)
	}
}
