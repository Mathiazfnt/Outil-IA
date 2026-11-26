package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"time"

	"exo_API/internal/pokemon"
)

const endpoint = "https://pokeapi.co/api/v2/pokemon"
const typeEndpoint = "https://pokeapi.co/api/v2/type"

var client = &http.Client{Timeout: 10 * time.Second}

type typeDamageRelations struct {
	DoubleDamageFrom []pokemon.NamedAPIResource `json:"double_damage_from"`
	HalfDamageFrom   []pokemon.NamedAPIResource `json:"half_damage_from"`
	NoDamageFrom     []pokemon.NamedAPIResource `json:"no_damage_from"`
}

type typeResponse struct {
	DamageRelations typeDamageRelations `json:"damage_relations"`
}

func Recuperer_Pokemon(nom string) (pokemon.Choix, string, error) {
	var resultat pokemon.Choix
	var resp *http.Response
	var err error

	resp, err = client.Get(endpoint + "/" + nom)
	if err != nil {
		return resultat, "", fmt.Errorf("Erreur de connexion : %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return resultat, resp.Status, fmt.Errorf("Erreur : Pokémon introuvable ou erreur du serveur : %s", resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(&resultat)
	if err != nil {
		return resultat, resp.Status, fmt.Errorf("Erreur JSON : %w", err)
	}

	resultat.Faiblesses, err = calculerFaiblesses(resultat.Types, recupererRelationsType)
	if err != nil {
		return resultat, resp.Status, fmt.Errorf("Erreur lors du chargement des faiblesses : %w", err)
	}

	return resultat, resp.Status, nil
}

func Recuperer_Noms_Pokemon() ([]string, error) {
	var resultat struct {
		Results []struct {
			Name string `json:"name"`
		} `json:"results"`
	}

	resp, err := client.Get(endpoint + "?limit=1328")
	if err != nil {
		return nil, fmt.Errorf("Erreur de connexion : %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Erreur lors du chargement des noms : %s", resp.Status)
	}
	if err := json.NewDecoder(resp.Body).Decode(&resultat); err != nil {
		return nil, fmt.Errorf("Erreur JSON : %w", err)
	}
	if len(resultat.Results) == 0 {
		return nil, fmt.Errorf("PokéAPI n’a retourné aucun nom de Pokémon")
	}

	noms := make([]string, 0, len(resultat.Results))
	for _, pokemon := range resultat.Results {
		noms = append(noms, pokemon.Name)
	}
	return noms, nil
}

func recupererRelationsType(nom string) (typeDamageRelations, error) {
	var resultat typeResponse
	resp, err := client.Get(typeEndpoint + "/" + url.PathEscape(nom))
	if err != nil {
		return resultat.DamageRelations, fmt.Errorf("Erreur de connexion : %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return resultat.DamageRelations, fmt.Errorf("Erreur lors du chargement du type %q : %s", nom, resp.Status)
	}
	if err := json.NewDecoder(resp.Body).Decode(&resultat); err != nil {
		return resultat.DamageRelations, fmt.Errorf("Erreur JSON pour le type %q : %w", nom, err)
	}
	return resultat.DamageRelations, nil
}

func calculerFaiblesses(types []pokemon.TypeSlot, chargerRelations func(string) (typeDamageRelations, error)) ([]pokemon.Faiblesse, error) {
	efficacite := make(map[string]float64)
	for _, typePokemon := range types {
		if typePokemon.Type.Name == "" {
			return nil, fmt.Errorf("un type Pokémon n’a pas de nom")
		}

		relations, err := chargerRelations(typePokemon.Type.Name)
		if err != nil {
			return nil, err
		}
		appliquerMultiplicateur(efficacite, relations.DoubleDamageFrom, 2)
		appliquerMultiplicateur(efficacite, relations.HalfDamageFrom, 0.5)
		appliquerMultiplicateur(efficacite, relations.NoDamageFrom, 0)
	}

	faiblesses := make([]pokemon.Faiblesse, 0, len(efficacite))
	for nom, multiplicateur := range efficacite {
		if multiplicateur > 1 {
			faiblesses = append(faiblesses, pokemon.Faiblesse{
				Nom:            nom,
				Multiplicateur: multiplicateur,
			})
		}
	}
	sort.Slice(faiblesses, func(i, j int) bool {
		return faiblesses[i].Nom < faiblesses[j].Nom
	})
	return faiblesses, nil
}

func appliquerMultiplicateur(efficacite map[string]float64, types []pokemon.NamedAPIResource, multiplicateur float64) {
	for _, typeAttaque := range types {
		if _, ok := efficacite[typeAttaque.Name]; !ok {
			efficacite[typeAttaque.Name] = 1
		}
		efficacite[typeAttaque.Name] *= multiplicateur
	}
}
