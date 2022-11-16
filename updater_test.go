package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_areYouSureTooApplyChanges(t *testing.T) {
	type args struct {
		changes         int
		keeps           int
		acceptedChanges int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "refuse de supprimer tous les utilisateurs",
			args: args{1234, 0, 0},
			want: false,
		}, {
			name: "accepte moins de changement(s) que le maximum autorisé",
			args: args{1, 1, 2},
			want: true,
		}, {
			name: "refuse plus de changement(s) que le maximum autorisé",
			args: args{2, 1, 1},
			want: false,
		}, {
			name: "accepte tous les changements si le nombre max de changements autorisé est strictement négatif",
			args: args{1234, 1, -1},
			want: true,
		}, {
			name: "accepte tous les changements si le nombre max de changements autorisé est égal à zéro",
			args: args{1234, 1, 0},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, areYouSureTooApplyChanges(tt.args.changes, tt.args.keeps, tt.args.acceptedChanges), "areYouSureTooApplyChanges(%v, %v, %v)", tt.args.changes, tt.args.keeps, tt.args.acceptedChanges)
		})
	}
}
