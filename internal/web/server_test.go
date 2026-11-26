package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"exo_API/internal/pokemon"
)

func testHandler(t *testing.T) http.Handler {
	t.Helper()

	handler, err := newHandlerWithLoaders(
		func() ([]string, error) {
			return []string{"bulbasaur", "pikachu"}, nil
		},
		func(name string) (pokemon.Choix, error) {
			id := 25
			if name == "bulbasaur" {
				id = 1
			}
			stats := []pokemon.BaseStat{
				{Stat: pokemon.NamedAPIResource{Name: "hp"}, BaseStat: 100},
				{Stat: pokemon.NamedAPIResource{Name: "attack"}, BaseStat: 80},
				{Stat: pokemon.NamedAPIResource{Name: "defense"}, BaseStat: 90},
			}
			if name == "bulbasaur" {
				stats[0].BaseStat = 50
				stats[1].BaseStat = 90
				stats[2].BaseStat = 80
			}
			return pokemon.Choix{
				Nom:   name,
				ID:    id,
				Stats: stats,
				Faiblesses: []pokemon.Faiblesse{
					{Nom: "water", Multiplicateur: 2},
				},
			}, nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	return handler
}

func TestHomeServesSearchPage(t *testing.T) {
	handler := testHandler(t)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want %d", response.Code, http.StatusOK)
	}
	body := response.Body.String()
	for _, expected := range []string{`action="/pokemon"`, `name="name"`, `list="pokemon-names"`, `<option value="bulbasaur">`, `href="style.css"`, `href="/combat"`} {
		if !strings.Contains(body, expected) {
			t.Errorf("GET / body does not contain %q", expected)
		}
	}
	for _, removed := range []string{`name="limit"`, `name="offset"`, `name="view"`, "Liste des Pokémon", "pokeapi.co/api/v2/pokemon?"} {
		if strings.Contains(body, removed) {
			t.Errorf("GET / body still contains removed list/search control %q", removed)
		}
	}
	if strings.Contains(body, "au hasard") {
		t.Error("home page still contains the removed random Pokémon feature")
	}
}

func TestCombatPageHasTwoSearchInputsAndVSButton(t *testing.T) {
	handler := testHandler(t)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/combat", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("GET /combat status = %d, want %d", response.Code, http.StatusOK)
	}
	body := response.Body.String()
	if count := strings.Count(body, `list="pokemon-names"`); count != 2 {
		t.Errorf("combat page has %d autocomplete inputs, want 2", count)
	}
	for _, expected := range []string{`id="red-pokemon"`, `id="blue-pokemon"`, `>VS</button>`, `<option value="pikachu">`, `href="/combat.css"`} {
		if !strings.Contains(body, expected) {
			t.Errorf("combat page does not contain %q", expected)
		}
	}
}

func TestSearchRejectsInvalidIdentifier(t *testing.T) {
	handler := testHandler(t)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/pokemon?name=not%2Fa%2Fpokemon", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("invalid search status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	if !strings.Contains(response.Body.String(), "Saisissez un nom ou un identifiant valide.") {
		t.Error("invalid search response does not explain the input error")
	}
}

func TestPokemonSearchDisplaysWeaknesses(t *testing.T) {
	handler := testHandler(t)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/pokemon?name=pikachu", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("GET /pokemon status = %d, want %d", response.Code, http.StatusOK)
	}
	if !strings.Contains(response.Body.String(), "water : 2 fois les dégâts normaux") {
		t.Error("Pokémon fiche does not show the type weakness")
	}
}

func TestCombatDisplaysDetailsForBothTeams(t *testing.T) {
	handler := testHandler(t)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/combat?red=pikachu&blue=bulbasaur", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("GET /combat status = %d, want %d", response.Code, http.StatusOK)
	}
	body := response.Body.String()
	for _, expected := range []string{"pikachu — ID 25", "bulbasaur — ID 1", "water : 2 fois les dégâts normaux"} {
		if !strings.Contains(body, expected) {
			t.Errorf("combat page does not contain %q", expected)
		}
	}
	if count := strings.Count(body, "Faiblesses de type"); count != 2 {
		t.Errorf("combat page has %d weakness sections, want one for each team", count)
	}
}

func TestCombatComparesStatsAndDisplaysWinner(t *testing.T) {
	handler := testHandler(t)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/combat?red=pikachu&blue=bulbasaur&battle=1", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("GET /combat battle status = %d, want %d", response.Code, http.StatusOK)
	}
	body := response.Body.String()
	for _, expected := range []string{
		"pikachu contre bulbasaur",
		"Score : pikachu 2 — bulbasaur 1",
		"pikachu gagne le combat !",
		"Attaque spéciale",
		"1 point par statistique gagnée",
		"Taille",
		"Poids",
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("battle response does not contain %q", expected)
		}
	}
}

func TestCombatRequiresBothPokemonsBeforeComparison(t *testing.T) {
	handler := testHandler(t)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/combat?red=pikachu&battle=1", nil)
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("incomplete battle status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	if !strings.Contains(response.Body.String(), "Choisissez un Pokémon pour chaque équipe") {
		t.Error("incomplete battle response does not explain the missing choice")
	}
}

func TestStyleServesCSS(t *testing.T) {
	handler := testHandler(t)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/style.css", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("GET /style.css status = %d, want %d", response.Code, http.StatusOK)
	}
	if contentType := response.Header().Get("Content-Type"); contentType != "text/css; charset=utf-8" {
		t.Errorf("CSS content type = %q", contentType)
	}
}
