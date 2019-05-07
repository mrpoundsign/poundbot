package types

import (
	"reflect"
	"testing"
)

func TestServer_UsersClan(t *testing.T) {
	clans := []Clan{
		Clan{Tag: "FoF", Members: []string{"one", "two"}},
	}
	type args struct {
		playerIDs []string
	}
	tests := []struct {
		name  string
		args  args
		want  bool
		want1 Clan
	}{
		{
			name:  "User is in no clans",
			args:  args{playerIDs: []string{"three"}},
			want:  false,
			want1: Clan{},
		},
		{
			name:  "User is in clan",
			args:  args{playerIDs: []string{"two"}},
			want:  true,
			want1: clans[0],
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Server{
				Clans: clans,
			}
			got, got1 := s.UsersClan(tt.args.playerIDs)
			if got != tt.want {
				t.Errorf("Server.UsersClan() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Server.UsersClan() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestAccount_ServerFromKey(t *testing.T) {
	servers := []Server{
		Server{Key: "one"},
		Server{Key: "two"},
	}
	type args struct {
		apiKey string
	}
	tests := []struct {
		name    string
		args    args
		want    Server
		wantErr bool
	}{
		{
			name:    "Server does not exist",
			args:    args{apiKey: "three"},
			want:    Server{},
			wantErr: true,
		},
		{
			name:    "Server exists",
			args:    args{apiKey: "two"},
			want:    servers[1],
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Account{
				Servers: servers,
			}
			got, err := a.ServerFromKey(tt.args.apiKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("Account.ServerFromKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Account.ServerFromKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAccount_GetCommandPrefix(t *testing.T) {
	tests := []struct {
		name          string
		commandPrefix string
		want          string
	}{
		{
			name: "no command prefix",
			want: "!pb",
		},
		{
			name:          "command prefix",
			commandPrefix: "!awwwyeah",
			want:          "!awwwyeah",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Account{
				BaseAccount: BaseAccount{CommandPrefix: tt.commandPrefix},
			}
			if got := a.GetCommandPrefix(); got != tt.want {
				t.Errorf("Account.GetCommandPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAccount_GetAdminIDs(t *testing.T) {
	tests := []struct {
		name        string
		baseAccount BaseAccount
		want        []string
	}{
		{
			name:        "owner only",
			baseAccount: BaseAccount{OwnerSnowflake: "one"},
			want:        []string{"one"},
		},
		{
			name:        "owner and admins",
			baseAccount: BaseAccount{OwnerSnowflake: "one", AdminSnowflakes: []string{"two", "three"}},
			want:        []string{"two", "three", "one"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Account{
				BaseAccount: tt.baseAccount,
			}
			if got := a.GetAdminIDs(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Account.GetAdminIDs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClan_SetGame(t *testing.T) {
	type fields struct {
		Tag        string
		OwnerID    string
		Members    []string
		Moderators []string
	}
	type args struct {
		game string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Clan{
				Tag:        tt.fields.Tag,
				OwnerID:    tt.fields.OwnerID,
				Members:    tt.fields.Members,
				Moderators: tt.fields.Moderators,
			}
			c.SetGame(tt.args.game)
		})
	}
}

func TestAccount_GetRegisteredPlayerIDs(t *testing.T) {
	type fields struct {
		BaseAccount BaseAccount
	}
	type args struct {
		game string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []string
	}{
		{
			name:   "empty",
			args:   args{"game"},
			fields: fields{BaseAccount: BaseAccount{}},
			want:   []string{},
		},
		{
			name: "different game",
			args: args{"game"},
			fields: fields{BaseAccount: BaseAccount{
				AuthenticatedPlayerIDs: []string{"rust:1234", "rust:2345"},
			}},
			want: []string{},
		},
		{
			name: "mixed game",
			args: args{"game"},
			fields: fields{BaseAccount: BaseAccount{
				AuthenticatedPlayerIDs: []string{
					"rust:1234", "rust:2345",
					"game:3456", "game:4567",
				},
			}},
			want: []string{"3456", "4567"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Account{
				BaseAccount: tt.fields.BaseAccount,
			}
			if got := a.GetRegisteredPlayerIDs(tt.args.game); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Account.GetRegisteredPlayerIDs() = %v, want %v", got, tt.want)
			}
		})
	}
}
