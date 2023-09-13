package logger

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/Nerzal/gocloak/v13"
	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/assert"

	"keycloakUpdater/v2/pkg/structs"
)

var fake faker.Faker

func init() {
	fake = faker.New()
}

func Test_levelConfiguration(t *testing.T) {
	ass := assert.New(t)
	loggerConfig := structs.LoggerConfig{
		Filename:        createTempFilename(t),
		Level:           "waRN",
		TimestampFormat: time.DateTime,
	}
	ConfigureWith(loggerConfig)
	log.Print("message d'info (à l'ancienne)")
	slog.Info("message d'info")
	slog.Warn("message de warn")
	slog.Error("message d'erreur", slog.Any("error", io.EOF))

	var logsFromFile []byte
	var err error
	logsFromFile, err = os.ReadFile(loggerConfig.Filename)
	ass.NoError(err)
	ass.NotContains(string(logsFromFile), "message d'info")
	ass.Contains(string(logsFromFile), "message de warn")
	ass.Contains(string(logsFromFile), "message d'erreur")
}

func Test_formatter_User(t *testing.T) {
	ass := assert.New(t)
	person := fake.Person()
	id := person.Contact().Email
	username := fake.Internet().User()
	firstname := person.FirstName()
	lastname := person.LastName()
	userToLog := gocloak.User{
		ID:        &id,
		Username:  &username,
		FirstName: &firstname,
		LastName:  &lastname,
	}
	logger := defaultDebugLogger(t)
	ConfigureWith(logger)
	key := fake.Lorem().Word()
	slog.Info("message d'info", slog.Any(key, userToLog))

	var logsFromFile []byte
	var err error
	logsFromFile, err = os.ReadFile(logger.Filename)
	ass.NoError(err)
	ass.NotContains(string(logsFromFile), id)
	ass.NotContains(string(logsFromFile), firstname)
	ass.NotContains(string(logsFromFile), lastname)
	ass.Contains(string(logsFromFile), key+"="+username)
}

func Test_formatter_Client(t *testing.T) {
	ass := assert.New(t)
	id := fake.Person().SSN()
	clientId := fake.Person().SSN()
	name := fake.Internet().User()
	objectToLog := gocloak.Client{
		ClientID: &clientId,
		ID:       &id,
		Name:     &name,
	}
	logger := defaultDebugLogger(t)
	ConfigureWith(logger)
	key := fake.Lorem().Word()
	slog.Info("message d'info", slog.Any(key, objectToLog))

	var logsFromFile []byte
	var err error
	logsFromFile, err = os.ReadFile(logger.Filename)
	ass.NoError(err)
	ass.NotContains(string(logsFromFile), id)
	ass.NotContains(string(logsFromFile), name)
	ass.Contains(string(logsFromFile), key+"="+clientId)
}

func Test_formatter_Role(t *testing.T) {
	ass := assert.New(t)
	id := fake.Person().SSN()
	name := fake.Internet().User()
	description := fake.Lorem().Text(256)

	objectToLog := gocloak.Role{
		ID:          &id,
		Name:        &name,
		Description: &description,
	}
	logger := defaultDebugLogger(t)
	ConfigureWith(logger)
	key := fake.Lorem().Word()
	slog.Info("message d'info", slog.Any(key, objectToLog))

	var logsFromFile []byte
	var err error
	logsFromFile, err = os.ReadFile(logger.Filename)
	ass.NoError(err)
	ass.NotContains(string(logsFromFile), id)
	ass.NotContains(string(logsFromFile), description)
	ass.Contains(string(logsFromFile), key+"="+name)
}

// func Test_formatter_Time(t *testing.T) {
//
// 	ass := assert.New(t)
// 	tuTime := time.Date(2023, 9, 11, 15, 47, 32, 99, time.Local)
// 	FakeTime(t, tuTime)
// 	t.Cleanup(func() { unfakeTime() })
//
// 	timeFormatter := timeFormatter(time.DateTime)
//
// 	formattingMiddleware := slogformatter.NewFormatterHandler(timeFormatter)
// 	logFilename := createTempFilename(t)
// 	logfile, err := os.Create(logFilename)
// 	ass.NoError(err)
// 	fileHandler := slog.NewTextHandler(logfile, &slog.HandlerOptions{})
//
// 	logger := slog.New(slogmulti.Pipe(formattingMiddleware).Handler(fileHandler))
//
// 	logger.Info("message d'info")
//
// 	var logsFromFile []byte
// 	logsFromFile, err = os.ReadFile(logFilename)
// 	ass.NoError(err)
// 	ass.Contains(string(logsFromFile), "time=2023-09-11 15:47:32")
// }

func defaultDebugLogger(t *testing.T) structs.LoggerConfig {
	loggerConfig := structs.LoggerConfig{
		Filename:        createTempFilename(t),
		Level:           "debug",
		TimestampFormat: time.DateTime,
	}
	return loggerConfig
}

func createTempFilename(t *testing.T) string {
	return fmt.Sprint(t.TempDir(), os.PathSeparator, t.Name())
}

// FakeTime méthode qui permet de fausser la méthode `time.Now` en la forçant à toujours retourner
// le paramètre `t`
func FakeTime(test *testing.T, t time.Time) {
	monkey.Patch(time.Now, func() time.Time {
		return t
	})
	test.Cleanup(unfakeTime)
}

func unfakeTime() {
	monkey.Unpatch(time.Now)
}
