package main

import (
	"errors"
	"github.com/signaux-faibles/libwekan"
)

// ManageBoardsLabelsTaskforce
// Objectif de maintenir les règles des tableaux pour les utilisateurs disposant d'un label dans la propriété User.taskforce []string{}
// - Traite les règles sur les tableaux
// - Traite la particitation aux cartes
func ManageBoardsLabelsTaskforce(wekan libwekan.Wekan, fromConfig Users) error {
	if err := ManageRules(); err != nil {
		return err
	}
	return ManageCardsMembers()
}

// ManageRules
// - Ajoute les règles manquantes
// - Supprime les règles superflues
func ManageRules() error { return errors.New("not implemented") }

// ManageCardsMembers
// - Ajoute les utilisateurs entrant dans une taskforce comme membre des cartes concernées
// - Supprime les utilisateurs sortant d'une taskforce des cartes concernées
func ManageCardsMembers() error { return errors.New("not implemented") }
