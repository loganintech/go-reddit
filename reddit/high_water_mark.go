package reddit

func NewHighWaterMark(cap uint32, items ...string) HighWaterMark {
	return &highWaterMark{marks: items, cap: cap}
}

type HighWaterMark interface {
	Len() int
	Top() string
	Push(item string) bool
	Pop() string
}

// Reddit is a crazy API. Using the before query param we're prone to failure because if you do ?before=id and id is deleted, we return no results
// Reddit does not return every item that came "before" (but really, after) the item if the item ID sent is from a deleted record
// So if we track the latest item, and the item gets deleted, we are perma-stuck querying no data. Which also means we can never recover
// How's that for pain in the ass
type highWaterMark struct {
	marks []string
	cap   uint32
}

func (h *highWaterMark) Len() int {
	if h == nil {
		return 0
	}
	return len(h.marks)
}
func (h *highWaterMark) Top() string {
	if h == nil {
		return ""
	}
	return h.marks[h.Len()-1]
}
func (h *highWaterMark) Push(item string) bool {
	if h == nil {
		panic("nil highWaterMark")
	}
	if uint32(h.Len()) == h.cap {
		// Drop from the bottom, we want to keep things most recently seen
		h.marks = h.marks[1:h.Len()]
		h.marks = append(h.marks, item)
		return true
	}
	h.marks = append(h.marks, item)
	return false
}
func (h *highWaterMark) Pop() string {
	if h == nil {
		return ""
	}
	if len(h.marks) == 0 {
		return ""
	}
	item := h.marks[h.Len()-1]
	h.marks = h.marks[:h.Len()-1]
	return item
}
