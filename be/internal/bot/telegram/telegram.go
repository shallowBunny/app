package telegram

import (
	"fmt"
	"html"
	"os"
	"regexp"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
	"github.com/shallowBunny/app/be/internal/bot"
	"github.com/shallowBunny/app/be/internal/bot/lineUp/inputs"
)

type Telegram struct {
	bot *bot.Bot
	api *tgbotapi.BotAPI
}

var (
	doOnce = true //false
)

func New(apiToken string, bot *bot.Bot) *Telegram {

	api, err := tgbotapi.NewBotAPI(apiToken)
	if err != nil {
		log.Error().Msg(err.Error())
		panic(err)
	}
	log.Trace().Msg("using api token <" + apiToken + ">")

	return &Telegram{
		bot: bot,
		api: api,
	}
}

func getUsername(update tgbotapi.Update) string {
	userString := ""
	if update.Message.From.UserName != "" {
		userString += "@" + update.Message.From.UserName
	} else {
		if update.Message.From.FirstName != "" {
			userString += update.Message.From.FirstName
		}
		if update.Message.From.LastName != "" {
			if userString != "" {
				userString += "."
			}
			userString += update.Message.From.LastName
		}
	}
	return userString
}

func escapeHTMLSpecialChars(text string) string {
	// Regular expression to match HTML tags
	re := regexp.MustCompile(`<[^>]*>`)

	// Find all matches of HTML tags
	matches := re.FindAllStringIndex(text, -1)

	// Escape the text outside the HTML tags
	result := ""
	lastIndex := 0
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		// Append text before the tag
		result += html.EscapeString(text[lastIndex:match[0]])
		// Append the tag itself without escaping
		result += text[match[0]:match[1]]
		// Update the last index
		lastIndex = match[1]
	}

	// Append any remaining text after the last tag
	result += html.EscapeString(text[lastIndex:])

	return result
}
func messageToMessageConfig(msgFromBot bot.Message) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(msgFromBot.UserID, msgFromBot.Text)

	if msgFromBot.Html {

		msg.ParseMode = "HTML"
		msg.Text = escapeHTMLSpecialChars(msg.Text)
		log.Trace().Msg(msg.Text)

	}
	msg.DisableWebPagePreview = true
	buttons := []tgbotapi.KeyboardButton{}

	if msgFromBot.Buttons != nil {
		for _, v := range msgFromBot.Buttons {
			buttons = append(buttons, tgbotapi.NewKeyboardButton(v))
		}
		if len(buttons) != 0 {
			var numericKeyboard = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(buttons...))
			msg.ReplyMarkup = numericKeyboard
		} else {
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		}
	} else {
		log.Debug().Msg("msgFromBot.Buttons == nil ")
	}

	return msg
}

func (t Telegram) SendMessages() {

	for {
		msgFromBot := <-t.bot.GetMessageChannel()

		msg := messageToMessageConfig(msgFromBot)

		if _, err := t.api.Send(msg); err != nil {
			log.Error().Msg(err.Error())
			if err.Error() == "Forbidden: bot was blocked by the user" {
				err = t.bot.DeleteUser(msg.ChatID)
				if err != nil {
					log.Error().Msg(err.Error())
				} else {
					log.Info().Msg("deleted user")
				}
			}
		}
	}
}

func (t Telegram) Listen() {

	go t.SendMessages()

	//bot := New(apiToken, redisclient)

	// Create a new UpdateConfig struct with an offset of 0. Offsets are used
	// to make sure Telegram knows we've handled previous values and we don't
	// need them repeated.
	updateConfig := tgbotapi.NewUpdate(0)

	// Tell Telegram we should wait up to 30 seconds on each request for an
	// update. This way we can get information just as quickly as making many
	// frequent requests without having to send nearly as many.
	updateConfig.Timeout = 30

	// Start polling Telegram for updates.
	updates := t.api.GetUpdatesChan(updateConfig)

	// Let's go through each update that we're getting from Telegram.
	for update := range updates {

		/*
			b, err := json.MarshalIndent(update, "", "   ")
			if err != nil {
				log.Error().Msg(err.Error())
			} else {
				fmt.Println(string(b))
			}
		*/

		// Telegram can send many types of updates depending on what your Bot
		// is up to. We only want to look at messages for now, so we can
		// discard any other updates.

		if update.MyChatMember != nil {
			log.Debug().Msg("update.MyChatMember...")
			continue
		}

		if update.Message == nil {
			log.Debug().Msg("update. update.Message == nil ...")
			continue
		}
		if update.Message.Chat.IsGroup() || update.Message.Chat.IsChannel() {
			log.Debug().Msg("update.Message.Chat.Type = group")
			userString := getUsername(update)
			group := update.Message.Chat.Title
			t.bot.GroupChange(update.Message.Chat.ID, userString, group)
			continue
		}

		// Now that we know we've gotten a new message, we can construct a
		// reply! We'll take the Chat ID and Text from the incoming message
		// and use it to create a new message.

		//PrintForTime(sets, teleportTimeString)
		//update.Message.Text
		var msg tgbotapi.MessageConfig

		if !doOnce {
			doOnce = true

			hostname, err := os.Hostname()
			if err != nil {
				panic(nil)
			}
			if hostname == "Mac-mini.local" {
				currentTime := time.Now().Add(24 * time.Hour)
				inputCommands := []string{
					inputs.InputCommand, "🔨 Hammahalle", "Fri", "23:00", "Fggg", "210", inputs.ValidateCommand,
					inputs.MergeCommand, inputs.MergeSubmitCommand,
					inputs.InputCommand, "🔨 Hammahalle", currentTime.Format("Mon"), "2:30", "DJ FART", "90", inputs.ValidateCommand,
					inputs.MergeCommand, inputs.MergeSubmitCommand}

				for _, tc := range inputCommands {
					answer := t.bot.ProcessCommand(update.Message.Chat.ID, tc, "test")
					log.Debug().Msg(fmt.Sprintf("xx %v answer = %v", tc, answer))
				}
			}
		}

		if update.Message.Chat.ID > 0 { // to skip joins in channels
			messages := t.bot.ProcessCommand(update.Message.Chat.ID, update.Message.Text, getUsername(update))

			for i, answer := range messages {

				if len(answer.Text) > 4096 {
					log.Error().Msg("had to trim inside!!?")
					answer.Text = bot.Trim(answer.Text)
				}

				msg = messageToMessageConfig(answer)

				t.bot.Log(update.Message.Chat.ID, update.Message.Text, getUsername(update))

				// We'll also say that this message is a reply to the previous message.
				// For any other specifications than Chat ID or Text, you'll need to
				// set fields on the `MessageConfig`.

				if i == 0 {
					msg.ReplyToMessageID = update.Message.MessageID
				}

				// Okay, we're sending our message off! We don't care about the message
				// we just sent, so we'll discard it.
				if _, err := t.api.Send(msg); err != nil {
					// Note that panics are a bad way to handle errors. Telegram can
					// have service outages or network errors, you should retry sending
					// messages or more gracefully handle failures.
					log.Error().Msg(fmt.Sprintf("%v %v", err.Error(), msg))
				}
			}
		}
	}
}
