package domain

import "sync"

type DeltaType string

const (
	Insert DeltaType = "INSERT"
	Delete DeltaType = "DELETE"
)

// Delta represents a single atomic change made to the document.
// It describes what was added, where, and by whom.
type Delta struct {
	UserID   string    `json:"user_id"`  // UserID identifies the author of the change.
	Type     DeltaType `json:"type"`     // Could be either insertion or deletion
	Position int       `json:"position"` // Position is the character-based (rune) index where the content is inserted.
	Content  string    `json:"content"`  // Content is the string sequence to be inserted at the given position.
	Length   int       `json:"length"`   // Length of the deleted text starting at position
}

// Document is the Aggregate Root representing the text file being edited.
// It ensures thread-safe operations on the text state.
type Document struct {
	Mutex sync.RWMutex

	ID      string // ID is the unique identifier for the document instance.
	Content string // Content holds the current state of the text as a UTF-8 string.
	Version int
}

// NewDocument initializes a new Document with the given unique identifier and an initial content.
func NewDocument(id, initialContent string) *Document {
	return &Document{ID: id, Content: initialContent}
}

// Transition performs a state transition on the document by inserting or deleting a Delta.
// It handles UTF-8 characters safely by using rune slices and clamps the
// position to valid bounds using built-in min/max functions.
// It returns the updated state of the document.
func (d *Document) Transition(delta *Delta) string {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()

	// Convert string to runes to handle multi-byte characters correctly (e.g., emojis)
	runes := []rune(d.Content)

	// Clamp the position safely between 0 and the current length of the document.
	// This prevents index-out-of-range panics from malformed client input.
	position := max(0, min(delta.Position, len(runes)))

	// Reconstruct the document content
	switch delta.Type {
	case Insert:
		insertion := []rune(delta.Content)
		d.Content = string(append(runes[:position], append(insertion, runes[position:]...)...))
	case Delete:
		// Ensure we don't delete past the end of the document
		end := max(position, min(position+delta.Length, len(runes)))
		d.Content = string(append(runes[:position], runes[end:]...))
	}

	d.Version++
	return d.Content
}
