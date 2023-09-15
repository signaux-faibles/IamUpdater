package main

import (
	"fmt"
	"strings"

	"github.com/tealeg/xlsx/v3"
)

var HEADERS = []string{
	"NIVEAU HABILITATION",
	"ENTITES",
	"ACCES GEOGRAPHIQUE",
	"FONCTION",
	"SEGMENT",
	"PRENOM",
	"NOM",
	"ADRESSE MAIL",
	"GOUP",
	"SCOPE",
	"BOARDS",
	"TASKFORCE",
}

var NOM_PREMIERE_PAGE = "utilisateurs"

func splitExcelValue(value string, sep string) []string {
	splitValue := strings.Split(value, sep)
	trimmedValue := mapSlice(splitValue, strings.TrimSpace)
	return selectSlice(trimmedValue, func(s string) bool { return s != "" })
}

func loadExcel(excelFileName string) (Users, map[string]Roles, error) {

	xlFile, err := xlsx.OpenFile(excelFileName)
	if err != nil {
		return nil, nil, err
	}
	err = checkExcelFormat(xlFile)
	if err != nil {
		return nil, nil, err
	}
	var table [][]string
	var zones [][]string
	for _, sheet := range xlFile.Sheets {

		if sheet.Name == NOM_PREMIERE_PAGE {
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
		email := Username(strings.TrimSpace(strings.ToLower(userRow[fields["ADRESSE MAIL"]])))

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
				scope:             splitExcelValue(userRow[fields["SCOPE"]], ","),
				boards:            splitExcelValue(userRow[fields["BOARDS"]], ","),
				taskforces:        splitExcelValue(userRow[fields["TASKFORCE"]], ","),
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

func checkExcelFormat(file *xlsx.File) error {
	return checkSheet1Format(file.Sheets[0])
}

func checkSheet1Format(sheet *xlsx.Sheet) error {
	if sheet.Name != NOM_PREMIERE_PAGE {
		return InvalidExcelFileError{msg: fmt.Sprintf("la première page n'a pas le bon nom (%s) : %s", NOM_PREMIERE_PAGE, sheet.Name)}
	}

	firstRow, err := sheet.Row(0)
	if err != nil {
		return InvalidExcelFileError{msg: "erreur lors de la récupération des headers", err: err}
	}
	for i := 0; i < len(HEADERS); i++ {
		cell := firstRow.GetCell(i)
		actual := cell.Value
		expected := HEADERS[i]
		if actual != expected {
			return InvalidExcelFileError{
				msg: fmt.Sprintf("l'entête %s en position %d ne correspond pas à la valeur attendue %s", actual, i, expected),
			}
		}
	}
	return nil
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
