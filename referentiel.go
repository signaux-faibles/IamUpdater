package main

import (
	"encoding/csv"
	"log"
	"os"
)

func buildReferentiel(input [][]string) map[string]Roles {
	compositeRoles := make(map[string]Roles)
	for _, z := range input[1:] {
		compositeRoles[z[0]] = append(
			compositeRoles[z[0]],
			z[2],
		)
		compositeRoles[z[1]] = append(
			compositeRoles[z[1]],
			z[2],
		)
	}
	return compositeRoles
}

func loadReferentiel(filename string) map[string]Roles {
	// open file
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	// remember to close the file at the end of the program
	defer f.Close()

	// read csv values using csv.Reader
	csvReader := csv.NewReader(f)
	csvReader.Comma = ';'
	data, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	// convert records to array of structs
	compositesRoles := buildReferentiel(data)

	return compositesRoles
}
