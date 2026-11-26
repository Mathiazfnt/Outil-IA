package web

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"exo_API/internal/api"
	"exo_API/internal/pokemon"
)

//go:embed templates/*.html static/*.css
var webFiles embed.FS

type pageData struct {
	SearchTerm      string
	Pokemon         *pokemon.Choix
	Error           string
	Suggestions     []string
	SuggestionError string
	RedSearchTerm   string
	BlueSearchTerm  string
	RedPokemon      *pokemon.Choix
	BluePokemon     *pokemon.Choix
	Battle          *pokemon.ResultatCombat
}

type server struct {
	page             *template.Template
	combatPage       *template.Template
	stylesheet       []byte
	combatStylesheet []byte
	loadSuggestions  func() ([]string, error)
	loadPokemon      func(string) (pokemon.Choix, error)
	suggestionsMu    sync.Mutex
	suggestions      []string
}

var validIdentifier = regexp.MustCompile(`^[a-zA-Z0-9-]{1,50}$`)

func Serve() error {
	handler, err := newHandler()
	if err != nil {
		return err
	}

	log.Println("Site disponible sur http://localhost:8080")
	return http.ListenAndServe("127.0.0.1:8080", handler)
}

func newHandler() (http.Handler, error) {
	return newHandlerWithLoaders(api.Recuperer_Noms_Pokemon, func(name string) (pokemon.Choix, error) {
		result, _, err := api.Recuperer_Pokemon(name)
		return result, err
	})
}

func newHandlerWithLoaders(loadSuggestions func() ([]string, error), loadPokemon func(string) (pokemon.Choix, error)) (http.Handler, error) {
	page, err := template.ParseFS(webFiles, "templates/index.html", "templates/pokemon-info.html")
	if err != nil {
		return nil, fmt.Errorf("chargement de index.html : %w", err)
	}
	combatPage, err := template.ParseFS(webFiles, "templates/combat.html", "templates/pokemon-info.html")
	if err != nil {
		return nil, fmt.Errorf("chargement de combat.html : %w", err)
	}
	stylesheet, err := webFiles.ReadFile("static/style.css")
	if err != nil {
		return nil, fmt.Errorf("chargement de style.css : %w", err)
	}
	combatStylesheet, err := webFiles.ReadFile("static/combat.css")
	if err != nil {
		return nil, fmt.Errorf("chargement de combat.css : %w", err)
	}

	app := &server{
		page:             page,
		combatPage:       combatPage,
		stylesheet:       stylesheet,
		combatStylesheet: combatStylesheet,
		loadSuggestions:  loadSuggestions,
		loadPokemon:      loadPokemon,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/style.css", app.style)
	mux.HandleFunc("/pokemon", app.search)
	mux.HandleFunc("/combat", app.combat)
	mux.HandleFunc("/combat.css", app.combatStyle)
	return mux, nil
}

func (s *server) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	data := pageData{}
	s.addSuggestions(&data)
	s.render(w, http.StatusOK, data)
}

func (s *server) style(w http.ResponseWriter, r *http.Request) {
	s.serveStylesheet(w, r, s.stylesheet)
}

func (s *server) combatStyle(w http.ResponseWriter, r *http.Request) {
	s.serveStylesheet(w, r, s.combatStylesheet)
}

func (s *server) serveStylesheet(w http.ResponseWriter, r *http.Request, stylesheet []byte) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	_, _ = w.Write(stylesheet)
}

func (s *server) search(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	searchTerm := strings.TrimSpace(r.URL.Query().Get("name"))
	data := pageData{SearchTerm: searchTerm}
	s.addSuggestions(&data)
	if !validIdentifier.MatchString(searchTerm) {
		data.Error = "Saisissez un nom ou un identifiant valide."
		s.render(w, http.StatusBadRequest, data)
		return
	}

	result, err := s.loadPokemon(strings.ToLower(searchTerm))
	if err != nil {
		log.Printf("Recherche du Pokémon %q impossible : %v", searchTerm, err)
		data.Error = fmt.Sprintf("Impossible de récupérer ce Pokémon : %v", err)
		s.render(w, http.StatusBadGateway, data)
		return
	}

	data.Pokemon = &result
	s.render(w, http.StatusOK, data)
}

func (s *server) combat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	data := pageData{
		RedSearchTerm:  strings.TrimSpace(r.URL.Query().Get("red")),
		BlueSearchTerm: strings.TrimSpace(r.URL.Query().Get("blue")),
	}
	s.addSuggestions(&data)
	for _, selection := range []struct {
		term   string
		result **pokemon.Choix
	}{
		{term: data.RedSearchTerm, result: &data.RedPokemon},
		{term: data.BlueSearchTerm, result: &data.BluePokemon},
	} {
		if selection.term == "" {
			continue
		}
		if !validIdentifier.MatchString(selection.term) {
			data.Error = fmt.Sprintf("Le nom ou l’identifiant %q est invalide.", selection.term)
			s.renderTemplate(w, http.StatusBadRequest, data, s.combatPage)
			return
		}
		result, err := s.loadPokemon(strings.ToLower(selection.term))
		if err != nil {
			log.Printf("Recherche du Pokémon %q impossible : %v", selection.term, err)
			data.Error = fmt.Sprintf("Impossible de récupérer %q : %v", selection.term, err)
			s.renderTemplate(w, http.StatusBadGateway, data, s.combatPage)
			return
		}
		*selection.result = &result
	}
	if r.URL.Query().Get("battle") == "1" {
		if data.RedPokemon == nil || data.BluePokemon == nil {
			data.Error = "Choisissez un Pokémon pour chaque équipe avant le combat."
			s.renderTemplate(w, http.StatusBadRequest, data, s.combatPage)
			return
		}
		resultat := pokemon.ComparerCombat(*data.RedPokemon, *data.BluePokemon)
		data.Battle = &resultat
	}
	s.renderTemplate(w, http.StatusOK, data, s.combatPage)
}

func (s *server) addSuggestions(data *pageData) {
	s.suggestionsMu.Lock()
	defer s.suggestionsMu.Unlock()

	if s.suggestions == nil {
		suggestions, err := s.loadSuggestions()
		if err != nil {
			log.Printf("Chargement des suggestions impossible : %v", err)
			data.SuggestionError = "Les suggestions sont momentanément indisponibles."
			return
		}
		s.suggestions = suggestions
	}
	data.Suggestions = s.suggestions
}

func (s *server) render(w http.ResponseWriter, status int, data pageData) {
	s.renderTemplate(w, status, data, s.page)
}

func (s *server) renderTemplate(w http.ResponseWriter, status int, data pageData, page *template.Template) {
	var content bytes.Buffer
	if err := page.Execute(&content, data); err != nil {
		log.Printf("Rendu de la page impossible : %v", err)
		http.Error(w, "Erreur lors de l’affichage de la page.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = content.WriteTo(w)
}
