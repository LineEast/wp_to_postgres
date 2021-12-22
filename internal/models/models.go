package models

type (
	Post struct {
		ID 			 uint 	`json:"id"`
		OldID        uint   `json:"oldId"`
		Author       uint   `json:"author"`
		Date         string `json:"date"`
		Content      string `json:"content"`
		ContentShort string `json:"ContentShort"`
		Title        string `json:"title"`
		Image        string `json:"image"`
		TagsID       []uint `json:"tagsID"`

		Views uint `json:"PostViewsCount"`
	}

	Tag struct {
		ID    		uint64 	`json:"id"`
		Name  		string 	`json:"name"`
		Alias 		string 	`json:"alias"`
		Type  		string 	`json:"type"`
	}
)