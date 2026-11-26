package pokemon

type NamedAPIResource struct {
	Name string `json:"name"`
}

type TypeSlot struct {
	Slot int              `json:"slot"`
	Type NamedAPIResource `json:"type"`
}

type BaseStat struct {
	BaseStat int              `json:"base_stat"`
	Stat     NamedAPIResource `json:"stat"`
}

type Faiblesse struct {
	Nom            string
	Multiplicateur float64
}

type Choix struct {
	Nom   string `json:"name"`
	ID    int    `json:"id"`
	Image struct {
		FrontDefault string `json:"front_default"`
	} `json:"sprites"`
	Taille int `json:"height"`
	Poids  int `json:"weight"`

	Types      []TypeSlot  `json:"types"`
	Stats      []BaseStat  `json:"stats"`
	Faiblesses []Faiblesse `json:"-"`
}
