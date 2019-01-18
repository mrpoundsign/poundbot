package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gorilla/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"bitbucket.org/mrpoundsign/poundbot/storage/mocks"
	ptime "bitbucket.org/mrpoundsign/poundbot/time"
	"bitbucket.org/mrpoundsign/poundbot/types"
)

func TestChat_Handle(t *testing.T) {
	ptime.Mock()
	t.Parallel()

	tests := []struct {
		name   string
		method string
		s      *chat
		body   string
		rBody  string
		status int
		in     *types.ChatMessage
		out    *types.ChatMessage
		log    string
	}{
		{
			name:   "chat GET",
			method: http.MethodGet,
			s:      &chat{},
			status: http.StatusOK,
			out: &types.ChatMessage{
				SteamInfo:   types.SteamInfo{SteamID: 1234},
				ClanTag:     "FoO",
				DisplayName: "player",
				Message:     "hello there!",
				Source:      "discord",
				Timestamp:   types.Timestamp{CreatedAt: ptime.Clock().Now().UTC()},
			},
			body: "{\"SteamID\":1234,\"ClanTag\":\"FoO\",\"DisplayName\":\"player\",\"Message\":\"hello there!\",\"Source\":\"discord\",\"CreatedAt\":\"1970-01-01T00:00:00Z\"}",
		},
		{
			name:   "chat POST",
			method: http.MethodPost,
			s:      &chat{},
			status: http.StatusOK,
			rBody: `
			{
				"SteamId":1234,
				"ClanTag":"FoO",
				"DisplayName":"player",
				"Message":"hello there!"
			}
			`,
			in: &types.ChatMessage{
				SteamInfo:   types.SteamInfo{SteamID: 1234},
				ClanTag:     "FoO",
				DisplayName: "player",
				Message:     "hello there!",
				Source:      "rust",
				ChannelID:   "1234",
				Timestamp:   types.Timestamp{CreatedAt: ptime.Clock().Now().UTC()},
			},
		},
		{
			name:   "chat POST bad json",
			method: http.MethodPost,
			s:      &chat{},
			status: http.StatusOK,
			rBody:  "not JSON",
			log:    "[C] [request-1] Invalid JSON: invalid character 'o' in literal null (expecting 'u')\n",
		},
	}

	for _, tt := range tests {
		var logBuffer bytes.Buffer
		tt.s.logger.SetOutput(&logBuffer)
		tt.s.logger.SetPrefix("[C] ")

		t.Run(tt.name, func(t *testing.T) {
			var in *types.ChatMessage

			tt.s.in = make(chan types.ChatMessage)

			cs := mocks.ChatsStore{}
			tt.s.cs = &cs

			if tt.out != nil {
				cs.On("GetNext", "bloop", mock.AnythingOfType("*types.ChatMessage")).
					Return(func(serverKey string, m *types.ChatMessage) error {
						tmp := *tt.out
						*m = tmp
						return nil
					})
			}

			req, err := http.NewRequest(tt.method, "/chat", strings.NewReader(tt.rBody))
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()

			var wg sync.WaitGroup

			// Collect any incoming messages
			if tt.in != nil {
				wg.Add(1)
				var thing types.ChatMessage
				go func() {
					defer wg.Done()
					select {
					case thing = <-tt.s.in:
						in = &thing
						break
					}
				}()
			}

			context.Set(req, "requestUUID", "request-1")
			context.Set(req, "serverKey", "bloop")
			context.Set(req, "account", types.Account{Servers: []types.Server{
				{ChatChanID: "1234", Key: "bloop"},
			}})
			handler := http.HandlerFunc(tt.s.Handle)
			handler.ServeHTTP(rr, req)

			wg.Wait()

			assert.Equal(t, tt.body, rr.Body.String(), "handler returned bad body")
			assert.Equal(t, tt.status, rr.Code, "handler returned wrong status code")
			assert.Equal(t, tt.in, in, "handler got wrong in message")
			assert.Equal(t, tt.log, logBuffer.String(), "log was incorrect")
		})
	}
}
