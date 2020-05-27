package models

import (
	"fmt"
	"image/color"

	"golang.org/x/image/colornames"
)

type GameMessageType int

const (
	GameMessageTypePlain GameMessageType = iota
	GameMessageTypeEmbed
)

type GameMessageEmbedStyle struct {
	Color string
}

// ColorInt converts the Color string to an integer
func (es GameMessageEmbedStyle) ColorInt() int {
	c, err := parseHexColor(es.Color)
	if err != nil {
		nColor, ok := colornames.Map[es.Color]
		if !ok {
			nColor = colornames.Map["blue"]
		}
		c = nColor
	}
	return (int(c.R) << 16) | (int(c.G) << 8) | int(c.B)
}

type GameMessagePart struct {
	Content string
	Escape  bool
}

// GameMessage is a message from the game server intended for discord
type GameMessage struct {
	Type          GameMessageType
	EmbedStyle    GameMessageEmbedStyle
	ChannelName   string
	MessageParts  []GameMessagePart
	Snowflake     string       `json:"-"`
	ErrorResponse chan<- error `json:"-"`
}

func parseHexColor(s string) (c color.RGBA, err error) {
	c.A = 0xff
	switch len(s) {
	case 7:
		_, err = fmt.Sscanf(s, "#%02x%02x%02x", &c.R, &c.G, &c.B)
	case 4:
		_, err = fmt.Sscanf(s, "#%1x%1x%1x", &c.R, &c.G, &c.B)
		// Double the hex digits:
		c.R *= 17
		c.G *= 17
		c.B *= 17
	default:
		err = fmt.Errorf("invalid length, must be 7 or 4")
	}
	return
}
