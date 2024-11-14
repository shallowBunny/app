package telegram

import (
	"fmt"
	"html"
	"regexp"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
	"github.com/shallowBunny/app/be/internal/bot"
)

type Telegram struct {
	bot *bot.Bot
	api *tgbotapi.BotAPI
}

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

func (t *Telegram) SendPhoto(chatID int64, photoFilePath string, caption string) error {
	// Create a new photo message with a local file
	photoMsg := tgbotapi.NewPhoto(chatID, tgbotapi.FilePath(photoFilePath))

	// Add a caption to the photo if needed
	if caption != "" {
		photoMsg.Caption = caption
		photoMsg.ParseMode = "HTML" // Optional: use HTML formatting if needed
	}

	// Send the photo message
	_, err := t.api.Send(photoMsg)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("Error sending photo: %v", err))
	}
	return err
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

func (t Telegram) Listen(quit <-chan struct{}) {
	// Start the goroutine to send messages
	go t.SendMessages()

	for {
		// Telegram polling configuration
		updateConfig := tgbotapi.NewUpdate(0)
		updateConfig.Timeout = 30

		// Get the updates channel from Telegram
		updates := t.api.GetUpdatesChan(updateConfig)
		restartListener := false
		for !restartListener {

			select {
			case update, ok := <-updates:
				if !ok {
					log.Warn().Msg("Updates channel closed, restarting down Telegram listener.")
					restartListener = true
					break
				}

				// Process the update if it's a valid message
				if update.Message == nil {
					log.Debug().Msg("update. update.Message == nil ...")
					continue
				}

				if t.bot.GetConfig().TelegramDeleteLeftTheGroupMessages {
					if update.Message.LeftChatMember != nil {
						log.Debug().Msg("deleting message")
						deleteMsg := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID)
						if _, err := t.api.Send(deleteMsg); err != nil {
							log.Error().Msg(fmt.Sprintf("Failed to delete message: %v", err))
						}
						continue
					}
				}

				if update.Message.Chat.IsGroup() || update.Message.Chat.IsChannel() {
					log.Debug().Msg("update.Message.Chat.Type = group")
					userString := getUsername(update)
					group := update.Message.Chat.Title
					t.bot.GroupChange(update.Message.Chat.ID, userString, group)
					continue
				}

				var msg tgbotapi.MessageConfig
				if update.Message.Chat.ID > 0 { // Skip joins in channels
					messages := t.bot.ProcessCommand(update.Message.Chat.ID, update.Message.Text, getUsername(update))
					for i, answer := range messages {

						if answer.ImagePath != "" {
							if err := t.SendPhoto(answer.UserID, answer.ImagePath, answer.Text); err != nil {
								log.Error().Msg("Failed to send image")
							}
							continue
						}

						if len(answer.Text) > 4096 {
							log.Error().Msg("Had to trim inside!!?")
							answer.Text = bot.Trim(answer.Text)
						}
						msg = messageToMessageConfig(answer)
						t.bot.Log(update.Message.Chat.ID, update.Message.Text, getUsername(update))

						if i == 0 {
							msg.ReplyToMessageID = update.Message.MessageID
						}

						if _, err := t.api.Send(msg); err != nil {
							log.Error().Msg(fmt.Sprintf("%v %v", err.Error(), msg))
						}
					}
				}

			case <-quit: // Listen for quit signal
				log.Info().Msg("Shutting down Telegram listener gracefully")
				return
			}
		}
		log.Warn().Msg("Restarting Telegram listener")
	}
}
