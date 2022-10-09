package main

import (
	"context"
	"github.com/signaux-faibles/keycloakUpdater/v2/logger"
	"github.com/signaux-faibles/libwekan"
)

func manageBoardsMembers(wekan libwekan.Wekan, fromConfig Users) error {
	fields := logger.DataForMethod("manageBoardsMembers")
	wekanBoardsMembers := fromConfig.selectScopeWekan().inferBoardsMember()
	domainBoards, err := wekan.SelectDomainBoards(context.Background())
	if err != nil {
		return err
	}
	wekanBoardsMembers.addBoards(domainBoards)

	logger.Info("> inscrit les utilisateurs dans les tableaux", fields)
	for boardSlug, boardMembers := range wekanBoardsMembers {
		err := setBoardMembers(wekan, boardSlug, boardMembers)
		if err != nil {
			return err
		}
	}
	return nil
}

func setBoardMembers(wekan libwekan.Wekan, boardSlug libwekan.BoardSlug, boardMembers Users) error {
	fields := logger.DataForMethod("setBoardMembers")
	fields.AddAny("board", boardSlug)

	board, err := wekan.GetBoardFromSlug(context.Background(), boardSlug)
	if err != nil {
		return err
	}
	currentMembersIDs := mapSlice(board.Members, func(member libwekan.BoardMember) libwekan.UserID { return member.UserID })

	// globalWekan.AdminUser() est membre de toutes les boards, ajoutons le ici pour ne pas risquer de l'oublier dans les utilisateurs
	wantedMembersUsernames := []libwekan.Username{wekan.AdminUsername()}
	for username := range boardMembers {
		wantedMembersUsernames = append(wantedMembersUsernames, libwekan.Username(username))
	}

	currentMembers, err := wekan.GetUsersFromIDs(context.Background(), currentMembersIDs)
	if err != nil {
		return err
	}
	wantedMembers, err := wekan.GetUsersFromUsernames(context.Background(), wantedMembersUsernames)
	if err != nil {
		return err
	}

	currentUserMap := mapifySlice(currentMembers, libwekan.User.GetID)
	wantedUserMap := mapifySlice(wantedMembers, libwekan.User.GetID)

	wantedMembersIDs := mapSlice(wantedMembers, func(user libwekan.User) libwekan.UserID { return user.ID })
	alreadyBoardMember, wantedInactiveBoardMember, ongoingBoardMember := intersect(currentMembersIDs, wantedMembersIDs)

	logger.Debug(">> examine les nouvelles inscriptions", fields)
	for _, userID := range append(alreadyBoardMember, ongoingBoardMember...) {
		fields.AddAny("username", wantedUserMap[userID].Username)
		logger.Debug(">>> examine l'utilisateur", fields)
		modified, err := wekan.EnsureUserIsActiveBoardMember(context.Background(), board.ID, userID)
		if err != nil {
			return err
		}
		if modified {
			logger.Info(">>> inscrit l'utilisateur", fields)
		}
	}

	logger.Debug(">> examine les radiations", fields)
	isOauth2UserID := func(userID libwekan.UserID) bool {
		return currentUserMap[userID].AuthenticationMethod == "oauth2" || currentUserMap[userID].Username == wekan.AdminUsername()
	}
	oauth2OnlyWantedInactiveBoardMember := selectSlice(wantedInactiveBoardMember, isOauth2UserID)
	for _, userID := range oauth2OnlyWantedInactiveBoardMember {
		fields.AddAny("username", currentUserMap[userID].Username)
		logger.Debug(">>> vérifie la non-participation", fields)
		modified, err := wekan.EnsureUserIsInactiveBoardMember(context.Background(), board.ID, userID)
		if err != nil {
			return err
		}
		if modified {
			logger.Info(">>> désinscrit le participant", fields)
		}
	}

	// globalWekan.AdminUser() est administrateur de toutes les boards, appliquons la règle
	logger.Debug(">> vérifie la participation de l'admin", fields)

	modified, err := wekan.EnsureUserIsBoardAdmin(context.Background(), board.ID, wekan.AdminID())
	if modified {
		fields.AddAny("username", wekan.AdminUsername())
		logger.Info(">>> donne les privilèges à l'admin", fields)
	}
	return err
}

func (users Users) inferBoardsMember() BoardsMembers {
	wekanBoardsUserSlice := make(map[libwekan.BoardSlug][]User)
	for _, user := range users {
		for _, boardSlug := range user.boards {
			if boardSlug != "" {
				boardSlug := libwekan.BoardSlug(boardSlug)
				wekanBoardsUserSlice[boardSlug] = append(wekanBoardsUserSlice[boardSlug], user)
			}
		}
	}

	wekanBoardsUsers := make(BoardsMembers)
	for boardSlug, userSlice := range wekanBoardsUserSlice {
		wekanBoardsUsers[boardSlug] = mapifySlice(userSlice, func(user User) Username { return user.email })
	}
	return wekanBoardsUsers
}

type BoardsMembers map[libwekan.BoardSlug]Users

func (boardsMembers BoardsMembers) addBoards(boards []libwekan.Board) BoardsMembers {
	if boardsMembers == nil {
		boardsMembers = make(BoardsMembers)
	}
	for _, b := range boards {
		if _, ok := boardsMembers[b.Slug]; !ok {
			boardsMembers[b.Slug] = make(Users)
		}
	}
	return boardsMembers
}
