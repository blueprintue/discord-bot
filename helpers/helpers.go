// Package helpers contains useful functions.
package helpers

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const limitMessageReactions = "100"

// MessageReactionsAll retrieve all Reactions from Message taking into account pagination.
func MessageReactionsAll(session *discordgo.Session, channelID, messageID, emojiID string) ([]*discordgo.User, error) {
	var listUsers []*discordgo.User

	emojiID = strings.ReplaceAll(emojiID, "#", "%23")
	uri := discordgo.EndpointMessageReactions(channelID, messageID, emojiID)

	urlValues := url.Values{}

	urlValues.Set("limit", limitMessageReactions)

	for {
		tempURI := uri
		if len(urlValues) > 0 {
			tempURI += "?" + urlValues.Encode()
		}

		body, err := session.RequestWithBucketID("GET", tempURI, nil, discordgo.EndpointMessageReaction(channelID, "", "", ""))
		if err != nil {
			return listUsers, fmt.Errorf("%w", err)
		}

		var listUsersFromAPI []*discordgo.User

		err = unmarshal(body, &listUsersFromAPI)
		if err != nil {
			return listUsers, err
		}

		if len(listUsersFromAPI) == 0 {
			break
		}

		for k := range listUsersFromAPI {
			ptr := *listUsersFromAPI[k]
			listUsers = append(listUsers, &ptr)
		}

		urlValues.Set("after", listUsersFromAPI[len(listUsersFromAPI)-1].ID)
	}

	return listUsers, nil
}

func unmarshal(data []byte, v any) error {
	err := json.Unmarshal(data, v)
	if err != nil {
		return discordgo.ErrJSONUnmarshal
	}

	return nil
}
