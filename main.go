package main

import (
	_ "embed"
	"fmt"
	"github.com/slack-go/slack"
	"os"
	"strings"
)

var token = os.Getenv("SLACK_TOKEN")

const (
	UserRateLimit   = 50
	PaginationLimit = 200
)

var (
	//go:embed resources/users.txt
	users string
	//go:embed resources/channels.txt
	channels string
	Users    = Strings(strings.Split(strings.TrimSpace(users), "\n"))
	Channels = Strings(strings.Split(strings.TrimSpace(channels), "\n"))
	api      = slack.New(token, slack.OptionDebug(true))
)

func main() {

	var cursor string
	var channelIds []string
	for conversations, nextCursor, err := getConversations(api, cursor); nextCursor != ""; conversations, nextCursor, err = getConversations(api, cursor) {

		if err != nil {
			fmt.Printf("%s\n", err)
			continue
		}

		for _, conversation := range conversations {
			if _, exists := Channels.Find(func(str string) bool {
				return str == conversation.Name
			}); exists {
				channelIds = append(channelIds, conversation.ID)
			}
		}
	}

	var userIds []string
	if len(Users) < UserRateLimit {
		userIds = getUsersByEmails(api)
	} else {
		userIds = getUsersBulk(api)
	}

	for _, channelId := range channelIds {
		_, err := api.InviteUsersToConversation(channelId, strings.Join(userIds, ","))

		if err != nil {
			fmt.Printf("%s\n", err)
		}
	}
}

func getUsersBulk(api *slack.Client) []string {
	realUsers, _ := api.GetUsers()

	userIds := make([]string, 0)
	for _, user := range realUsers {
		if _, exists := Users.Find(func(str string) bool {
			return str == user.Profile.Email
		}); exists {
			userIds = append(userIds, user.ID)
		}
	}
	return userIds
}

func getUsersByEmails(api *slack.Client) []string {
	userIds := make([]string, 0)
	for _, userEmail := range Users {
		user, err := api.GetUserByEmail(userEmail)

		if err != nil {
			fmt.Printf("%s\n", err)
			continue
		}

		userIds = append(userIds, user.ID)
	}
	return userIds
}

func getConversations(api *slack.Client, cursor string) ([]slack.Channel, string, error) {
	return api.GetConversations(&slack.GetConversationsParameters{
		Cursor:          cursor,
		ExcludeArchived: true,
		Limit:           PaginationLimit,
		Types:           []string{"public_channel", "private_channel"},
	})
}

type Strings []string

func (strings Strings) Find(test func(str string) bool) (string, bool) {
	for _, str := range strings {
		if test(str) {
			return str, true
		}
	}
	return "", false
}
