package dto

// ConsentDocumentResponseDTO representa a resposta de documento de consentimento
type ConsentDocumentResponseDTO struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Version     int    `json:"version"`
	Status      string `json:"status"`
	ContentURL  string `json:"contentUrl"`
	FileSize    int64  `json:"fileSize"`
	MimeType    string `json:"mimeType"`
	CreatedBy   string `json:"createdBy"`
	CreatedAt   string `json:"createdAt"`
	UpdatedBy   string `json:"updatedBy,omitempty"`
	UpdatedAt   string `json:"updatedAt,omitempty"`
	PublishedAt string `json:"publishedAt,omitempty"`
	ArchivedAt  string `json:"archivedAt,omitempty"`
}
