package board

import "slices"

const columns = 5

// nav is the Board's selection: focused column, selected Card and scroll offset per column.
type nav struct {
	col      int          // focused column 0..4
	sel      [columns]int // selected Card index per column
	off      [columns]int // index of the first visible Card per column
	pendingG bool         // first "g" of "gg" seen
}

// key applies one navigation key given the Cards per column and how many fit in a column;
// other keys only cancel a pending "g".
func (n *nav) key(k string, counts [columns]int, visible int) {
	g := n.pendingG
	n.pendingG = false
	half := max(1, visible/2)
	switch k {
	case "h":
		n.col = n.next(counts, n.col-1, -1)
	case "l":
		n.col = n.next(counts, n.col+1, 1)
	case "1", "2", "3", "4", "5":
		if c := int(k[0] - '1'); counts[c] > 0 {
			n.col = c
		}
	case "j":
		n.sel[n.col]++
	case "k":
		n.sel[n.col]--
	case "g":
		n.pendingG = !g
		if g {
			n.sel[n.col] = 0
		}
	case "G":
		n.sel[n.col] = counts[n.col] - 1
	case "ctrl+d":
		n.sel[n.col] += half
	case "ctrl+u":
		n.sel[n.col] -= half
	default:
		return
	}
	n.fit(n.col, counts[n.col], visible)
}

// keep restores each column's selection by Change name after a reload, clamping the index if the name is gone.
func (n *nav) keep(old, fresh [columns][]string, visible int) {
	for c := range columns {
		if n.sel[c] < len(old[c]) {
			if i := slices.Index(fresh[c], old[c][n.sel[c]]); i >= 0 {
				n.sel[c] = i
			}
		}
		n.fit(c, len(fresh[c]), visible)
	}
	var counts [columns]int
	for c := range columns {
		counts[c] = len(fresh[c])
	}
	if counts[n.col] == 0 { // look right first, then left; Draft when everything is empty
		n.col = n.next(counts, n.col, 1)
		n.col = n.next(counts, n.col, -1)
		if counts[n.col] == 0 {
			n.col = 0
		}
	}
}

// next returns the first column with Cards from c on in direction dir, or the focused column if none.
func (n *nav) next(counts [columns]int, c, dir int) int {
	for ; c >= 0 && c < columns; c += dir {
		if counts[c] > 0 {
			return c
		}
	}
	return n.col
}

// fit clamps column c's selection to its count Cards and scrolls so the selection is visible.
func (n *nav) fit(c, count, visible int) {
	n.sel[c] = max(0, min(n.sel[c], count-1))
	n.off[c] = min(n.sel[c], max(min(n.off[c], count-visible), n.sel[c]-visible+1, 0))
}
