package entity

type UserProfile struct {
	GenderPref  string   `json:"gender_pref" db:"gender_pref"`
	Styles      []string `json:"styles" db:"styles"`
	Colors      []string `json:"colors" db:"colors"`
	MusicGenres []string `json:"music_genres" db:"music_genres"`
}

type RecommendationContext struct {
	Season      string   `json:"season"`
	Formality   int      `json:"formality"`
	Styles      []string `json:"styles"`
	Colors      []string `json:"colors"`
	MusicGenres []string `json:"music_genres"`
	Gender      string   `json:"gender"`
}

type RequestData struct {
	UserProfile UserProfile           `json:"user_profile" db:"user_profile"`
	Context     RecommendationContext `json:"context" db:"context"`
	K           int                   `json:"k" db:"k"`
}

type OutfitItem struct {
	ItemID   string  `json:"item_id"`
	Score    float64 `json:"score"`
	Category string  `json:"category"`
	Material string  `json:"material"`
}

type RecommendationRequest struct {
	Formality int `json:"formality" validate:"required,min=1,max=5"`
}
