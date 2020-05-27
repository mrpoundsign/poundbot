// +build integration

package mongodb

import (
	"testing"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/poundbot/poundbot/pbclock"
	"github.com/poundbot/poundbot/pkg/models"
	"github.com/poundbot/poundbot/storage/mongodb/mongotest"
	"github.com/stretchr/testify/assert"
)

var baseAccount = models.Account{
	BaseAccount: models.BaseAccount{GuildSnowflake: "snowflake"},
	Timestamp:   models.Timestamp{CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond)},
	Servers:     []models.AccountServer{{Key: "key"}},
}

func NewAccounts(t *testing.T) (*Accounts, *mongotest.Collection) {
	coll, err := mongotest.NewCollection(accountsCollection)
	if err != nil {
		t.Fatal(err)
	}
	return &Accounts{collection: coll.C}, coll
}

func TestAccounts_All(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		want    []models.Account
		wantErr bool
	}{
		{
			name: "empty",
			want: nil,
		},
		{
			name: "some",
			want: []models.Account{
				{
					ID:        bson.NewObjectId(),
					Timestamp: models.Timestamp{CreatedAt: time.Now().UTC().Truncate(time.Millisecond)},
					Servers:   []models.AccountServer{{Key: "key", Clans: []models.Clan{}}},
				},
				{
					ID:        bson.NewObjectId(),
					Timestamp: models.Timestamp{CreatedAt: time.Now().UTC().Add(-1 * time.Hour).Truncate(time.Millisecond)},
					Servers:   []models.AccountServer{{Key: "key2", Clans: []models.Clan{}}},
				},
				{
					ID:        bson.NewObjectId(),
					Timestamp: models.Timestamp{CreatedAt: time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Millisecond)},
					Servers:   []models.AccountServer{{Key: "key3", Clans: []models.Clan{}}},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accounts, coll := NewAccounts(t)
			defer coll.Close()

			for _, account := range tt.want {
				coll.C.Insert(account)
			}

			var res []models.Account

			if err := accounts.All(&res); (err != nil) != tt.wantErr {
				t.Errorf("Accounts.GetByDiscordGuild() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.want, res, "Did not get acounts.")
		})
	}
}

func TestAccounts_GetByDiscordGuild(t *testing.T) {
	t.Parallel()

	id := bson.NewObjectId()

	tests := []struct {
		name    string
		key     string
		want    models.Account
		wantErr bool
	}{
		{
			name: "exists",
			want: models.Account{
				ID:          id,
				Timestamp:   models.Timestamp{CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond)},
				BaseAccount: models.BaseAccount{GuildSnowflake: "found"},
				Servers:     []models.AccountServer{{Key: "key2", Clans: []models.Clan{}}},
			},
			key: "found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accounts, coll := NewAccounts(t)
			defer coll.Close()

			make := []models.Account{
				{
					Timestamp:   models.Timestamp{CreatedAt: time.Now().UTC().Truncate(time.Millisecond)},
					BaseAccount: models.BaseAccount{GuildSnowflake: "lost"},
					Servers:     []models.AccountServer{{Key: "key", Clans: []models.Clan{}}},
				},
				{
					ID:          id,
					Timestamp:   models.Timestamp{CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond)},
					BaseAccount: models.BaseAccount{GuildSnowflake: "found"},
					Servers:     []models.AccountServer{{Key: "key2", Clans: []models.Clan{}}},
				},
				{
					Timestamp:   models.Timestamp{CreatedAt: time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Millisecond)},
					BaseAccount: models.BaseAccount{GuildSnowflake: "lost2"},
					Servers:     []models.AccountServer{{Key: "key3", Clans: []models.Clan{}}},
				},
			}

			for _, account := range make {
				coll.C.Insert(account)
			}

			got, err := accounts.GetByDiscordGuild(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Accounts.GetByDiscordGuild() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.want, got, "Account is not what we expected")
		})
	}
}

