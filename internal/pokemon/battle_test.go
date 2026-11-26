package pokemon

import "testing"

func TestComparerCombatUsesNumberOfHigherStats(t *testing.T) {
	rouge := Choix{Stats: []BaseStat{
		{Stat: NamedAPIResource{Name: "hp"}, BaseStat: 100},
		{Stat: NamedAPIResource{Name: "attack"}, BaseStat: 80},
		{Stat: NamedAPIResource{Name: "defense"}, BaseStat: 70},
	}}
	bleu := Choix{Stats: []BaseStat{
		{Stat: NamedAPIResource{Name: "hp"}, BaseStat: 90},
		{Stat: NamedAPIResource{Name: "attack"}, BaseStat: 80},
		{Stat: NamedAPIResource{Name: "defense"}, BaseStat: 60},
	}}

	resultat := ComparerCombat(rouge, bleu)
	if resultat.Gagnant != GagnantRouge {
		t.Fatalf("winner = %q, want %q", resultat.Gagnant, GagnantRouge)
	}
	if resultat.ScoreRouge != 2 || resultat.ScoreBleu != 0 {
		t.Errorf("scores = %d-%d, want 2-0", resultat.ScoreRouge, resultat.ScoreBleu)
	}
	if resultat.Statistiques[1].Gagnant != Egalite {
		t.Errorf("tied stat winner = %q, want %q", resultat.Statistiques[1].Gagnant, Egalite)
	}
}

func TestComparerCombatReturnsDrawWhenScoresMatch(t *testing.T) {
	rouge := Choix{Stats: []BaseStat{
		{Stat: NamedAPIResource{Name: "hp"}, BaseStat: 100},
		{Stat: NamedAPIResource{Name: "attack"}, BaseStat: 70},
	}}
	bleu := Choix{Stats: []BaseStat{
		{Stat: NamedAPIResource{Name: "hp"}, BaseStat: 90},
		{Stat: NamedAPIResource{Name: "attack"}, BaseStat: 80},
	}}

	resultat := ComparerCombat(rouge, bleu)
	if resultat.Gagnant != Egalite {
		t.Fatalf("winner = %q, want %q", resultat.Gagnant, Egalite)
	}
}

func TestComparerCombatDoesNotScoreMissingStats(t *testing.T) {
	rouge := Choix{Stats: []BaseStat{
		{Stat: NamedAPIResource{Name: "hp"}, BaseStat: 100},
		{Stat: NamedAPIResource{Name: "attack"}, BaseStat: 70},
	}}
	bleu := Choix{Stats: []BaseStat{
		{Stat: NamedAPIResource{Name: "hp"}, BaseStat: 90},
	}}

	resultat := ComparerCombat(rouge, bleu)
	if resultat.ScoreRouge != 1 || resultat.ScoreBleu != 0 {
		t.Errorf("scores = %d-%d, want 1-0", resultat.ScoreRouge, resultat.ScoreBleu)
	}
	if resultat.Statistiques[1].Disponible {
		t.Error("stat missing from blue Pokémon should not be comparable")
	}
}

func TestComparerCombatDoublesPointsForTypeAdvantage(t *testing.T) {
	rouge := Choix{
		Types:      []TypeSlot{{Type: NamedAPIResource{Name: "electric"}}},
		Faiblesses: []Faiblesse{{Nom: "water", Multiplicateur: 2}},
		Stats: []BaseStat{
			{Stat: NamedAPIResource{Name: "hp"}, BaseStat: 100},
			{Stat: NamedAPIResource{Name: "attack"}, BaseStat: 50},
		},
	}
	bleu := Choix{
		Types:      []TypeSlot{{Type: NamedAPIResource{Name: "water"}}},
		Faiblesses: []Faiblesse{{Nom: "electric", Multiplicateur: 4}},
		Stats: []BaseStat{
			{Stat: NamedAPIResource{Name: "hp"}, BaseStat: 90},
			{Stat: NamedAPIResource{Name: "attack"}, BaseStat: 60},
		},
	}

	resultat := ComparerCombat(rouge, bleu)
	if resultat.MultiplicateurTypeRouge != 4 {
		t.Errorf("red type multiplier = %g, want 4", resultat.MultiplicateurTypeRouge)
	}
	if resultat.MultiplicateurTypeBleu != 2 {
		t.Errorf("blue type multiplier = %g, want 2", resultat.MultiplicateurTypeBleu)
	}
	if resultat.PointsParStatRouge != 2 || resultat.PointsParStatBleu != 1 {
		t.Errorf("points per stat = %d-%d, want 2-1", resultat.PointsParStatRouge, resultat.PointsParStatBleu)
	}
	if resultat.ScoreRouge != 2 || resultat.ScoreBleu != 1 || resultat.Gagnant != GagnantRouge {
		t.Errorf("result = %d-%d, winner %q; want 2-1, %q", resultat.ScoreRouge, resultat.ScoreBleu, resultat.Gagnant, GagnantRouge)
	}
}

func TestComparerCombatIncludesHeightAndWeight(t *testing.T) {
	rouge := Choix{Taille: 20, Poids: 150}
	bleu := Choix{Taille: 10, Poids: 200}

	resultat := ComparerCombat(rouge, bleu)
	if resultat.Statistiques[6].Nom != "Taille" || resultat.Statistiques[6].Gagnant != GagnantRouge {
		t.Errorf("height comparison = %#v, want red winner", resultat.Statistiques[6])
	}
	if resultat.Statistiques[7].Nom != "Poids" || resultat.Statistiques[7].Gagnant != GagnantBleu {
		t.Errorf("weight comparison = %#v, want blue winner", resultat.Statistiques[7])
	}
}
