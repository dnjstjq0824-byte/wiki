package diff

// Operation represents an edit operation type
type Operation int

const (
	Equal  Operation = 0
	Insert Operation = 1
	Delete Operation = 2
)

// Chunk represents a diff fragment
type Chunk struct {
	Op   Operation `json:"op"`
	Text string    `json:"text"`
}

// DiffLines calculates line-by-line diff using Myers Diff Algorithm
func DiffLines(a, b []string) []Chunk {
	n := len(a)
	m := len(b)
	max := n + m

	if max == 0 {
		return []Chunk{}
	}

	v := make([]int, 2*max+1)
	trace := make([][]int, 0, max+1)

	for d := 0; d <= max; d++ {
		vCopy := make([]int, len(v))
		copy(vCopy, v)
		trace = append(trace, vCopy)

		for k := -d; k <= d; k += 2 {
			kIdx := k + max
			var x int
			if k == -d || (k != d && v[kIdx-1] < v[kIdx+1]) {
				x = v[kIdx+1]
			} else {
				x = v[kIdx-1] + 1
			}

			y := x - k

			for x < n && y < m && a[x] == b[y] {
				x++
				y++
			}

			v[kIdx] = x

			if x >= n && y >= m {
				return backtrack(trace, a, b, n, m, max)
			}
		}
	}

	return nil
}

func backtrack(trace [][]int, a, b []string, n, m, max int) []Chunk {
	var chunks []Chunk
	x := n
	y := m

	for d := len(trace) - 1; d > 0; d-- {
		k := x - y
		kIdx := k + max
		v := trace[d]

		var prevK int
		if k == -d || (k != d && v[kIdx-1] < v[kIdx+1]) {
			prevK = k + 1
		} else {
			prevK = k - 1
		}

		prevX := trace[d-1][prevK+max]
		prevY := prevX - prevK

		for x > prevX && y > prevY {
			chunks = append(chunks, Chunk{Op: Equal, Text: a[x-1]})
			x--
			y--
		}

		if d > 0 {
			if x == prevX {
				chunks = append(chunks, Chunk{Op: Insert, Text: b[prevY]})
			} else if y == prevY {
				chunks = append(chunks, Chunk{Op: Delete, Text: a[prevX]})
			}
		}

		x = prevX
		y = prevY
	}

	for x > 0 && y > 0 {
		chunks = append(chunks, Chunk{Op: Equal, Text: a[x-1]})
		x--
		y--
	}

	// Reverse to chronological order
	for i, j := 0, len(chunks)-1; i < j; i, j = i+1, j-1 {
		chunks[i], chunks[j] = chunks[j], chunks[i]
	}

	return chunks
}

