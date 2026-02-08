package welcome

func (w *Manager) isUserBot(userID string) bool {
	return w.discordSession.State.User.ID == userID
}
