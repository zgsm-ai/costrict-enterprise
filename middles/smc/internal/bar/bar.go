package bar

import (
	pb "github.com/cheggaaa/pb/v3"
)

var pbar *pb.ProgressBar

/**
 * Create new progress bar
 * @param fileSize total size for progress calculation
 * @param title prefix text to display
 * @return *pb.ProgressBar created progress bar
 */
func CreatePbar(fileSize int64, title string) *pb.ProgressBar {
	pbar = pb.Full.Start64(fileSize)
	pbar.Set(pb.Bytes, true)
	pbar.Set("prefix", title)
	return pbar
}

/**
 * Complete and clean up progress bar
 * @param title final prefix text to display
 */
func FinishPbar(title string) {
	if pbar == nil {
		return
	}
	pbar.Set("prefix", title)
	pbar.Finish()
	pbar = nil
}

/**
 * Increment progress bar by specified amount
 * @param deltaSize number of bytes/units to increment progress
 */
func Add64(deltaSize int64) {
	if pbar == nil {
		return
	}
	pbar.Add64(deltaSize)
}
