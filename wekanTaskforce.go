package main

import (
	"context"
	"errors"
	"github.com/signaux-faibles/keycloakUpdater/v2/logger"
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

func AddMissingRules(wekan libwekan.Wekan, users Users) error {
	fields := logger.DataForMethod("AddMissingRules")
	for _, user := range users {
		fields.AddAny("username", user.email)
		wekanUser, err := wekan.GetUserFromUsername(context.Background(), libwekan.Username(user.email))
		if err != nil {
			return err
		}

		for _, boardSlug := range user.boards {
			board, err := wekan.GetBoardFromSlug(context.Background(), libwekan.BoardSlug(boardSlug))
			fields.AddAny("board", boardSlug)
			if err != nil {
				return err
			}
			userAcceptTaskforceLabel := func(label libwekan.BoardLabel) bool {
				return contains(user.taskforce, string(label.Name))
			}
			labels := selectSlice(board.Labels, userAcceptTaskforceLabel)
			for _, label := range labels {
				fields.AddAny("label", label.Name)
				logger.Info("s'assure de la présence de la règle", fields)
				err := wekan.EnsureRuleExists(context.Background(), wekanUser, board, label)
				if err != nil {
					return err
				}
			}
		}
		//wekan.SelectBoardsFromMemberID(current.boards)

	}
	return nil
}

func RemoveExtraRules(wekan libwekan.Wekan, users Users) error {
	return errors.New("not implemented")
}

func AddMissingCardsMembers(wekan libwekan.Wekan, users Users) error {
	return errors.New("not implemented")
}

func RemoveExtraCardsMembers(wekan libwekan.Wekan, users Users) error {
	return errors.New("not implemented")
}