func TestAccounts_GetByServerKey(t *testing.T) {
	t.Parallel()

	id := bson.NewObjectId()

	type args struct {
		key string
	}
	tests := []struct {
		name    string
		args    args
		want    models.Account
		wantErr bool
	}{
		{
			name: "result",
			args: args{key: "key2"},
			want: models.Account{
				ID:        id,
				Timestamp: models.Timestamp{CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond)},
				Servers:   []models.AccountServer{{Key: "key2", Clans: []models.Clan{}}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accounts, coll := NewAccounts(t)
			defer coll.Close()

			docs := []models.Account{
				{
					Timestamp: models.Timestamp{CreatedAt: time.Now().UTC().Truncate(time.Millisecond)},
					Servers:   []models.AccountServer{{Key: "key", Clans: []models.Clan{}}},
				},
				{
					ID:        id,
					Timestamp: models.Timestamp{CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond)},
					Servers:   []models.AccountServer{{Key: "key2", Clans: []models.Clan{}}},
				},
				{
					Timestamp: models.Timestamp{CreatedAt: time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Millisecond)},
					Servers:   []models.AccountServer{{Key: "key3", Clans: []models.Clan{}}},
				},
			}

			for _, account := range docs {
				coll.C.Insert(account)
			}

			got, err := accounts.GetByServerKey(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Accounts.GetByDiscordGuild() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.want, got, "Account is not what we expected")
		})
	}
}

func TestAccounts_UpsertBase(t *testing.T) {
	t.Parallel()

	var baseAccount = models.Account{
		Timestamp:   models.Timestamp{CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond)},
		BaseAccount: models.BaseAccount{GuildSnowflake: "yarp1"},
	}

	tests := []struct {
		name      string
		account   models.Account
		wantCount int
		wantErr   bool
	}{
		{
			name:      "insert",
			account:   models.Account{BaseAccount: models.BaseAccount{GuildSnowflake: "yuss"}},
			wantCount: 2,
		},
		{
			name:      "upsert",
			account:   models.Account{BaseAccount: models.BaseAccount{GuildSnowflake: "yarp1"}},
			wantCount: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accounts, coll := NewAccounts(t)
			defer coll.Close()

			err := coll.C.Insert(baseAccount)
			if err != nil {
				t.Fatal(err)
			}

			if err = accounts.UpsertBase(tt.account.BaseAccount); (err != nil) != tt.wantErr {
				t.Fatalf("Accounts.UpsertBase() error = %v, wantErr %v", err, tt.wantErr)
			}

			count, err := coll.C.Count()
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tt.wantCount, count)
		})
	}
}

func TestAccounts_Remove(t *testing.T) {
	t.Parallel()

	type args struct {
		key string
	}
	tests := []struct {
		name      string
		args      args
		wantCount int
		wantErr   bool
	}{
		{
			name:      "not found",
			args:      args{"nonexistant"},
			wantErr:   true,
			wantCount: 0,
		},
		{
			name:      "upsert",
			args:      args{"snowflake2"},
			wantCount: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accounts, coll := NewAccounts(t)
			defer coll.Close()

			make := []models.Account{
				{
					Timestamp:   models.Timestamp{CreatedAt: iclock().Now().UTC().Truncate(time.Millisecond)},
					BaseAccount: models.BaseAccount{GuildSnowflake: "snowflake1"},
				},
				{
					Timestamp:   models.Timestamp{CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond)},
					BaseAccount: models.BaseAccount{GuildSnowflake: "snowflake2"},
				},
				{
					Timestamp:   models.Timestamp{CreatedAt: iclock().Now().UTC().Add(-2 * time.Hour).Truncate(time.Millisecond)},
					BaseAccount: models.BaseAccount{GuildSnowflake: "snowflake3"},
				},
			}

			for _, account := range make {
				err := coll.C.Insert(account)
				if err != nil {
					t.Fatal(err)
				}
			}
			if err := accounts.Remove(tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("Accounts.Remove() error = %v, wantErr %v", err, tt.wantErr)
			}

			count, err := coll.C.Find(bson.M{"disabled": true}).Count()
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tt.wantCount, count, "Disabled count is wrong")

			count, err = coll.C.Count()
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, 3, count, "All count is wrong")
		})
	}
}

