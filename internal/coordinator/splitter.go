package coordinator

import "fmt"

type Block struct {
	ID     string
	X      int
	Y      int
	Width  int
	Height int
}

const minBlockSize = 3

// SplitImage divides an image into blocks, matching the original VB6 algorithm.
// Remainder blocks smaller than minBlockSize are merged with the previous block.
func SplitImage(picWidth, picHeight, blockWidth, blockHeight int) []Block {
	if blockWidth <= 0 || blockHeight <= 0 {
		return []Block{{ID: "0-0", X: 0, Y: 0, Width: picWidth, Height: picHeight}}
	}

	cols := blocksInDim(picWidth, blockWidth)
	rows := blocksInDim(picHeight, blockHeight)

	blocks := make([]Block, 0, cols*rows)
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			x := col * blockWidth
			y := row * blockHeight
			w := blockWidth
			h := blockHeight

			// Last column: take remainder
			if col == cols-1 {
				w = picWidth - x
			}
			// Last row: take remainder
			if row == rows-1 {
				h = picHeight - y
			}

			blocks = append(blocks, Block{
				ID:     fmt.Sprintf("%d-%d", col, row),
				X:      x,
				Y:      y,
				Width:  w,
				Height: h,
			})
		}
	}
	return blocks
}

func blocksInDim(total, blockSize int) int {
	rest := total % blockSize
	full := total / blockSize
	if rest == 0 {
		return full
	}
	// Merge remainder with last block if too small
	if rest < minBlockSize {
		return full
	}
	return full + 1
}
