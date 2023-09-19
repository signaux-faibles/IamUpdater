package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_composeReplaceAttr_functions(t *testing.T) {
	ass := assert.New(t)
	expectedLevel := slog.LevelInfo
	type args struct {
		funcs []func(unused []string, a slog.Attr) slog.Attr
		level slog.Level
	}
	tests := []struct {
		name     string
		args     args
		want     []string
		dontWant []string
	}{
		{
			name: "on brouille l heure et le level",
			args: args{
				funcs: []func(unused []string, a slog.Attr) slog.Attr{onMasqueLHeure, onMasqueLeLevel},
				level: expectedLevel,
			},
			want:     []string{"time=on_masque_l_heure", "level=on_masque_le_level"},
			dontWant: []string{"level=" + expectedLevel.String()},
		},
		{
			name: "on brouille l'heure 2x ainsi que le level",
			args: args{
				funcs: []func(unused []string, a slog.Attr) slog.Attr{onMasqueLHeure, onMasqueLeLevel, onAfficheUneHeureFausse},
				level: slog.LevelInfo,
			},
			want:     []string{"time=00h00_l_heure_du_crime", "level=on_masque_le_level"},
			dontWant: []string{"time=on_masque_l_heure", "level=" + expectedLevel.String()},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// création du fichier de log
			logFilename := createTempFilename(t)
			logfile, err := os.Create(logFilename)
			ass.NoError(err)

			fileHandler := slog.NewTextHandler(logfile, &slog.HandlerOptions{
				ReplaceAttr: composeReplaceAttrs(tt.args.funcs...),
				AddSource:   true,
			})
			logger := slog.New(fileHandler)
			logger.Log(context.Background(), tt.args.level, "message de log")

			var logsFromFile []byte
			logsFromFile, err = os.ReadFile(logFilename)
			fmt.Println(string(logsFromFile))
			ass.NoError(err)
			for _, expected := range tt.want {
				ass.Contains(string(logsFromFile), expected)
			}
			for _, expectedNot := range tt.dontWant {
				ass.NotContains(string(logsFromFile), expectedNot)
			}
		})
	}
}

func onMasqueLHeure(_ []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey {
		a.Value = slog.StringValue("on_masque_l_heure")
	}
	return a
}

func onMasqueLeLevel(_ []string, a slog.Attr) slog.Attr {
	if a.Key == slog.LevelKey {
		a.Value = slog.StringValue("on_masque_le_level")
	}
	return a
}

func onAfficheUneHeureFausse(_ []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey {
		a.Value = slog.StringValue("00h00_l_heure_du_crime")
	}
	return a
}