func TestAccounts_RemoveNotInDiscordGuildList(t *testing.T) {
	pbclock.Mock()
	t.Parallel()

	accounts, coll := NewAccounts(t)
	defer coll.Close()

	docs := []models.Account{
		{
			ID:          bson.NewObjectId(),
			Timestamp:   models.Timestamp{CreatedAt: iclock().Now().UTC().Truncate(time.Millisecond)},
			BaseAccount: models.BaseAccount{GuildSnowflake: "snowflake1"},
			Disabled:    true,
		},
		{
			ID:          bson.NewObjectId(),
			Timestamp:   models.Timestamp{CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond)},
			BaseAccount: models.BaseAccount{GuildSnowflake: "snowflake2"},
			Disabled:    false,
		},
	}

	for _, doc := range docs {
		err := coll.C.Insert(doc)
		if err != nil {
			t.Fatal(err)
		}
	}

	wantDocs := []models.Account{
		{
			ID:          docs[0].ID,
			Timestamp:   models.Timestamp{CreatedAt: iclock().Now().UTC().Truncate(time.Millisecond)},
			BaseAccount: models.BaseAccount{GuildSnowflake: "snowflake1"},
			Disabled:    true,
		},
		{
			ID: docs[1].ID,
			Timestamp: models.Timestamp{
				CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond),
				UpdatedAt: iclock().Now().UTC().Truncate(time.Millisecond),
			},
			BaseAccount: models.BaseAccount{GuildSnowflake: "snowflake2"},
			Disabled:    false,
		},
		{
			Timestamp: models.Timestamp{
				CreatedAt: iclock().Now().UTC().Truncate(time.Millisecond),
				UpdatedAt: iclock().Now().UTC().Truncate(time.Millisecond),
			},
			BaseAccount: models.BaseAccount{GuildSnowflake: "snowflake3"},
			Disabled:    false,
		},
	}

	args := []models.BaseAccount{
		wantDocs[1].BaseAccount,
		wantDocs[2].BaseAccount,
	}

	err := accounts.RemoveNotInDiscordGuildList(args)
	if err != nil {
		t.Errorf("Accounts.RemoveNotInDiscordGuildList() error %v", err)
		return
	}

	count, err := coll.C.Count()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 3, count, "Count is wrong")

	err = coll.C.Find(bson.M{}).Sort(accountsKeyField).All(&docs)
	if err != nil {
		t.Fatal(err)
	}

	// Since we don't know the inserted ID, we'll set it ourselves.
	wantDocs[2].ID = docs[2].ID

	assert.Equal(t, wantDocs, docs, "Docs are wrong: %v", docs)
}

func TestAccounts_AddClan(t *testing.T) {
	t.Parallel()

	type args struct {
		key  string
		clan models.Clan
	}
	tests := []struct {
		name    string
		args    args
		want    models.Account
		wantErr bool
	}{
		{
			name: "result",
			want: models.Account{
				Timestamp: models.Timestamp{CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond)},
				Servers: []models.AccountServer{{Key: "key2", Clans: []models.Clan{
					{Tag: "bloops"},
					{Tag: "bloops2"},
				}}},
			},
			args: args{key: "key2", clan: models.Clan{Tag: "bloops2"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accounts, coll := NewAccounts(t)
			defer coll.Close()

			id := bson.NewObjectId()
			tt.want.ID = id

			coll.C.Insert(models.Account{
				ID:        id,
				Timestamp: models.Timestamp{CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond)},
				Servers:   []models.AccountServer{{Key: "key2", Clans: []models.Clan{{Tag: "bloops"}}}},
			})

			if err := accounts.AddClan(tt.args.key, tt.args.clan); (err != nil) != tt.wantErr {
				t.Errorf("Accounts.AddClan() error = %v, wantErr %v", err, tt.wantErr)
			}

			var account models.Account
			coll.C.Find(bson.M{}).One(&account)
			assert.Equal(t, tt.want, account)
		})
	}
}

