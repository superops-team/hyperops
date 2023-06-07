package tools

import (
	"fmt"
	"math"
	"strings"
)

type DeltaType int

const (
	Common DeltaType = iota
	LeftOnly
	RightOnly
)

// String returns a string representation for DeltaType.
func (t DeltaType) String() string {
	switch t {
	case Common:
		return " "
	case LeftOnly:
		return "-"
	case RightOnly:
		return "+"
	}
	return "?"
}

type DiffRecord struct {
	Payload   string
	Delta     DeltaType
	LineLeft  int
	LineRight int
}

// String returns a string representation of d. The string is a
// concatenation of the delta type and the payload.
func (d DiffRecord) String() string {
	return fmt.Sprintf("%s %s", d.Delta, d.Payload)
}

// Diff returns the result of diffing the seq1 and seq2.
func Diff(seq1, seq2 []string) (diff []DiffRecord) {
	// Trims any common elements at the heads and tails of the
	// sequences before running the diff algorithm. This is an
	// optimization.
	start, end := numEqualStartAndEndElements(seq1, seq2)

	for i, content := range seq1[:start] {
		diff = append(diff, DiffRecord{content, Common, i, i})
	}

	diffRes := compute(seq1[start:len(seq1)-end], seq2[start:len(seq2)-end], start)
	diff = append(diff, diffRes...)

	for i, content := range seq1[len(seq1)-end:] {
		diff = append(diff, DiffRecord{content, Common, len(seq1) - end + i, len(seq2) - end + i})
	}
	return
}

// PPDiff returns the results of diffing left and right as an pretty
// printed string. It will display all the lines of both the sequences
// that are being compared.
// When the left is different from right it will prepend a " - |" before
// the line.
// When the right is different from left it will prepend a " + |" before
// the line.
// When the right and left are equal it will prepend a "   |" before
// the line.
func PPDiff(left, right []string) string {
	var sb strings.Builder

	recs := Diff(right, left)

	for _, diff := range recs {
		var mark string

		switch diff.Delta {
		case RightOnly:
			mark = " + "
		case LeftOnly:
			mark = " - "
		case Common:
			mark = "   "
		}

		// make sure to have line numbers to make sure diff is truly unique
		sb.WriteString(fmt.Sprintf("%s%s\n", mark, diff.Payload))
	}

	return sb.String()
}

// numEqualStartAndEndElements returns the number of elements a and b
// have in common from the beginning and from the end. If a and b are
// equal, start will equal len(a) == len(b) and end will be zero.
func numEqualStartAndEndElements(seq1, seq2 []string) (start, end int) {
	for start < len(seq1) && start < len(seq2) && seq1[start] == seq2[start] {
		start++
	}
	i, j := len(seq1)-1, len(seq2)-1
	for i > start && j > start && seq1[i] == seq2[j] {
		i--
		j--
		end++
	}
	return
}

// intMatrix returns a 2-dimensional slice of ints with the given
// number of rows and columns.
func intMatrix(rows, cols int) [][]int {
	matrix := make([][]int, rows)
	for i := 0; i < rows; i++ {
		matrix[i] = make([]int, cols)
	}
	return matrix
}

// longestCommonSubsequenceMatrix returns the table that results from
// applying the dynamic programming approach for finding the longest
// common subsequence of seq1 and seq2.
func longestCommonSubsequenceMatrix(seq1, seq2 []string) [][]int {
	matrix := intMatrix(len(seq1)+1, len(seq2)+1)
	for i := 1; i < len(matrix); i++ {
		for j := 1; j < len(matrix[i]); j++ {
			if seq1[len(seq1)-i] == seq2[len(seq2)-j] {
				matrix[i][j] = matrix[i-1][j-1] + 1
			} else {
				matrix[i][j] = int(math.Max(float64(matrix[i-1][j]),
					float64(matrix[i][j-1])))
			}
		}
	}
	return matrix
}

// compute is the unexported helper for Diff that returns the results of
// diffing left and right.
func compute(seq1, seq2 []string, startLine int) (diff []DiffRecord) {
	matrix := longestCommonSubsequenceMatrix(seq1, seq2)
	i, j := len(seq1), len(seq2)
	for i > 0 || j > 0 {
		if i > 0 && matrix[i][j] == matrix[i-1][j] {
			diff = append(diff, DiffRecord{seq1[len(seq1)-i], LeftOnly, startLine + len(seq1) - i, startLine + len(seq2) - j})
			i--
		} else if j > 0 && matrix[i][j] == matrix[i][j-1] {
			diff = append(diff, DiffRecord{seq2[len(seq2)-j], RightOnly, startLine + len(seq1) - i, startLine + len(seq2) - j})
			j--
		} else if i > 0 && j > 0 {
			diff = append(diff, DiffRecord{seq1[len(seq1)-i], Common, startLine + len(seq1) - i, startLine + len(seq2) - j})
			i--
			j--
		}
	}
	return
}
