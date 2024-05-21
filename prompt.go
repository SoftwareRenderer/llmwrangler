package main

type ImageData struct {
	Id   int    `json:"id"`
	Data string `json:"data"`
}
type Completion struct {
	Cache_prompt      bool        `json:"cache_prompt"`
	Frequency_penalty float32     `json:"frequency_penalty"`
	Grammar           string      `json:"grammar"`
	Image_data        []ImageData `json:"image_data"`
	Mirostat          int         `json:"mirostat"`
	Mirostat_eta      float32     `json:"mirostat_eta"`
	Mirostat_tau      float32     `json:"mirostat_tau"`
	N_predict         int         `json:"n_predict"`
	N_probs           int         `json:"n_probs"`
	Presence_penalty  float32     `json:"presence_penalty"`
	Prompt            string      `json:"prompt"`
	Repeat_last_n     int         `json:"repeat_last_n"`
	Repeat_penalty    float32     `json:"repeat_penalty"`
	Slot_id           int         `json:"slot_id"`
	Stopwords         []string    `json:"stop"`
	Stream            bool        `json:"stream"`
	Temperature       float32     `json:"temperature"`
	Tfs_z             float32     `json:"tfs_z"`
	Top_k             int         `json:"top_k"`
	Top_p             float32     `json:"top_p"`
	Typical_p         float32     `json:"typical_p"`
	Min_p             float32     `json:"min_p"`
}
