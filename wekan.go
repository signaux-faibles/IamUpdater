package main

import (
	"context"
	"fmt"
	"github.com/signaux-faibles/keycloakUpdater/v2/logger"
	"github.com/signaux-faibles/libwekan"
)

type Pipeline []PipelineStage

type PipelineStage struct {
	run func(libwekan.Wekan, Users) error
	id  string
}

func (pipeline Pipeline) Run(wekan libwekan.Wekan, fromConfig Users) error {
	for _, stage := range pipeline {
		err := stage.run(wekan, fromConfig)
		if err != nil {
			return PipelineRunError{
				err:   err,
				stage: stage,
			}
		}
	}
	return nil
}

func (pipeline Pipeline) StopAfter(wekan libwekan.Wekan, fromConfig Users, lastStage PipelineStage) error {
	fields := logger.DataForMethod("StopAfter")
	for _, stage := range pipeline {
		fields.AddAny("stage", stage.id)
		logger.Debug("Application du pipeline", fields)
		err := stage.run(wekan, fromConfig)
		if err != nil || stage.id == lastStage.id {
			return err
		}
	}
	return nil
}

var StageCheckBoardSlugs = PipelineStage{CheckBoardSlugs, "CheckBoardSlugs"}
var StageManageUsers = PipelineStage{ManageUsers, "ManageUsers"}
var StageManageBoardsMembers = PipelineStage{ManageBoardsMembers, "ManageBoardsMembers"}
var StageAddMissingRulesAndCardMembership = PipelineStage{AddMissingRulesAndCardMembership, "AddMissingRulesAndCardMembership"}
var StageRemoveExtraRulesAndCardMembership = PipelineStage{RemoveExtraRulesAndCardsMembership, "RemoveExtraRulesAndCardMembership"}
var StageCheckNativeUsers = PipelineStage{CheckNativeUsers, "CheckNativeUsers"}

var pipeline = Pipeline{
	StageCheckBoardSlugs,
	StageCheckNativeUsers,
	StageManageUsers,
	StageManageBoardsMembers,
	StageAddMissingRulesAndCardMembership,
	StageRemoveExtraRulesAndCardMembership,
}

func WekanUpdate(url, database, admin string, users Users, slugDomainRegexp string) error {
	wekan, err := initWekan(url, database, admin, slugDomainRegexp)
	if err != nil {
		return err
	}
	return pipeline.Run(wekan, users)
}

func addAdmin(usersFromExcel Users, wekan libwekan.Wekan) {
	usersFromExcel[Username(wekan.AdminUsername())] = User{
		email: Username(wekan.AdminUsername()),
		scope: []string{"wekan"},
	}
}

func initWekan(url string, database string, admin string, slugDomainRegexp string) (libwekan.Wekan, error) {
	wekan, err := libwekan.Init(context.Background(), url, database, libwekan.Username(admin), slugDomainRegexp)
	if err != nil {
		return libwekan.Wekan{}, err
	}

	if err := wekan.Ping(context.Background()); err != nil {
		return libwekan.Wekan{}, err
	}
	err = wekan.AssertPrivileged(context.Background())
	if err != nil {
		return libwekan.Wekan{}, err
	}
	return wekan, nil
}

func (users Users) ListWekanChanges(wekanUsers libwekan.Users) (
	creations libwekan.Users,
	enable libwekan.Users,
	disable libwekan.Users,
) {
	wekanUsernames := mapSlice(wekanUsers, usernameFromWekanUser)
	configUsernames := users.Usernames()

	both, onlyWekan, notInWekan := intersect(wekanUsernames, configUsernames)
	creations = UsernamesSelect(users, notInWekan).BuildWekanUsers()
	enable = WekanUsernamesSelect(wekanUsers, both)
	disable = WekanUsernamesSelect(wekanUsers, onlyWekan)

	return creations, enable, disable
}

func usernameFromWekanUser(user libwekan.User) Username {
	return Username(user.Username)
}

func firstChar(s string) string {
	if len(s) > 0 {
		return s[0:1]
	}
	return ""
}

func CheckBoardSlugs(wekan libwekan.Wekan, users Users) error {
	domainBoards, err := wekan.SelectDomainBoards(context.Background())
	if err != nil {
		return err
	}

	domainActiveBoards := selectSlice(domainBoards, func(board libwekan.Board) bool { return !board.Archived })
	boardToSlug := func(board libwekan.Board) libwekan.BoardSlug { return board.Slug }
	domainActiveSlugs := mapSlice(domainActiveBoards, boardToSlug)

	configBoardsMembers := users.inferBoardsMember()
	var configSlugs []libwekan.BoardSlug
	for slug := range configBoardsMembers {
		configSlugs = append(configSlugs, slug)
	}
	_, _, onlyConfig := intersect(domainActiveSlugs, configSlugs)

	if len(onlyConfig) > 0 {
		return InvalidExcelFileError{msg: fmt.Sprintf("le fichier contient des références de boards inexistantes : %s", onlyConfig)}
	}
	return nil
}
