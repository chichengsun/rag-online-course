// Package chunktext 提供按字符窗口与重叠切分文本，供知识库分块预览与落库。
package chunktext

// Segment 表示一个分块在源文本中的区间与内容；CharStart/CharEnd 为 UTF-8 字节偏移，便于 SUBSTRING。
type Segment struct {
	Index     int    `json:"index"`
	Content   string `json:"content"`
	CharStart int    `json:"char_start"`
	CharEnd   int    `json:"char_end"`
}

// Split 按 UTF-8 字符（非字节）窗口切分；chunkSize、overlap 为字符（rune）数。
func Split(text string, chunkSize, overlap int) []Segment {
	if chunkSize <= 0 {
		chunkSize = 800
	}
	if overlap < 0 {
		overlap = 0
	}
	if overlap >= chunkSize {
		overlap = chunkSize / 5
		if overlap >= chunkSize {
			overlap = 0
		}
	}
	runes := []rune(text)
	if len(runes) == 0 {
		return nil
	}
	var out []Segment
	start := 0
	idx := 0
	for start < len(runes) {
		end := start + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		piece := string(runes[start:end])
		bs, be := runeRangeToByteOffsets(text, start, end)
		out = append(out, Segment{
			Index:     idx,
			Content:   piece,
			CharStart: bs,
			CharEnd:   be,
		})
		idx++
		if end >= len(runes) {
			break
		}
		next := end - overlap
		if next <= start {
			next = start + 1
		}
		start = next
	}
	return out
}

// runeRangeToByteOffsets 将 [rStart, rEnd) 的 rune 下标转为源字符串字节区间 [byteStart, byteEnd)。
func runeRangeToByteOffsets(s string, rStart, rEnd int) (byteStart, byteEnd int) {
	byteEnd = len(s)
	i := 0
	for pos := range s {
		if i == rStart {
			byteStart = pos
		}
		if i == rEnd {
			byteEnd = pos
			return byteStart, byteEnd
		}
		i++
	}
	if rEnd >= i {
		byteEnd = len(s)
	}
	return byteStart, byteEnd
}
