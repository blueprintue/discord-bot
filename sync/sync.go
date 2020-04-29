package sync

// todo : faire la synchro avec l'intégration de twitch subscriber
// 		  -> ajouter les gens au bon role

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
)

type Manager struct {
	channelID string
	session *discordgo.Session
}

func NewSyncManager(
	channelID string,
	session *discordgo.Session,
) *Manager {
	return &Manager{
		channelID: channelID,
		session: session,
	}
}

func (s *Manager) Run() {
	fmt.Println("Welcome -> Run")

	s.SyncIntegration()
	s.session.AddHandler(s.onMessageReactionAdd)
}

func (s *Manager) SyncIntegration() {
	// liste des user avec un role -> s.session.user
}

func (s *Manager) onMessageReactionAdd(session *discordgo.Session, reaction *discordgo.Role) {

}