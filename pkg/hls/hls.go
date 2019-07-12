package hls

import (
	"os/exec"
)

// Segment runs the mediafilesegmenter command-line tool. This tool takes a
// media file as an input, wraps it in an MPEG-2 transport stream, and produces
// a series of equal-length files from it, suitable for use in HTTP Live
// Streaming. It also produces an produce an index (playlist) file.
func Segment(songPath string, destPath string) error {
	cmd := exec.Command("mediafilesegmenter", "-a", "-f", destPath, songPath)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
