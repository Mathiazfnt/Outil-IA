package api

import (
	"fmt"
	"reflect"
	"testing"

	"exo_API/internal/pokemon"
)

func TestWeaknessMultipliersCombineDefendingTypes(t *testing.T) {
	relations := map[string]typeDamageRelations{
		"fire": {
			DoubleDamageFrom: []pokemon.NamedAPIResource{
				{Name: "water"},
				{Name: "ground"},
				{Name: "rock"},
			},
			HalfDamageFrom: []pokemon.NamedAPIResource{{Name: "ice"}},
		},
		"flying": {
			DoubleDamageFrom: []pokemon.NamedAPIResource{
				{Name: "electric"},
				{Name: "ice"},
				{Name: "rock"},
			},
			NoDamageFrom: []pokemon.NamedAPIResource{{Name: "ground"}},
		},
	}
	types := []pokemon.TypeSlot{
		{Type: pokemon.NamedAPIResource{Name: "fire"}},
		{Type: pokemon.NamedAPIResource{Name: "flying"}},
	}

	got, err := calculerFaiblesses(types, func(name string) (typeDamageRelations, error) {
		result, ok := relations[name]
		if !ok {
			return typeDamageRelations{}, fmt.Errorf("type inattendu : %s", name)
		}
		return result, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []pokemon.Faiblesse{
		{Nom: "electric", Multiplicateur: 2},
		{Nom: "rock", Multiplicateur: 4},
		{Nom: "water", Multiplicateur: 2},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("calculerFaiblesses() = %#v, want %#v", got, want)
	}
}
