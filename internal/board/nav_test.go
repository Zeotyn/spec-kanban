package board

import "testing"

func press(n *nav, counts [columns]int, visible int, keys ...string) {
	for _, k := range keys {
		n.key(k, counts, visible)
	}
}

func TestNavColumnsRememberSelection(t *testing.T) {
	counts := [columns]int{0, 5, 4, 0, 0}
	n := nav{col: 1, sel: [columns]int{0, 2, 1, 0, 0}}
	press(&n, counts, 10, "l")
	if n.col != 2 || n.sel[2] != 1 {
		t.Fatalf("after l: col %d sel %d, want 2 1", n.col, n.sel[2])
	}
	press(&n, counts, 10, "h")
	if n.col != 1 || n.sel[1] != 2 {
		t.Fatalf("after h: col %d sel %d, want 1 2", n.col, n.sel[1])
	}
}

func TestNavJumpToEnd(t *testing.T) {
	counts := [columns]int{12}
	var n nav
	press(&n, counts, 20, "G")
	if n.sel[0] != 11 {
		t.Fatalf("after G: sel %d, want 11", n.sel[0])
	}
	press(&n, counts, 20, "g", "g")
	if n.sel[0] != 0 {
		t.Fatalf("after gg: sel %d, want 0", n.sel[0])
	}
}

func TestNavGCancelledByOtherKey(t *testing.T) {
	counts := [columns]int{12}
	n := nav{sel: [columns]int{5}}
	press(&n, counts, 20, "g", "j", "g")
	if n.sel[0] != 6 || !n.pendingG {
		t.Fatalf("sel %d pendingG %v, want 6 true", n.sel[0], n.pendingG)
	}
}

func TestNavJumpToColumn(t *testing.T) {
	counts := [columns]int{1, 1, 1, 0, 1}
	var n nav
	if n.key("5", counts, 1); n.col != 4 {
		t.Fatalf("after 5: col %d, want 4", n.col)
	}
	if n.key("4", counts, 1); n.col != 4 {
		t.Fatalf("after 4 on an empty column: col %d, want 4", n.col)
	}
	press(&n, counts, 1, "1")
	if n.col != 0 {
		t.Fatalf("after 1: col %d, want 0", n.col)
	}
}

func TestNavSkipsEmptyColumns(t *testing.T) {
	counts := [columns]int{2, 0, 3, 0, 0}
	var n nav
	if press(&n, counts, 5, "l"); n.col != 2 {
		t.Fatalf("l: col %d, want 2", n.col)
	}
	if press(&n, counts, 5, "l"); n.col != 2 {
		t.Fatalf("l with only empty columns to the right: col %d, want 2", n.col)
	}
	if press(&n, counts, 5, "h"); n.col != 0 {
		t.Fatalf("h: col %d, want 0", n.col)
	}
}

func TestKeepMovesFocusOffEmptyColumn(t *testing.T) {
	tests := []struct {
		name  string
		col   int
		fresh [columns][]string
		want  int
	}{
		{"right first", 3, [columns][]string{0: {"a"}, 4: {"b"}}, 4},
		{"left when nothing to the right", 3, [columns][]string{1: {"a"}}, 1},
		{"stays when it has Cards", 3, [columns][]string{0: {"a"}, 3: {"c"}}, 3},
		{"Draft when all are empty", 3, [columns][]string{}, 0},
		{"first load into empty Draft", 0, [columns][]string{2: {"a"}}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := nav{col: tt.col}
			n.keep([columns][]string{}, tt.fresh, 5)
			if n.col != tt.want {
				t.Errorf("col %d, want %d", n.col, tt.want)
			}
		})
	}
}

func TestNavEmptyColumn(t *testing.T) {
	var n nav
	press(&n, [columns]int{}, 5, "j", "k", "G", "ctrl+d", "ctrl+u")
	if n.sel[0] != 0 || n.off[0] != 0 {
		t.Fatalf("sel %d off %d, want 0 0", n.sel[0], n.off[0])
	}
}

func TestNavScrollsDown(t *testing.T) {
	counts := [columns]int{10}
	var n nav
	press(&n, counts, 6, "j", "j", "j", "j", "j", "j", "j")
	if n.sel[0] != 7 || n.off[0] != 2 {
		t.Fatalf("sel %d off %d, want 7 2", n.sel[0], n.off[0])
	}
	press(&n, counts, 6, "g", "g")
	if n.off[0] != 0 {
		t.Fatalf("after gg: off %d, want 0", n.off[0])
	}
}

func TestNavHalfPage(t *testing.T) {
	counts := [columns]int{10}
	var n nav
	press(&n, counts, 6, "ctrl+d")
	if n.sel[0] != 3 {
		t.Fatalf("ctrl+d: sel %d, want 3", n.sel[0])
	}
	press(&n, counts, 6, "ctrl+d", "ctrl+d", "ctrl+d")
	if n.sel[0] != 9 || n.off[0] != 4 {
		t.Fatalf("ctrl+d clamp: sel %d off %d, want 9 4", n.sel[0], n.off[0])
	}
	press(&n, counts, 6, "ctrl+u")
	if n.sel[0] != 6 {
		t.Fatalf("ctrl+u: sel %d, want 6", n.sel[0])
	}
	press(&n, counts, 1, "ctrl+u")
	if n.sel[0] != 5 || n.off[0] != 5 {
		t.Fatalf("ctrl+u visible 1: sel %d off %d, want 5 5", n.sel[0], n.off[0])
	}
}

func TestNavBounds(t *testing.T) {
	counts := [columns]int{3, 0, 0, 0, 3}
	var n nav
	press(&n, counts, 5, "h", "k")
	if n.col != 0 || n.sel[0] != 0 {
		t.Fatalf("at start: col %d sel %d, want 0 0", n.col, n.sel[0])
	}
	press(&n, counts, 5, "l", "l", "l", "l", "l", "j", "j", "j", "j")
	if n.col != 4 || n.sel[4] != 2 {
		t.Fatalf("at end: col %d sel %d, want 4 2", n.col, n.sel[4])
	}
}

func TestNavOtherKey(t *testing.T) {
	n := nav{pendingG: true}
	if n.key("x", [columns]int{3}, 5); n.pendingG || n.sel[0] != 0 {
		t.Fatalf("x: pendingG kept or selection moved")
	}
}

func TestKeepByName(t *testing.T) {
	old := [columns][]string{2: {"a", "m1-board-view", "c"}}
	fresh := [columns][]string{2: {"m1-board-view", "c"}, 3: {"a"}}
	n := nav{col: 2, sel: [columns]int{2: 1}}
	n.keep(old, fresh, 5)
	if n.sel[2] != 0 {
		t.Fatalf("sel %d, want 0", n.sel[2])
	}
}

func TestKeepClampsWhenGone(t *testing.T) {
	old := [columns][]string{{"a", "b", "c", "d", "e", "f"}}
	fresh := [columns][]string{{"a", "b", "c"}}
	n := nav{sel: [columns]int{5}, off: [columns]int{4}}
	n.keep(old, fresh, 2)
	if n.sel[0] != 2 || n.off[0] != 1 {
		t.Fatalf("sel %d off %d, want 2 1", n.sel[0], n.off[0])
	}
	n.keep(fresh, [columns][]string{}, 2)
	if n.sel[0] != 0 || n.off[0] != 0 {
		t.Fatalf("empty: sel %d off %d, want 0 0", n.sel[0], n.off[0])
	}
}
