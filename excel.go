package main

import (
	"strings"

	"github.com/spf13/viper"
	"github.com/tealeg/xlsx/v3"
)

func loadExcel() (Users, map[string]Roles, error) {
	excelFileName := viper.GetString("base")
	xlFile, err := xlsx.OpenFile(excelFileName)
	if err != nil {
		return nil, nil, err
	}
	var table [][]string
	var zones [][]string
	for _, sheet := range xlFile.Sheets {
		if sheet.Name == "utilisateurs" {
			sheet.ForEachRow(func(row *xlsx.Row) error {
				var r []string
				row.ForEachCell(func(cell *xlsx.Cell) error {
					r = append(r, cell.Value)
					return nil
				})
				table = append(table, r)
				return nil
			})
		}
		if sheet.Name == "zones" {
			sheet.ForEachRow(func(row *xlsx.Row) error {
				var r []string
				row.ForEachCell(func(cell *xlsx.Cell) error {
					r = append(r, cell.Value)
					return nil
				})
				zones = append(zones, r)
				return nil
			})
		}
	}

	users := make(Users)

	fields := make(map[string]int)
	for i, f := range table[0] {
		fields[f] = i
	}

	zoneFields := make(map[string]int)
	for i, f := range zones[0] {
		zoneFields[f] = i
	}

	for _, userRow := range table[1:] {
		niveau := userRow[fields["NIVEAU HABILITATION"]]
		email := strings.Trim(strings.ToLower(userRow[fields["ADRESSE MAIL"]]), " ")

		if email != "" && len(userRow[fields["PRENOM"]]) > 1 {
			user := User{
				niveau:            strings.ToLower(niveau),
				email:             email,
				nom:               strings.ToUpper(userRow[fields["NOM"]]),
				prenom:            strings.ToUpper(userRow[fields["PRENOM"]][0:1]) + strings.ToLower(userRow[fields["PRENOM"]][1:]),
				poste:             userRow[fields["POSTE"]],
				fonction:          userRow[fields["FONCTION"]],
				employeur:         userRow[fields["ENTITES"]],
				goup:              userRow[fields["GOUP"]],
				accesGeographique: userRow[fields["ACCES GEOGRAPHIQUE"]],
			}
			scope := strings.Split(userRow[fields["SCOPE"]], ",")
			if len(scope) != 1 || scope[0] != "" {
				user.scope = scope
			}
			users[email] = user
		}
	}
	compositeRoles := make(map[string]Roles)
	for _, z := range zones[1:] {
		compositeRoles[z[zoneFields["REGION"]]] = append(
			compositeRoles[z[zoneFields["REGION"]]],
			z[zoneFields["DEPARTEMENT"]],
		)
		compositeRoles[z[zoneFields["ANCIENNE REGION"]]] = append(
			compositeRoles[z[zoneFields["ANCIENNE REGION"]]],
			z[zoneFields["DEPARTEMENT"]],
		)
	}
	return users, compositeRoles, nil
}
