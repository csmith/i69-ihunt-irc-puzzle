package main

import (
	_ "embed"
	"fmt"
	"github.com/ergochat/irc-go/ircevent"
	"github.com/ergochat/irc-go/ircmsg"
	"log"
	"math/rand"
	"strings"
)

type Anna struct {
	ircevent.Connection
	channels map[string]string
	words    map[string]bool
	colours  []string
}

func must(err error) {
	if err != nil {
		log.Fatalf("Error invoking IRC command: %v", err)
	}
}

func main() {
	bot := &Anna{
		Connection: ircevent.Connection{
			Server:   "irc:6667",
			Nick:     "Anna",
			RealName: "Anna",
			UseSASL:  false,
			UseTLS:   false,
			Debug:    true,
		},
		channels: make(map[string]string),
		colours:  strings.Split(colours, "\n"),
		words:    loadWords(),
	}

	bot.AddConnectCallback(func(message ircmsg.Message) {
		must(bot.Send("OPER", "anna", "*Qc$ZRDB8WT2Kw"))
	})

	// RPL_YOUREOPER
	bot.AddCallback("381", func(message ircmsg.Message) {
		must(bot.Send("SAJOIN", bot.CurrentNick(), "#admin"))
		must(bot.Send("SAMODE", "#admin", "+Cnsti"))
	})

	bot.AddCallback("NOTICE", func(message ircmsg.Message) {
		if !strings.ContainsRune(message.Source, '.') {
			// Ignore notices that aren't from servers
			return
		}

		var nick string
		n, _ := fmt.Sscanf(message.Params[1], "\x0314-\x0fQUIT\x0314-\x03 %s exited the network", &nick)
		if n == 1 {
			nick = strings.TrimSuffix(nick, "\x0F")
			must(bot.Send("PRIVMSG", "#admin", fmt.Sprintf("User disconnected - %s - closed channel %s", nick, bot.channels[nick])))
			must(bot.Send("PART", bot.channels[nick]))
			delete(bot.channels, nick)
		}

		n, _ = fmt.Sscanf(message.Params[1], "\x0314-\x0fCONNECT\x0314-\x03 Client connected [%s", &nick)
		if n == 1 {
			nick = strings.TrimSuffix(nick, "]")
			log.Printf("User connected: %s", nick)
			channel := RandChannel()
			bot.channels[nick] = channel
			must(bot.Send("SAJOIN", bot.CurrentNick(), channel))
			must(bot.Send("TOPIC", channel, "Welcome! Feel free to chat but please be sure to obey all channel rules."))
			must(bot.Send("SAJOIN", nick, channel))
			must(bot.Send("PRIVMSG", "#admin", fmt.Sprintf("New user connected - %s - joined to channel %s", nick, channel)))
		}

		var newNick string
		n, _ = fmt.Sscanf(message.Params[1], "\x0314-\x0fNICK\x0314-\x03 %s changed nickname to %s", &nick, &newNick)
		if n == 2 {
			nick = strings.TrimSuffix(nick, "\x0F")
			log.Printf("User changed names: %s -> %s", nick, newNick)
			must(bot.Send("PRIVMSG", "#admin", fmt.Sprintf("User changed nicks - %s -> %s", nick, newNick)))
			bot.channels[newNick] = bot.channels[nick]
			delete(bot.channels, nick)
		}
	})

	bot.AddCallback("PRIVMSG", func(message ircmsg.Message) {
		body := strings.Join(message.Params[1:], " ")
		nick := strings.Split(message.Source, "!")[0]
		channel := message.Params[0]

		result := bot.checkMessage(body)
		log.Printf("User %s said '%s' in %s result: %s", nick, body, channel, result)
		if result != correctString {
			must(bot.Send("KICK", channel, nick, result))
		}
	})

	must(bot.Connection.Connect())

	bot.Loop()
}

//go:embed words.txt
var wordList string

//go:embed colours.txt
var colours string

func loadWords() map[string]bool {
	res := make(map[string]bool)
	words := strings.Split(wordList, "\n")
	for i := range words {
		res[words[i]] = true
	}
	return res
}

const correctString = "Correct"

func (a *Anna) checkMessage(text string) string {
	text = strings.ToLower(text)
	for i := range text {
		if text[i] < 'a' || text[i] > 'z' {
			return "Rule 1: A-Z only, no spaces, punctuation or accents."
		}
	}

	if strings.ContainsRune(text, 'e') {
		return "Rule 2: No use of the letter 'E' is permitted."
	}

	if strings.ContainsAny(text, "iou") {
		return "Rule 3: Vowels other than 'A' are not supported."
	}

	if len(text) < 5 {
		return "Rule 4: Words must be at least 5 letters long."
	}

	if len(text) > 10 {
		return "Rule 5: Ain't nobody got time to read words longer than 10 letters."
	}

	if !a.words[text] {
		return "Rule 6: Non-dictionary words are forbidden."
	}

	found := false
	for i := range a.colours {
		if strings.Contains(text, a.colours[i]) {
			found = true
			break
		}
	}
	if !found {
		return "Rule 7: Word must contain a colour from Joseph's Technicolour Dreamcoat"
	}

	if !strings.Contains(text, "black") {
		return "Rule 8: The colour from rule 7 must be black."
	}

	if !strings.HasPrefix(text, "black") {
		return "Rule 9: The colour from rule 7 must be at the start of the word."
	}

	if len(text) != 9 {
		return "Rule 10: Words must be exactly 9 letters long"
	}

	if text != "blackjack" {
		return "Rule 11: Words must name a card game where you attempt to get close to - but not above - 21."
	}

	return correctString
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandChannel() string {
	b := make([]byte, 32)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return "#" + string(b)
}
