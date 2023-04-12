package elevenlabs

type VoicesResponse struct {
	Voices []Voices `json:"voices"`
}
type Samples struct {
	SampleID  string `json:"sample_id"`
	FileName  string `json:"file_name"`
	MimeType  string `json:"mime_type"`
	SizeBytes int    `json:"size_bytes"`
	Hash      string `json:"hash"`
}
type Recording struct {
	RecordingID    string `json:"recording_id"`
	MimeType       string `json:"mime_type"`
	SizeBytes      int    `json:"size_bytes"`
	UploadDateUnix int    `json:"upload_date_unix"`
	Transcription  string `json:"transcription"`
}
type VerificationAttempts struct {
	Text                string    `json:"text"`
	DateUnix            int       `json:"date_unix"`
	Accepted            bool      `json:"accepted"`
	Similarity          int       `json:"similarity"`
	LevenshteinDistance int       `json:"levenshtein_distance"`
	Recording           Recording `json:"recording"`
}
type FineTuning struct {
	ModelID                   string                 `json:"model_id"`
	IsAllowedToFineTune       bool                   `json:"is_allowed_to_fine_tune"`
	FineTuningRequested       bool                   `json:"fine_tuning_requested"`
	FinetuningState           string                 `json:"finetuning_state"`
	VerificationAttempts      []VerificationAttempts `json:"verification_attempts"`
	VerificationFailures      []string               `json:"verification_failures"`
	VerificationAttemptsCount int                    `json:"verification_attempts_count"`
	SliceIds                  []string               `json:"slice_ids"`
}
type Labels struct {
	AdditionalProp1 string `json:"additionalProp1"`
	AdditionalProp2 string `json:"additionalProp2"`
	AdditionalProp3 string `json:"additionalProp3"`
}
type Settings struct {
	Stability       int `json:"stability"`
	SimilarityBoost int `json:"similarity_boost"`
}
type Voices struct {
	VoiceID           string     `json:"voice_id"`
	Name              string     `json:"name"`
	Samples           []Samples  `json:"samples"`
	Category          string     `json:"category"`
	FineTuning        FineTuning `json:"fine_tuning"`
	Labels            Labels     `json:"labels"`
	Description       string     `json:"description"`
	PreviewURL        string     `json:"preview_url"`
	AvailableForTiers []string   `json:"available_for_tiers"`
	Settings          Settings   `json:"settings"`
}