func TestAccounts_RemoveClan(t *testing.T) {
	t.Parallel()

	id := bson.NewObjectId()

	type args struct {
		key     string
		clanTag string
	}
	tests := []struct {
		name    string
		want    models.Account
		args    args
		wantErr bool
	}{
		{
			name: "result",
			want: models.Account{
				ID:        id,
				Timestamp: models.Timestamp{CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond)},
				Servers:   []models.AccountServer{{Key: "key2", Clans: []models.Clan{{Tag: "bloops2"}}}},
			},
			args: args{key: "key2", clanTag: "bloops"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accounts, coll := NewAccounts(t)
			defer coll.Close()

			coll.C.Insert(models.Account{
				ID:        id,
				Timestamp: models.Timestamp{CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond)},
				Servers:   []models.AccountServer{{Key: "key2", Clans: []models.Clan{{Tag: "bloops"}, {Tag: "bloops2"}}}},
			})

			if err := accounts.RemoveClan(tt.args.key, tt.args.clanTag); (err != nil) != tt.wantErr {
				t.Errorf("Accounts.RemoveClan() error = %v, wantErr %v", err, tt.wantErr)
			}

			var account models.Account
			coll.C.Find(bson.M{}).One(&account)
			assert.Equal(t, tt.want, account)
		})
	}
}

func TestAccounts_SetClans(t *testing.T) {
	t.Parallel()

	id := bson.NewObjectId()

	type args struct {
		key   string
		clans []models.Clan
	}
	tests := []struct {
		name    string
		args    args
		want    models.Account
		wantErr bool
	}{
		{
			name: "modified",
			want: models.Account{
				ID:        id,
				Timestamp: models.Timestamp{CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond)},
				Servers:   []models.AccountServer{{Key: "key2", Clans: []models.Clan{{Tag: "foo"}}}},
			},
			args: args{key: "key2", clans: []models.Clan{{Tag: "foo"}}},
		},
		{
			name: "existing",
			want: models.Account{
				ID:        id,
				Timestamp: models.Timestamp{CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond)},
				Servers:   []models.AccountServer{{Key: "key2", Clans: []models.Clan{{Tag: "existing"}}}},
			},
			args: args{key: "key2", clans: []models.Clan{{Tag: "existing"}}},
		},
		{
			name:    "not found",
			args:    args{key: "key", clans: []models.Clan{}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accounts, coll := NewAccounts(t)
			defer coll.Close()

			coll.C.Insert(models.Account{
				ID:        id,
				Timestamp: models.Timestamp{CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond)},
				Servers: []models.AccountServer{{
					Key:   "key2",
					Clans: []models.Clan{{Tag: "existing"}},
				}},
			})

			if err := accounts.SetClans(tt.args.key, tt.args.clans); (err != nil) != tt.wantErr {
				t.Errorf("Accounts.SetClans() error = %v, wantErr %v", err, tt.wantErr)
			}

			var account models.Account
			coll.C.Find(bson.M{serverKeyField: tt.args.key}).One(&account)
			assert.Equal(t, tt.want, account)
		})
	}
}

func TestAccounts_AddServer(t *testing.T) {
	type args struct {
		snowflake string
		server    models.AccountServer
	}
	tests := []struct {
		name    string
		args    args
		want    models.Account
		wantErr bool
	}{
		{
			name: "add",
			args: args{server: models.AccountServer{Key: "key"}, snowflake: "snowflake"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accounts, coll := NewAccounts(t)
			defer coll.Close()

			coll.C.Insert(baseAccount)

			if err := accounts.AddServer(tt.args.snowflake, tt.args.server); (err != nil) != tt.wantErr {
				t.Errorf("Accounts.AddServer() error = %v, wantErr %v", err, tt.wantErr)
			}
			var account models.Account
			coll.C.Find(bson.M{serverKeyField: tt.args.snowflake}).One(&account)
			assert.Equal(t, tt.want, account)
		})
	}
}

func TestAccounts_RemoveServer(t *testing.T) {
	type args struct {
		snowflake string
		serverKey string
	}
	tests := []struct {
		name    string
		s       Accounts
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.RemoveServer(tt.args.snowflake, tt.args.serverKey); (err != nil) != tt.wantErr {
				t.Errorf("Accounts.RemoveServer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAccounts_UpdateServer(t *testing.T) {
	type args struct {
		snowflake string
		oldKey    string
		server    models.AccountServer
	}
	tests := []struct {
		name    string
		s       Accounts
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.UpdateServer(tt.args.snowflake, tt.args.oldKey, tt.args.server); (err != nil) != tt.wantErr {
				t.Errorf("Accounts.UpdateServer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
