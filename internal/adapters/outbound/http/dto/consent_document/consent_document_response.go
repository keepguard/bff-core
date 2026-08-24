package dto

// ConsentDocumentResponseDTO representa a resposta de documento de consentimento
type ConsentDocumentResponseDTO struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Type          string `json:"type"`
	Version       int    `json:"version"`
	Status        string `json:"status"`
	S3URL         string `json:"s3Url"`
	ContentHash   string `json:"contentHash"`
	FileSizeBytes int64  `json:"fileSizeBytes"`
	MimeType      string `json:"mimeType"`
	CreatedBy     string `json:"createdBy"`
	CreatedAt     string `json:"createdAt"`
	UpdatedBy     string `json:"updatedBy,omitempty"`
	PublishedAt   string `json:"publishedAt,omitempty"`
}
