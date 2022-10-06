package main

import "sort"

type Geographie struct {
	value []row
}

type row struct {
	Region         string
	AncienneRegion string
	Departement    string
}

var franceEntiere = "France Entière"

var referentiel = Geographie{
	[]row{
		{"Grand Est", "Alsace", "67"},
		{"Grand Est", "Alsace", "68"},
		{"Grand Est", "Champagne-Ardenne", "08"},
		{"Grand Est", "Champagne-Ardenne", "10"},
		{"Grand Est", "Champagne-Ardenne", "51"},
		{"Grand Est", "Champagne-Ardenne", "52"},
		{"Grand Est", "Lorraine", "54"},
		{"Grand Est", "Lorraine", "55"},
		{"Grand Est", "Lorraine", "57"},
		{"Grand Est", "Lorraine", "88"},
		{"Nouvelle-Aquitaine", "Aquitaine", "24"},
		{"Nouvelle-Aquitaine", "Aquitaine", "33"},
		{"Nouvelle-Aquitaine", "Aquitaine", "40"},
		{"Nouvelle-Aquitaine", "Aquitaine", "47"},
		{"Nouvelle-Aquitaine", "Aquitaine", "64"},
		{"Nouvelle-Aquitaine", "Limousin", "19"},
		{"Nouvelle-Aquitaine", "Limousin", "23"},
		{"Nouvelle-Aquitaine", "Limousin", "87"},
		{"Nouvelle-Aquitaine", "Poitou-Charentes", "16"},
		{"Nouvelle-Aquitaine", "Poitou-Charentes", "17"},
		{"Nouvelle-Aquitaine", "Poitou-Charentes", "79"},
		{"Nouvelle-Aquitaine", "Poitou-Charentes", "86"},
		{"Auvergne-Rhône-Alpes", "Auvergne", "03"},
		{"Auvergne-Rhône-Alpes", "Auvergne", "15"},
		{"Auvergne-Rhône-Alpes", "Auvergne", "43"},
		{"Auvergne-Rhône-Alpes", "Auvergne", "63"},
		{"Auvergne-Rhône-Alpes", "Rhône-Alpes", "01"},
		{"Auvergne-Rhône-Alpes", "Rhône-Alpes", "07"},
		{"Auvergne-Rhône-Alpes", "Rhône-Alpes", "26"},
		{"Auvergne-Rhône-Alpes", "Rhône-Alpes", "38"},
		{"Auvergne-Rhône-Alpes", "Rhône-Alpes", "42"},
		{"Auvergne-Rhône-Alpes", "Rhône-Alpes", "69"},
		{"Auvergne-Rhône-Alpes", "Rhône-Alpes", "73"},
		{"Auvergne-Rhône-Alpes", "Rhône-Alpes", "74"},
		{"Normandie", "Basse-Normandie", "14"},
		{"Normandie", "Basse-Normandie", "50"},
		{"Normandie", "Basse-Normandie", "61"},
		{"Normandie", "Haute-Normandie", "27"},
		{"Normandie", "Haute-Normandie", "76"},
		{"Bourgogne-Franche-Comté", "Bourgogne", "21"},
		{"Bourgogne-Franche-Comté", "Bourgogne", "58"},
		{"Bourgogne-Franche-Comté", "Bourgogne", "71"},
		{"Bourgogne-Franche-Comté", "Bourgogne", "89"},
		{"Bourgogne-Franche-Comté", "Franche-Comté", "25"},
		{"Bourgogne-Franche-Comté", "Franche-Comté", "39"},
		{"Bourgogne-Franche-Comté", "Franche-Comté", "70"},
		{"Bourgogne-Franche-Comté", "Franche-Comté", "90"},
		{"Bretagne", "Bretagne", "22"},
		{"Bretagne", "Bretagne", "29"},
		{"Bretagne", "Bretagne", "35"},
		{"Bretagne", "Bretagne", "56"},
		{"Centre-Val de Loire", "Centre", "18"},
		{"Centre-Val de Loire", "Centre", "28"},
		{"Centre-Val de Loire", "Centre", "36"},
		{"Centre-Val de Loire", "Centre", "37"},
		{"Centre-Val de Loire", "Centre", "41"},
		{"Centre-Val de Loire", "Centre", "45"},
		{"Corse", "Corse", "2A"},
		{"Corse", "Corse", "2B"},
		{"Guadeloupe", "Guadeloupe", "971"},
		{"Martinique", "Martinique", "972"},
		{"Guyane", "Guyane", "973"},
		{"La Réunion", "La Réunion", "974"},
		{"Mayotte", "Mayotte", "976"},
		{"Île-de-France", "Île-de-France", "75"},
		{"Île-de-France", "Île-de-France", "77"},
		{"Île-de-France", "Île-de-France", "78"},
		{"Île-de-France", "Île-de-France", "91"},
		{"Île-de-France", "Île-de-France", "92"},
		{"Île-de-France", "Île-de-France", "93"},
		{"Île-de-France", "Île-de-France", "94"},
		{"Île-de-France", "Île-de-France", "95"},
		{"Occitanie", "Languedoc-Roussillon", "11"},
		{"Occitanie", "Languedoc-Roussillon", "30"},
		{"Occitanie", "Languedoc-Roussillon", "34"},
		{"Occitanie", "Languedoc-Roussillon", "48"},
		{"Occitanie", "Languedoc-Roussillon", "66"},
		{"Occitanie", "Midi-Pyrénées", "09"},
		{"Occitanie", "Midi-Pyrénées", "12"},
		{"Occitanie", "Midi-Pyrénées", "31"},
		{"Occitanie", "Midi-Pyrénées", "32"},
		{"Occitanie", "Midi-Pyrénées", "46"},
		{"Occitanie", "Midi-Pyrénées", "65"},
		{"Occitanie", "Midi-Pyrénées", "81"},
		{"Occitanie", "Midi-Pyrénées", "82"},
		{"Hauts-de-France", "Nord-Pas-de-Calais", "59"},
		{"Hauts-de-France", "Nord-Pas-de-Calais", "62"},
		{"Hauts-de-France", "Picardie", "02"},
		{"Hauts-de-France", "Picardie", "60"},
		{"Hauts-de-France", "Picardie", "80"},
		{"Pays de la Loire", "Pays de la Loire", "44"},
		{"Pays de la Loire", "Pays de la Loire", "49"},
		{"Pays de la Loire", "Pays de la Loire", "53"},
		{"Pays de la Loire", "Pays de la Loire", "72"},
		{"Pays de la Loire", "Pays de la Loire", "85"},
		{"Provence-Alpes-Côte d'Azur", "Provence-Alpes-Côte d'Azur", "04"},
		{"Provence-Alpes-Côte d'Azur", "Provence-Alpes-Côte d'Azur", "05"},
		{"Provence-Alpes-Côte d'Azur", "Provence-Alpes-Côte d'Azur", "06"},
		{"Provence-Alpes-Côte d'Azur", "Provence-Alpes-Côte d'Azur", "13"},
		{"Provence-Alpes-Côte d'Azur", "Provence-Alpes-Côte d'Azur", "83"},
		{"Provence-Alpes-Côte d'Azur", "Provence-Alpes-Côte d'Azur", "84"},
	},
}

func (ref Geographie) toRoles() CompositeRoles {
	roles := make(CompositeRoles)
	for _, row := range ref.value {
		roles.addRole(row.Region, row.Departement)
		roles.addRole(row.AncienneRegion, row.Departement)
		roles.addRole(franceEntiere, row.Departement)
	}
	ordonne(roles)
	return roles
}

func ordonne(roles CompositeRoles) {
	keys := make([]string, 0, len(roles))
	for k := range roles {
		keys = append(keys, k)
		sort.Strings(roles[k])
	}
	sort.Strings(keys)
}
