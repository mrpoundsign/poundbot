package models

import "testing"

func TestDiscordAuth_GetPlayerID(t *testing.T) {
	tests := []struct {
		name PlayerID
		d    DiscordAuth
		want PlayerID
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(string(tt.name), func(t *testing.T) {
			if got := tt.d.GetPlayerID(); got != tt.want {
				t.Errorf("DiscordAuth.GetPlayerID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiscordAuth_GetDiscordID(t *testing.T) {
	tests := []struct {
		name PlayerDiscordID
		d    DiscordAuth
		want PlayerDiscordID
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(string(tt.name), func(t *testing.T) {
			if got := tt.d.GetDiscordID(); got != tt.want {
				t.Errorf("DiscordAuth.GetDiscordID() = %v, want %v", got, tt.want)
			}
		})
	}
}
