package helpers

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// MessageReactionsAll retrieve all Reactions from Message taking into account pagination
func MessageReactionsAll(session *discordgo.Session, channelID, messageID, emojiID string) (st []*discordgo.User, err error) {
	var body []byte
	var listUsers []*discordgo.User
	emojiID = strings.Replace(emojiID, "#", "%23", -1)
	uri := discordgo.EndpointMessageReactions(channelID, messageID, emojiID)

	v := url.Values{}

	v.Set("limit", strconv.Itoa(100))

	for {
		tempURI := uri
		if len(v) > 0 {
			tempURI += "?" + v.Encode()
		}

		body, err = session.RequestWithBucketID("GET", tempURI, nil, discordgo.EndpointMessageReaction(channelID, "", "", ""))
		if err != nil {
			return
		}
		err = unmarshal(body, &listUsers)

		if len(listUsers) == 0 {
			break
		}

		for k := range listUsers {
			ptr := *listUsers[k]
			st = append(st, &ptr)
		}

		v.Set("after", listUsers[len(listUsers)-1].ID)
	}

	return
}

func unmarshal(data []byte, v interface{}) error {
	err := json.Unmarshal(data, v)
	if err != nil {
		return discordgo.ErrJSONUnmarshal
	}

	return nil
}
