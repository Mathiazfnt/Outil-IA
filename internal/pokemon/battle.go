package pokemon

const (
	GagnantRouge        = "red"
	GagnantBleu         = "blue"
	Egalite             = "draw"
	ComparaisonInconnue = "unknown"
)

type ComparaisonStat struct {
	Nom         string
	Unite       string
	Rouge       int
	Bleu        int
	Gagnant     string
	PointsRouge int
	PointsBleu  int
	Disponible  bool
}

type ResultatCombat struct {
	Statistiques            []ComparaisonStat
	ScoreRouge              int
	ScoreBleu               int
	MultiplicateurTypeRouge float64
	MultiplicateurTypeBleu  float64
	PointsParStatRouge      int
	PointsParStatBleu       int
	Gagnant                 string
}

func ComparerCombat(rouge, bleu Choix) ResultatCombat {
	statsRouges := valeursStats(rouge.Stats)
	statsBleues := valeursStats(bleu.Stats)
	statsRouges["height"] = rouge.Taille
	statsRouges["weight"] = rouge.Poids
	statsBleues["height"] = bleu.Taille
	statsBleues["weight"] = bleu.Poids
	nomsStats := ordreStats(rouge.Stats, bleu.Stats)
	multiplicateurRouge := multiplicateurType(rouge, bleu)
	multiplicateurBleu := multiplicateurType(bleu, rouge)

	resultat := ResultatCombat{
		Statistiques:            make([]ComparaisonStat, 0, len(nomsStats)),
		MultiplicateurTypeRouge: multiplicateurRouge,
		MultiplicateurTypeBleu:  multiplicateurBleu,
		PointsParStatRouge:      pointsParStat(multiplicateurRouge, multiplicateurBleu),
		PointsParStatBleu:       pointsParStat(multiplicateurBleu, multiplicateurRouge),
	}
	for _, nom := range nomsStats {
		valeurRouge, rougeDisponible := statsRouges[nom]
		valeurBleue, bleuDisponible := statsBleues[nom]
		comparaison := ComparaisonStat{
			Nom:        libelleStat(nom),
			Unite:      uniteStat(nom),
			Rouge:      valeurRouge,
			Bleu:       valeurBleue,
			Disponible: rougeDisponible && bleuDisponible,
			Gagnant:    ComparaisonInconnue,
		}
		if !comparaison.Disponible {
			resultat.Statistiques = append(resultat.Statistiques, comparaison)
			continue
		}

		switch {
		case valeurRouge > valeurBleue:
			comparaison.Gagnant = GagnantRouge
			comparaison.PointsRouge = resultat.PointsParStatRouge
			resultat.ScoreRouge += comparaison.PointsRouge
		case valeurBleue > valeurRouge:
			comparaison.Gagnant = GagnantBleu
			comparaison.PointsBleu = resultat.PointsParStatBleu
			resultat.ScoreBleu += comparaison.PointsBleu
		default:
			comparaison.Gagnant = Egalite
		}
		resultat.Statistiques = append(resultat.Statistiques, comparaison)
	}

	switch {
	case resultat.ScoreRouge == 0 && resultat.ScoreBleu == 0 && len(resultat.Statistiques) == 0:
		resultat.Gagnant = ComparaisonInconnue
	case resultat.ScoreRouge > resultat.ScoreBleu:
		resultat.Gagnant = GagnantRouge
	case resultat.ScoreBleu > resultat.ScoreRouge:
		resultat.Gagnant = GagnantBleu
	default:
		resultat.Gagnant = Egalite
	}
	return resultat
}

func valeursStats(stats []BaseStat) map[string]int {
	valeurs := make(map[string]int, len(stats))
	for _, stat := range stats {
		valeurs[stat.Stat.Name] = stat.BaseStat
	}
	return valeurs
}

func multiplicateurType(attaquant, defenseur Choix) float64 {
	faiblesses := make(map[string]float64, len(defenseur.Faiblesses))
	for _, faiblesse := range defenseur.Faiblesses {
		faiblesses[faiblesse.Nom] = faiblesse.Multiplicateur
	}

	multiplicateur := 1.0
	for _, typePokemon := range attaquant.Types {
		if valeur := faiblesses[typePokemon.Type.Name]; valeur > multiplicateur {
			multiplicateur = valeur
		}
	}
	return multiplicateur
}

func pointsParStat(multiplicateur, multiplicateurAdverse float64) int {
	if multiplicateur > 1 && multiplicateur >= multiplicateurAdverse {
		return 2
	}
	return 1
}

func ordreStats(rouges, bleues []BaseStat) []string {
	noms := []string{"hp", "attack", "defense", "special-attack", "special-defense", "speed", "height", "weight"}
	vus := make(map[string]bool, len(noms)+len(rouges)+len(bleues))
	for _, nom := range noms {
		vus[nom] = true
	}
	for _, stats := range [][]BaseStat{rouges, bleues} {
		for _, stat := range stats {
			if !vus[stat.Stat.Name] {
				vus[stat.Stat.Name] = true
				noms = append(noms, stat.Stat.Name)
			}
		}
	}
	return noms
}

func uniteStat(nom string) string {
	switch nom {
	case "height":
		return "dm"
	case "weight":
		return "hg"
	default:
		return ""
	}
}

func libelleStat(nom string) string {
	switch nom {
	case "hp":
		return "HP"
	case "attack":
		return "Attaque"
	case "defense":
		return "Défense"
	case "special-attack":
		return "Attaque spéciale"
	case "special-defense":
		return "Défense spéciale"
	case "speed":
		return "Vitesse"
	case "height":
		return "Taille"
	case "weight":
		return "Poids"
	default:
		return nom
	}
}
