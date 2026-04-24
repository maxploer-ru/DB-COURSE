package dto

type CreateVideoRequest struct {
	ChannelID   int    `json:"channel_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	FileKey     string `json:"file_key"`
}

type UpdateVideoRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
}

type VideoResponse struct {
	ID          int    `json:"id"`
	ChannelID   int    `json:"channel_id"`
	ChannelName string `json:"channel_name"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Views       int    `json:"views"`
	Likes       int    `json:"likes"`
	Dislikes    int    `json:"dislikes"`
	Comments    int    `json:"comments"`
	CreatedAt   string `json:"created_at"`
}

type UploadPresignedURLRequest struct {
	ChannelID int    `json:"channel_id"`
	Filename  string `json:"filename"`
}

type UploadPresignedURLResponse struct {
	URL     string `json:"url"`
	FileKey string `json:"file_key"`
}

type StreamingURLResponse struct {
	URL string `json:"url"`
}
