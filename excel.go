package main

import (
	"strings"

	"github.com/tealeg/xlsx/v3"
)

func loadExcel(excelFileName string) (Users, map[string]Roles, error) {
	xlFile, err := xlsx.OpenFile(excelFileName)
	if err != nil {
		return nil, nil, err
	}
	var table [][]string
	var zones [][]string
	for _, sheet := range xlFile.Sheets {
		if sheet.Name == "utilisateurs" {
			table = loadSheet(*sheet)
		}
		if sheet.Name == "zones" {
			zones = loadSheet(*sheet)
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
		email := Username(strings.Trim(strings.ToLower(userRow[fields["ADRESSE MAIL"]]), " "))

		if email != "" && len(userRow[fields["PRENOM"]]) > 1 {
			user := User{
				niveau:            strings.ToLower(niveau),
				email:             email,
				nom:               strings.ToUpper(userRow[fields["NOM"]]),
				prenom:            strings.ToUpper(userRow[fields["PRENOM"]][0:1]) + strings.ToLower(userRow[fields["PRENOM"]][1:]),
				segment:           userRow[fields["SEGMENT"]],
				fonction:          userRow[fields["FONCTION"]],
				employeur:         userRow[fields["ENTITES"]],
				goup:              userRow[fields["GOUP"]],
				accesGeographique: userRow[fields["ACCES GEOGRAPHIQUE"]],
				boards:            strings.Split(strings.Trim(userRow[fields["BOARDS"]], ""), ","),
				taskforce:         strings.Split(strings.Trim(userRow[fields["TASKFORCE"]], ""), ","),
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

func loadSheet(sheet xlsx.Sheet) [][]string {
	var r [][]string
	_ = sheet.ForEachRow(func(row *xlsx.Row) error {
		var line []string
		_ = row.ForEachCell(func(cell *xlsx.Cell) error {
			line = append(line, cell.Value)
			return nil
		})
		if isNotEmpty(line) {
			r = append(r, line)
		}
		return nil
	})
	return r
}

func isNotEmpty(array []string) bool {
	if array == nil {
		return false
	}
	for _, current := range array {
		if current != "" {
			return true
		}
	}
	return false
}
