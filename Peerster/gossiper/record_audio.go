package gossiper

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/gordonklaus/portaudio"
	wave "github.com/zenwerk/go-wave"
	"go.dedis.ch/onet/log"

	"github.com/briandowns/spinner"
)

func RecordAudioFromMic() {

	stream, waveWriter, buffer := setup()
	s := spinner.New(spinner.CharSets[43], 100*time.Millisecond)
	// start reading from microphone
	errCheck(stream.Start())
	fmt.Printf("\r Now recording.\n")
	s.Start()

	go detectEscapeKeyInput(stream, waveWriter, s)
	for {
		errCheck(stream.Read())

		// write input bytes to file
		_, err := waveWriter.Write([]byte(buffer)) // WriteSample16 for 16 bits
		errCheck(err)
	}
}

// 			HELPERS 			//
// =====================
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

// Detect when user presses Escape (used to stop recording)
// https://github.com/eiannone/keyboard
func detectEscapeKeyInput(stream *portaudio.Stream, waveWriter *wave.Writer, s *spinner.Spinner) {
	for {
		_, key, err := keyboard.GetSingleKey()
		if err != nil {
			panic(err)
		}
		if key == keyboard.KeyEsc {
			// On escape key-press, stop recording and close wave writer and stream, terminate portaudio
			s.Stop()
			fmt.Println("\n Finish recording. Clearing resources")
			waveWriter.Close()
			errCheck(stream.Stop())
			stream.Close()
			portaudio.Terminate()
			os.Exit(0)
		}
	}
}

func errCheck(err error) {
	if err != nil {
		log.Error(err)
		panic(err)
	}
}

// https://github.com/gordonklaus/portaudio/blob/master/examples/record.go
// https://socketloop.com/tutorials/golang-record-voice-audio-from-microphone-to-wav-file
// https://medium.com/@valentijnnieman_79984/how-to-build-an-audio-streaming-server-in-go-part-1-1676eed93021
