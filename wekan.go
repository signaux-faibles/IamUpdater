package main

import (
	"context"
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
			return err
		}
	}
	return nil
}

func (pipeline Pipeline) StopAfter(wekan libwekan.Wekan, fromConfig Users, lastStage PipelineStage) error {
	fields := logger.DataForMethod("StopAfter")
	for _, stage := range pipeline {
		fields.AddAny("stage", stage.id)
		logger.Info("Application du pipeline", fields)
		err := stage.run(wekan, fromConfig)
		if err != nil || stage.id == lastStage.id {
			return err
		}
	}
	return nil
}

var StageManageUsers = PipelineStage{ManageUsers, "ManageUsers"}
var StageManageBoardsMembers = PipelineStage{ManageBoardsMembers, "ManageBoardsMembers"}
var StageAddMissingRules = PipelineStage{AddMissingRules, "AddMissingRules"}
var StageRemoveExtraRules = PipelineStage{RemoveExtraRules, "RemoveExtraRules"}
var StageAddMissingCardsMembers = PipelineStage{AddMissingCardsMembers, "AddMissingCardsMembers"}
var StageRemoveExtraCardsMembers = PipelineStage{RemoveExtraCardsMembers, "RemoveExtraCardsMembers"}

var pipeline = Pipeline{
	StageManageUsers,
	StageManageBoardsMembers,
	StageAddMissingCardsMembers,
	StageRemoveExtraCardsMembers,
	StageAddMissingRules,
	StageRemoveExtraRules,
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
	err = wekan.AssertHasAdmin(context.Background())
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
