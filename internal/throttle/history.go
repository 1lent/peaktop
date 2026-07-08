package throttle

const defaultBufferSize = 60

var sparkChars = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

type RingBuffer struct {
	data   []float64
	size   int
	cursor int
	isFull bool
}

func NewRingBuffer(size int) *RingBuffer {
	if size <= 0 {
		size = defaultBufferSize
	}
	return &RingBuffer{
		data: make([]float64, size),
		size: size,
	}
}

func (rb *RingBuffer) Push(value float64) {
	rb.data[rb.cursor] = value
	rb.cursor++
	if rb.cursor >= rb.size {
		rb.cursor = 0
		rb.isFull = true
	}
}

func (rb *RingBuffer) Values() []float64 {
	if !rb.isFull {
		result := make([]float64, rb.cursor)
		copy(result, rb.data[:rb.cursor])
		return result
	}

	result := make([]float64, rb.size)
	copy(result, rb.data[rb.cursor:])
	copy(result[rb.size-rb.cursor:], rb.data[:rb.cursor])
	return result
}

func (rb *RingBuffer) Len() int {
	if !rb.isFull {
		return rb.cursor
	}
	return rb.size
}

func (rb *RingBuffer) Sparkline(width int) string {
	if width <= 0 {
		width = rb.Len()
	}

	values := rb.Values()
	numValues := len(values)
	if numValues == 0 {
		return ""
	}

	step := 1
	if numValues > width {
		step = numValues / width
		if step < 1 {
			step = 1
		}
	}

	min, max := rb.minMax(values)

	runes := make([]rune, width)
	for i := 0; i < width; i++ {
		srcIdx := i * step
		if srcIdx >= numValues {
			srcIdx = numValues - 1
		}

		normalized := normalizeValue(values[srcIdx], min, max)
		idx := int(normalized * float64(len(sparkChars)-1))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(sparkChars) {
			idx = len(sparkChars) - 1
		}
		runes[i] = sparkChars[idx]
	}

	return string(runes)
}

func (rb *RingBuffer) minMax(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0, 1
	}

	min := values[0]
	max := values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	if min == max {
		return min, max + 1
	}
	return min, max
}

func normalizeValue(value, min, max float64) float64 {
	rangeSize := max - min
	if rangeSize <= 0 {
		return 0
	}
	return (value - min) / rangeSize
}
