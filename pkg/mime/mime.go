package mime

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/vorbis"
	"github.com/faiface/beep/wav"
)

// ConvertToWAV converts an audio file to WAV format.
func ConvertToWAV(inputPath, outputPath string) error {
	f, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %v", err)
	}
	defer f.Close()

	var streamer beep.Streamer
	var format beep.Format

	switch filepath.Ext(inputPath) {
	case ".mp3":
		streamer, format, err = mp3.Decode(f)
		if err != nil {
			return fmt.Errorf("failed to decode mp3: %v", err)
		}
	case ".ogg":
		streamer, format, err = vorbis.Decode(f)
		if err != nil {
			return fmt.Errorf("failed to decode ogg: %v", err)
		}
	// Add other cases for m4a, webm etc. Note that you might need additional libraries or tools for some formats.
	default:
		return fmt.Errorf("unsupported audio format")
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer outFile.Close()

	// Encode to WAV and write to the output file
	err = wav.Encode(outFile, streamer, format)
	if err != nil {
		return fmt.Errorf("failed to encode to wav: %v", err)
	}

	return nil
}
