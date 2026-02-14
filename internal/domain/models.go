package domain

// DocumentRequest - Data sent to the AI
type DocumentRequest struct {
	DocID    string  `json:"doc_id"`
	Metadata Meta    `json:"metadata"` // Author, date, etc.
	Content  []Block `json:"content"`  // List of paragraphs, tables, formulas
}

// Meta - Metadata for the document
type Meta struct {
    Author string `json:"author"`
    Date   string `json:"date"`
}

// Block - A single entity within the document
type Block struct {
	ID      string `json:"id"`      // IMPORTANT: Unique ID (e.g., "p_105")
	Type    string `json:"type"`    // "paragraph", "heading", "table", "formula"
	Text    string `json:"text"`    // Plain text content
	Context string `json:"context"` // Extra info (e.g., "Bibliography Item")
}

// PatchResponse - Data returned by the AI
type PatchResponse struct {
	Changes []ChangeOp `json:"changes"`
}

// ChangeOp - Specific instruction for modification
type ChangeOp struct {
	TargetID  string `json:"target_id"` // ID of the block to change (e.g., "p_105")
	Operation string `json:"op"`        // "replace_text", "delete", "insert_after"
	NewText   string `json:"new_text,omitempty"`
	Comment   string `json:"comment,omitempty"` // AI's reasoning for the change
}
