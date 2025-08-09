package reddit

import (
	"testing"
)

func TestHighWaterMark_BasicStackBehavior(t *testing.T) {
	hwm := NewHighWaterMark(3)

	if hwm.Len() != 0 {
		t.Errorf("Expected empty mark to have length 0, got %d", hwm.Len())
	}

	hwm.Push("A")
	hwm.Push("B")
	hwm.Push("C")

	if hwm.Len() != 3 {
		t.Errorf("Expected length 3, got %d", hwm.Len())
	}

	if top := hwm.Top(); top != "C" {
		t.Errorf("Expected top to be 'C', got '%s'", top)
	}

	if popped := hwm.Pop(); popped != "C" {
		t.Errorf("Expected to pop 'C', got '%s'", popped)
	}

	if popped := hwm.Pop(); popped != "B" {
		t.Errorf("Expected to pop 'B', got '%s'", popped)
	}

	if popped := hwm.Pop(); popped != "A" {
		t.Errorf("Expected to pop 'A', got '%s'", popped)
	}

	if hwm.Len() != 0 {
		t.Errorf("Expected empty stack after popping all items, got length %d", hwm.Len())
	}
}

func TestHighWaterMark_CapacityLimitsAndDropping(t *testing.T) {
	hwm := NewHighWaterMark(3)

	dropped := hwm.Push("A")
	if dropped {
		t.Error("Expected Push to return false when not at capacity")
	}

	hwm.Push("B")
	hwm.Push("C")

	dropped = hwm.Push("D")
	if !dropped {
		t.Error("Expected Push to return true when at capacity and dropping items")
	}

	if hwm.Len() != 3 {
		t.Errorf("Expected length to remain 3 after exceeding capacity, got %d", hwm.Len())
	}

	if top := hwm.Top(); top != "D" {
		t.Errorf("Expected top to be 'D' after pushing when at capacity, got '%s'", top)
	}

	if popped := hwm.Pop(); popped != "D" {
		t.Errorf("Expected to pop 'D', got '%s'", popped)
	}
	if popped := hwm.Pop(); popped != "C" {
		t.Errorf("Expected to pop 'C', got '%s'", popped)
	}
	if popped := hwm.Pop(); popped != "B" {
		t.Errorf("Expected to pop 'B', got '%s'", popped)
	}

	hwm.Push("E")
	hwm.Push("F")
	hwm.Push("G")
	hwm.Push("H")

	if hwm.Len() != 3 {
		t.Errorf("Expected length to be 3 after multiple pushes beyond capacity, got %d", hwm.Len())
	}

	if popped := hwm.Pop(); popped != "H" {
		t.Errorf("Expected to pop 'H', got '%s'", popped)
	}
	if popped := hwm.Pop(); popped != "G" {
		t.Errorf("Expected to pop 'G', got '%s'", popped)
	}
	if popped := hwm.Pop(); popped != "F" {
		t.Errorf("Expected to pop 'F', got '%s'", popped)
	}
}

func TestHighWaterMark_EdgeCases(t *testing.T) {
	hwm := NewHighWaterMark(2)

	if popped := hwm.Pop(); popped != "" {
		t.Errorf("Expected empty string when popping from empty stack, got '%s'", popped)
	}

	hwm.Push("single")
	if hwm.Len() != 1 {
		t.Errorf("Expected length 1 with single item, got %d", hwm.Len())
	}

	if top := hwm.Top(); top != "single" {
		t.Errorf("Expected top to be 'single', got '%s'", top)
	}

	if popped := hwm.Pop(); popped != "single" {
		t.Errorf("Expected to pop 'single', got '%s'", popped)
	}

	if hwm.Len() != 0 {
		t.Errorf("Expected length 0 after popping single item, got %d", hwm.Len())
	}

	hwm2 := NewHighWaterMark(1)
	hwm2.Push("first")
	hwm2.Push("second")

	if hwm2.Len() != 1 {
		t.Errorf("Expected length 1 with capacity 1, got %d", hwm2.Len())
	}

	if top := hwm2.Top(); top != "second" {
		t.Errorf("Expected top to be 'second' (most recent), got '%s'", top)
	}

	if popped := hwm2.Pop(); popped != "second" {
		t.Errorf("Expected to pop 'second', got '%s'", popped)
	}
}

func TestHighWaterMark_ConstructorWithInitialItems(t *testing.T) {
	hwm := NewHighWaterMark(5, "X", "Y", "Z")

	if hwm.Len() != 3 {
		t.Errorf("Expected length 3 with initial items, got %d", hwm.Len())
	}

	if top := hwm.Top(); top != "Z" {
		t.Errorf("Expected top to be 'Z' (last initial item), got '%s'", top)
	}

	if popped := hwm.Pop(); popped != "Z" {
		t.Errorf("Expected to pop 'Z', got '%s'", popped)
	}
	if popped := hwm.Pop(); popped != "Y" {
		t.Errorf("Expected to pop 'Y', got '%s'", popped)
	}
	if popped := hwm.Pop(); popped != "X" {
		t.Errorf("Expected to pop 'X', got '%s'", popped)
	}

	hwm2 := NewHighWaterMark(2, "A", "B", "C", "D")

	if hwm2.Len() != 4 {
		t.Errorf("Expected length 4 (all initial items stored even exceeding capacity), got %d", hwm2.Len())
	}

	hwm2.Push("E")

	if hwm2.Len() != 2 {
		t.Errorf("Expected length 2 after push (capacity enforced), got %d", hwm2.Len())
	}

	if popped := hwm2.Pop(); popped != "E" {
		t.Errorf("Expected to pop 'E' (most recent), got '%s'", popped)
	}
	if popped := hwm2.Pop(); popped != "D" {
		t.Errorf("Expected to pop 'D' (kept after capacity enforcement), got '%s'", popped)
	}
}
